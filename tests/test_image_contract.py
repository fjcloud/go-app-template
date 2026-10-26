"""Run the pipeline's real validation script against controlled CLI responses."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

import yaml


ROOT = Path(__file__).resolve().parents[1]
PIPELINE = yaml.safe_load((ROOT / 'pipeline/base/pipeline.yaml').read_text())
DEPLOYMENT = yaml.safe_load((ROOT / 'deploy/base/deployment.yaml').read_text())
TASKS = {task['name']: task for task in PIPELINE['spec']['tasks']}
SCRIPT = TASKS['validate-image']['taskSpec']['steps'][0]['script']
IMAGE = 'image-registry.openshift-image-registry.svc:5000/test-app-build/test-app:latest'
FAKE_OC = r'''#!/usr/bin/env python3
import json,os,sys
with open(os.environ['OC_CALLS'],'a') as f: f.write(json.dumps(sys.argv[1:])+'\n')
if sys.argv[1]=='kustomize':
    if os.environ.get('RENDER_FAIL'): sys.exit('Kustomize could not render the manifests')
    print('Rendered Kubernetes resources')
elif sys.argv[1]=='create' and '--dry-run=client' in sys.argv:
    print(os.environ['RENDERED_IMAGE'],end='')
else:
    sys.exit('Unexpected or mutating oc command')
'''


class ImageContract(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.directory = Path(self.temp.name)
        self.calls = self.directory/'oc-calls.jsonl'
        oc = self.directory/'oc'
        oc.write_text(FAKE_OC)
        oc.chmod(0o755)
        self.env = dict(os.environ, PATH=str(self.directory)+os.pathsep+os.environ['PATH'],
                        BUILD_NAMESPACE='test-app-build', BUILD_IMAGE=IMAGE,
                        RENDERED_IMAGE=IMAGE, OC_CALLS=str(self.calls))

    def run_guard(self):
        return subprocess.run(['sh'], input=SCRIPT.replace('$(workspaces.source.path)', str(self.directory)), text=True, capture_output=True,
                              env=self.env, cwd=self.directory, timeout=10)

    def test_template_and_default_build_image_are_identical(self):
        app = DEPLOYMENT['metadata']['name']
        expected = f'image-registry.openshift-image-registry.svc:5000/{app}-build/{app}:latest'
        container = DEPLOYMENT['spec']['template']['spec']['containers'][0]
        self.assertEqual(container['image'], expected)
        self.assertEqual(container['imagePullPolicy'], 'Always')
        self.assertEqual(next(p['default'] for p in PIPELINE['spec']['params'] if p['name']=='image'), expected)

    def test_build_requires_successful_validation(self):
        self.assertEqual(TASKS['validate-image']['runAfter'], ['clone'])
        self.assertEqual(TASKS['build']['runAfter'], ['validate-image'])
        self.assertNotIn('when', TASKS['validate-image'])
        self.assertNotIn('onError', TASKS['validate-image']['taskSpec']['steps'][0])

    def test_matching_images_pass_and_render_kustomize(self):
        result = self.run_guard()
        self.assertEqual(result.returncode, 0, result.stderr)
        calls = [json.loads(line) for line in self.calls.read_text().splitlines()]
        self.assertEqual(calls[0], ['kustomize', 'deploy/base'])
        self.assertIn('--dry-run=client', calls[1])
        self.assertIn('jsonpath={..image}', calls[1])

    def test_placeholder_build_parameter_stops_before_any_cli_call(self):
        self.env['BUILD_IMAGE'] = 'placeholder'
        result = self.run_guard()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('Invalid build image', result.stderr)
        self.assertFalse(self.calls.exists())

    def test_placeholder_in_rendered_deployment_stops(self):
        self.env['RENDERED_IMAGE'] = 'placeholder'
        result = self.run_guard()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('Deployment image mismatch: placeholder', result.stderr)
        self.assertIn('commit and push', result.stderr)

    def test_wrong_build_namespace_image_stops(self):
        self.env['BUILD_IMAGE'] = IMAGE.replace('test-app-build/', 'other-build/')
        result = self.run_guard()
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse(self.calls.exists())

    def test_other_rendered_image_stops(self):
        self.env['RENDERED_IMAGE'] = 'docker.io/library/test-app:latest'
        self.assertNotEqual(self.run_guard().returncode, 0)

    def test_missing_or_multiple_images_stop(self):
        for value in ['', IMAGE+' '+IMAGE]:
            with self.subTest(rendered=value):
                self.env['RENDERED_IMAGE'] = value
                self.assertNotEqual(self.run_guard().returncode, 0)

    def test_render_failure_stops_before_client_create(self):
        self.env['RENDER_FAIL'] = '1'
        self.assertNotEqual(self.run_guard().returncode, 0)
        calls = [json.loads(line) for line in self.calls.read_text().splitlines()]
        self.assertEqual(len(calls), 1)

    def test_wrong_namespace_stops_before_any_cli_call(self):
        self.env['BUILD_NAMESPACE'] = 'default'
        self.assertNotEqual(self.run_guard().returncode, 0)
        self.assertFalse(self.calls.exists())


if __name__ == '__main__':
    unittest.main()
