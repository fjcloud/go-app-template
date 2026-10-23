# Go App Template — ARO Live Transcript

Use the `aro-transcription` branch for the [ARO workshop](https://fjcloud.github.io/aro-ai-app-platform/modules/index.html). The `main` branch remains the ROSA Fortune Cookie template.

OpenCode runs in Dev Spaces and uses **DeepSeek-V4 Pro on Foundry** to write the Go app. **RedHatAI Whisper Large v3 Turbo INT8**, served through OpenShift AI and Red Hat AI Inference on one T4, transcribes the app’s audio. These are separate model paths.

The browser captures short WAV clips and sends them to the Go backend, which calls the internal Whisper service. Results appear after each clip; this is near-real-time transcription.

| Path | Purpose |
| --- | --- |
| `AGENTS.md` | Generation instructions, delivery steps and the speech API contract |
| `devfile.yaml` | Browser workspace and ephemeral Git client |
| `opencode.json` | Project model configuration; Lab 2.1 replaces the inherited pin with the platform’s Foundry configuration |
| `deploy/base/` | Go application Deployment, Service and HTTPS Route |
| `pipeline/base/` | Tekton image build |
| `gitops/base/` | Developer-owned Argo CD instance |

Follow the workshop to prepare the platform, authenticate `oc` inside the workspace and copy the platform model configuration. Then plan, generate, build and deploy with OpenCode. Verify the matching Git revision in Argo CD and test the microphone in the HTTPS app. The model key stays in workspace environment variables; speech audio goes to Whisper through the backend.

The Dockerfile ships only the Go binary, so embed all web assets. The [tested reference app](https://github.com/fjcloud/aro-ai-app-platform/tree/main/examples/live-transcript) provides the capture and proxy baseline. The exercise adds a Copy transcript button.

This template uses an external ephemeral Git service and fixed app namespaces. Use one participant per assigned cluster unless the facilitator adapts app names and permissions. Export code before the RHDP assignment expires.

The Deployment already points to the internal image built by Tekton and uses `imagePullPolicy: Always`. The pipeline runs `clone → validate-image → build`: a placeholder, wrong namespace or mismatched deployment image stops the build with an explicit error. For older personal checkouts, fix the image in Git before rebuilding; Argo CD reads the committed manifest.

Template regression checks:

```sh
python3 -m pip install -r tests/requirements.txt
python3 -m unittest discover -s tests
```
