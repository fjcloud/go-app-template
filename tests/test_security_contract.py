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

class CardScoutContract(unittest.TestCase):
    def test_public_https_route_reaches_the_app_directly(self):
        spec = yaml.safe_load((ROOT/'deploy/base/deployment.yaml').read_text())['spec']['template']['spec']
        self.assertEqual(spec['serviceAccountName'], 'signal-desk')
        self.assertEqual(len(spec['containers']), 1)
        app = spec['containers'][0]
        self.assertIn({'name':'LISTEN_ADDR','value':':8080'}, app['env'])
        self.assertIn({'name':'http','containerPort':8080}, app['ports'])
        service = yaml.safe_load((ROOT/'deploy/base/service.yaml').read_text())
        self.assertEqual(service['spec']['ports'], [{'name':'http','port':8080,'targetPort':'http'}])
        route = yaml.safe_load((ROOT/'deploy/base/route.yaml').read_text())
        self.assertEqual(route['spec']['port']['targetPort'], 'http')
        self.assertEqual(route['spec']['tls'], {'termination':'edge','insecureEdgeTerminationPolicy':'Redirect'})
        self.assertEqual(route['metadata']['annotations']['haproxy.router.openshift.io/timeout'], '120s')

    def test_local_model_needs_no_runtime_api_key(self):
        spec = yaml.safe_load((ROOT/'deploy/base/deployment.yaml').read_text())['spec']['template']['spec']
        self.assertNotIn('runtime-model', {v['name'] for v in spec.get('volumes', [])})
        app = spec['containers'][0]
        self.assertEqual(app['envFrom'], [{'configMapRef': {'name': 'signal-desk-config'}}])
        self.assertNotIn('OPENAI_API_KEY', {e['name'] for e in app['env']})
        self.assertTrue(all(p['httpGet']['scheme']=='HTTP' and p['httpGet']['port']=='http' for n,p in app.items() if n.endswith('Probe')))

    def test_thin_app_needs_no_credentials_or_persistent_volume(self):
        deployment = yaml.safe_load((ROOT/'deploy/base/deployment.yaml').read_text())['spec']
        self.assertEqual(deployment['strategy']['type'], 'Recreate')
        spec = deployment['template']['spec']
        self.assertFalse(spec['automountServiceAccountToken'])
        self.assertEqual(spec.get('volumes', []), [])
        self.assertEqual(spec['containers'][0].get('volumeMounts', []), [])
        self.assertFalse(any(e['name'].endswith('KEY_FILE') for e in spec['containers'][0]['env']))

    def test_contract_generates_new_app_against_prepared_api(self):
        contract = (ROOT/'APP_CONTRACT.md').read_text()
        agents = (ROOT/'AGENTS.md').read_text()
        self.assertIn('from scratch', contract)
        self.assertIn('CARD_SCOUT_URL', contract)
        self.assertNotIn('cp -R', agents)
        self.assertNotIn('briefing-desk', agents)
