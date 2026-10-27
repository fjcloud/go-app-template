# AGENTS.md

OpenShift Dev Space. `oc` is authenticated. `$GIT_SERVER` and `$APP_NAME` are set.

No Ansible. No `scripts/`. Do not rewrite `deploy/`, `pipeline/`, `gitops/`, `devfile.yaml`, or `opencode.json` (only change `image:` in `deploy/base/deployment.yaml` when shipping).

## Generate

`main.go` — HTTP on `$LISTEN_ADDR` (default `:8080`, `127.0.0.1:8080` behind the prepared OAuth proxy), `GET /`, `GET /healthz` → `{"status":"ok"}`
`go.mod` — `module <app-name>` / `go 1.24.0` (MCP SDK requires Go 1.24)

Dockerfile (exact; go-toolset cannot write to `/app`):

```dockerfile
FROM registry.access.redhat.com/ubi9/go-toolset:latest AS builder
WORKDIR /tmp/build
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN CGO_ENABLED=0 go build -buildvcs=false -o /tmp/app .

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
COPY --from=builder /tmp/app /usr/local/bin/app
USER 1001
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/app"]
```

Then: `CGO_ENABLED=0 go build -buildvcs=false -o /dev/null .`

## Deploy (git + oc)

When asked to ship, complete Git, Tekton and Argo CD and wait for the matching
Git revision to be Synced / Healthy. A successful local build alone is not shipping.

The platform has prepared `${APP_NAME}-build` and `${APP_NAME}-dev`, the pipeline
service account and image-pull access. Dev Spaces injects a developer kubeconfig
limited to those two namespaces. Verify that `oc auth can-i create pipelineruns -n ${APP_NAME}-build` and `oc auth can-i create deployments -n ${APP_NAME}-dev` return yes.
If not, stop and ask the facilitator to run the platform developer-preparation step.
Never ask for an administrator login, create namespaces, grant or modify SCCs,
create RBAC bindings, impersonate another identity, or search for credentials.
Never use privileged, anyuid or cluster-admin to work around an error.

Image: `image-registry.openshift-image-registry.svc:5000/${APP_NAME}-build/${APP_NAME}:latest`

The Deployment already has this exact image and `imagePullPolicy: Always`.
Keep it: never use `placeholder`, a bare app name, Docker Hub or an invented tag.
Before the first commit, verify the Deployment image matches IMAGE. For an older
checkout containing `image: placeholder`, replace that value with IMAGE, then
commit and push BEFORE starting Tekton. The pipeline checks the rendered deployment
against its image parameter before buildah; do not bypass this validation.

Use the existing namespaces `${APP_NAME}-build` and `${APP_NAME}-dev`. Never use namespace `default` for the app, Argo CD, or the Application.

1. If `git config user.email` is empty: `git config user.email "dev@workshop.local"` and `git config user.name "developer"`. Then `gitpop init --host "$GIT_SERVER" --name "$APP_NAME"`, run `git branch -M main` for the personal repository, commit, push `main`.
2. In `${APP_NAME}-build`: use the prepared `pipeline` service account and `oc apply -k pipeline/base -n ${APP_NAME}-build`. Start `build-app` with `oc create -f -` (do **not** use `oc create pipelinerun --param` — this `oc` rejects it):

```
GIT_URL=$(git remote get-url origin)
IMAGE=image-registry.openshift-image-registry.svc:5000/${APP_NAME}-build/${APP_NAME}:latest
oc create -n ${APP_NAME}-build -f - <<EOF
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  generateName: build-app-
spec:
  taskRunTemplate:
    serviceAccountName: pipeline
    podTemplate:
      hostUsers: false
      securityContext:
        runAsUser: 0
        runAsGroup: 1000
        fsGroup: 1000
  pipelineRef:
    name: build-app
  params:
    - name: git-url
      value: ${GIT_URL}
    - name: image
      value: ${IMAGE}
  workspaces:
    - name: source
      persistentVolumeClaim:
        claimName: build-ws
EOF
```

The build runs without privileged containers, under the standard `pipelines-scc`.
UID 0 is inside a Linux user namespace, mapped to a non-root host UID. Keep
`hostUsers: false` and the provided pod security context. Do not add privileges.
The source workspace is mounted at `/source` by the platform tasks.

Wait Succeeded. If validate-image fails, fix the image in Git, commit/push and rerun.
Before creating the Argo CD Application, also verify the pushed tag exists:
`oc get imagestreamtag "${APP_NAME}:latest" -n "${APP_NAME}-build" -o name`.
Never start Argo CD delivery before the build and this registry check succeed.

