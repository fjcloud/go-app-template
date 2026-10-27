# Go App Template — ARO Cluster Voice

Use the `aro-transcription` branch for the [ARO workshop](https://fjcloud.github.io/aro-ai-app-platform/modules/index.html). The technical application name remains `live-transcript`; the `main` branch remains the ROSA Fortune Cookie template.

Speak to your OpenShift cluster: **Whisper on one T4 → editable question → DeepSeek through MaaS → read-only metrics tools → answer with evidence**. Text questions also work. The same connector exposes MCP tools for the OpenShift AI Playground, and OpenTelemetry exports model/tool traces to MLflow.

OpenCode uses a coding MaaS key to adapt the [reference implementation](https://github.com/fjcloud/aro-ai-app-platform/tree/main/examples/live-transcript). The deployed Go agent has a separate runtime key, quota and ServiceAccount. Audio stays on the cluster; questions and selected monitoring data go to Foundry. Model keys and service-account credentials stay out of browser code.

| Path | Purpose |
| --- | --- |
| `AGENTS.md` | Reference app, bounded tool/speech contracts and delivery instructions |
| `devfile.yaml`, `opencode.json` | Dev Spaces and OpenCode coding model |
| `deploy/base/` | Restricted Go app, OpenShift OAuth proxy, HTTPS Service and reencrypt Route |
| `pipeline/base/` | Tekton clone, app-image validation and build |
| `gitops/base/` | Developer-owned Argo CD instance |

Complete the workshop platform steps, including **Lab 1.6 runtime preparation**, before deploying. Copy all reference Go files, dependency files and embedded web assets; keep the prepared deployment scaffold. Ask a spoken question, review its text and submit it. Inspect actual PromQL, units, timestamps and missing-data errors in the evidence panel. The change exercise adds a **CPU over 1h** shortcut.

The app listens on loopback behind its OpenShift OAuth proxy. Participants need permission to read the application Service. The connector executes only server-owned metric expressions; it has no shell or mutation tools and never grants SCC privileges. The platform prepares read-only monitoring access and namespace-scoped MLflow integration.

The Deployment points to the exact app image built by Tekton with `imagePullPolicy: Always`. The image guard selects the application container by name so the prepared OAuth sidecar is validated separately by platform setup. Never replace the app image with a placeholder or bypass the guard.

This template uses an external ephemeral Git service and fixed namespaces. Use one participant per assigned cluster unless the facilitator isolates names and permissions. Export code before RHDP expiration. Runtime and coding MaaS keys expire after 24 hours; refresh them through their respective platform setup steps.

```sh
python3 -m pip install -r tests/requirements.txt
python3 -m unittest discover -s tests
```
