"""Keep delivery scaffolding within the workshop security boundaries."""
from pathlib import Path
import unittest
import yaml
ROOT = Path(__file__).resolve().parents[1]

class SecurityContract(unittest.TestCase):
    def test_application_requires_restricted_scc(self):
        d = yaml.safe_load((ROOT/'deploy/base/deployment.yaml').read_text())['spec']['template']
        self.assertEqual(d['metadata']['annotations']['openshift.io/required-scc'], 'restricted-v2')
        self.assertTrue(d['spec']['securityContext']['runAsNonRoot'])
        for c in d['spec']['containers']:
            self.assertFalse(c['securityContext']['allowPrivilegeEscalation'])
            self.assertEqual(c['securityContext']['capabilities']['drop'], ['ALL'])
            self.assertFalse(c['securityContext'].get('privileged', False))

    def test_argocd_cannot_sync_cluster_resources(self):
        project = yaml.safe_load((ROOT/'gitops/base/appproject.yaml').read_text())
        self.assertEqual(project['spec']['clusterResourceWhitelist'], [])

    def test_guard_reads_real_mounted_workspace(self):
        pipeline = yaml.safe_load((ROOT/'pipeline/base/pipeline.yaml').read_text())
        guard = next(t for t in pipeline['spec']['tasks'] if t['name']=='validate-image')['taskSpec']
        self.assertEqual(guard['workspaces'][0]['mountPath'], '/source')
        step = guard['steps'][0]
        self.assertEqual(step['workingDir'], '/')
        self.assertIn('cd "$(workspaces.source.path)"', step['script'])

    def test_runbook_uses_user_namespace_without_scc_grants(self):
        agents = (ROOT/'AGENTS.md').read_text()
        self.assertIn('hostUsers: false', agents)
        self.assertNotIn('add-scc-to-user', agents)
        self.assertNotIn('privileged: true', agents)
        self.assertNotIn('oc create namespace', agents)