3. Image pull permissions are already prepared by the platform. Do not add RBAC or SCC grants. Verify the preconfigured Deployment image. In `${APP_NAME}-dev`: `oc apply -k gitops/base -n ${APP_NAME}-dev`. Create the Argo CD Application **in ${APP_NAME}-dev** (destination namespace `${APP_NAME}-dev`, repo origin, path `deploy/base`, auto-sync). Wait for the matching Git revision to be Synced / Healthy and the Deployment rollout to finish. Verify its running image equals IMAGE. Only then print app + `argocd-server` Route URLs.

Later change: local build, git push, new PipelineRun (same `oc create -f -`), `oc rollout restart deployment/$APP_NAME -n ${APP_NAME}-dev`.

## OpenShift

Deployment: app requests 50m/64Mi, limits 500m/256Mi, HTTPS probes `/healthz:8443` through the OAuth proxy, `runAsNonRoot` + drop ALL, no `runAsUser`, no Route `spec.host`, TLS reencrypt + Redirect. Keep the OAuth proxy, loopback app listener, mounted runtime Secrets, service CA and cluster-voice ServiceAccount intact. The platform supplies these; never grant monitoring permissions yourself.

## Cluster Voice application contract

Keep the technical app name `live-transcript` and existing Git/Tekton/Argo CD path.
This is one application: speech transcription plus a runtime cluster-health agent.
Copy ALL Go files, go.mod, go.sum, Dockerfile and web assets from the tested reference
linked below as the baseline. Preserve its tool validation, timeout and trace behavior;
do not regenerate the monitoring connector or agent loop from a prose description.
Read the reference README for configuration and tests.

The browser transcribes clips, lets the user check/edit the question, then explicitly
submits it to POST /api/chat. Include text entry and bounded conversation history.
Use the server-side runtime MaaS key file AGENT_KEY_FILE (separate from coding
OPENAI_API_KEY). Read it for every request. Never put model or service account
credentials in the browser. The agent uses DeepSeek tool calling with the bounded
list_metrics, cluster_health and query_metric tools. Only catalog PromQL executes;
no arbitrary queries, URLs, shell, Kubernetes mutation, remediation or SCC changes.
Evidence must show metric ID, unit, exact PromQL, sample/query timestamps and errors.
Missing or stale data is unknown, never healthy/zero. /mcp exposes the same read-only
tools over Streamable HTTP for the OpenShift AI Playground. Keep standard MCP SDK
and OpenTelemetry/MLflow instrumentation, including model/tool latency and trace ID.
Do not log audio, prompts, answers, model results or credentials in traces.

## Speech capture contract

Both coding and runtime use DeepSeek on Foundry through OpenShift AI MaaS, with separate keys and quotas.
Use the injected MaaS base URL and Bearer key; never look up or forward Azure keys. Whisper is a separate OpenShift AI service
used by the application, not by OpenCode. Build a Go browser microphone app.
Embed all HTML/JS/CSS using go:embed: the Dockerfile ships only the binary.
Serve only embedded assets; never serve the project directory or .git files.
Use Web Audio + AudioWorklet to produce independent mono 16 kHz PCM16 WAV clips
(1–4 seconds), POST them sequentially to the Go backend, display returned text.
Do not use the browser SpeechRecognition API or send microphone audio to Foundry.
Keep audio only in memory, skip silence, bound queued clips, flush on Stop, stop
all microphone tracks on Stop/error/page exit, show permission/network errors.
Use textContent for transcripts, never innerHTML. Scripts must be external files
so a restrictive Content-Security-Policy works. No external JS dependencies.

Backend POST /api/transcribe accepts multipart field file (WAV) and language (fr/en/empty), limits
upload size and concurrency, validates mono PCM16 16 kHz WAV duration <=10 seconds,
forwards multipart model=whisper and response_format=json to
WHISPER_URL/audio/transcriptions. Default WHISPER_URL is
http://whisper-predictor.speech-serving.svc.cluster.local/v1 .
Automatic detection means omitting the upstream language field, not language=auto.
Use request deadlines, never log audio or credentials, return useful generic errors.
The browser only calls its own Go origin. No CORS, public model Route or model key
in browser code. Transcription is near-real-time via short clips, not native Whisper
WebSocket streaming. French is the default language; offer English and automatic.

A tested reference is available in the ARO workshop repository at
examples/live-transcript:
https://github.com/fjcloud/aro-ai-app-platform/tree/main/examples/live-transcript
Reuse this tested capture/proxy implementation as the baseline for the app.
