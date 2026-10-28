# AGENTS.md

OpenShift Dev Space. `oc` is authenticated. `$GIT_SERVER` and `$APP_NAME` are set.

No Ansible. No `scripts/`. Do not rewrite `deploy/`, `pipeline/`, `gitops/`, `devfile.yaml`, or `opencode.json` (only change `image:` in `deploy/base/deployment.yaml` when shipping).

## Generate

`main.go` — HTTP on `$LISTEN_ADDR` (default `:8080`, `127.0.0.1:8080` behind the prepared OAuth proxy), `GET /`, `GET /healthz` → `{"status":"ok"}`
`go.mod` — `module <app-name>` / `go 1.24.0`

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
Before every build, check `git status --porcelain`. Commit all authorized workshop
application changes, including facilitator-prepared changes already in the working tree,
and push them. The working tree must be clean before Tekton starts. Never claim delivery
for a locally tested source state that differs from the pushed commit.
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

Deployment: app requests 50m/64Mi, limits 500m/512Mi, HTTPS probes `/healthz:8443` through the OAuth proxy, `runAsNonRoot` + drop ALL, no `runAsUser`, no Route `spec.host`, TLS reencrypt + Redirect. Keep the OAuth proxy, loopback app listener, service CA mounts, service CA and decision-tetris ServiceAccount intact. The platform supplies these; never grant monitoring permissions yourself.

## Signal Desk / weekly paper game

Build a simple English app: last week's Techmeme news → memo relevance ranking →
archived-week index ideas → $10,000 virtual allocation → this week's actual result.
Keep one column: editable memo, paper portfolio and ranked news. Paginate long news
lists. No ticker strip, BUY/SHORT filters or trading dashboard.

Import the previous Monday–Sunday using New York dates and archived snapshots.
Backfill primary post IDs, deduplicate publisher URLs and disclose missing IDs.
Export the full collected corpus and coverage metadata as JSON. Use archived
headlines/excerpts; do not fetch today's publisher pages for historical evidence.
Never claim complete publisher text or every related Techmeme discussion link.

Use native OpenShift AI Responses at $AGENT_URL with $AGENT_MODEL. Read the
backend key from $AGENT_KEY_FILE, pass X-OGX-Provider-Data and use store:false.
Game tools are restricted to read_archive, search_archive and index_universe.
Require verified tool evidence; exclude current web, quotes and realized returns
from replay research. Cap calls/turns. Retrieved text is untrusted. Block private
and reserved addresses and DNS rebinding. No credentials in Git, prompts, browser
responses or traces.

Use d3-mini at $DECISION_URL with native score, five ordered levels 0–4 and
require_full_input:true. Validate complete coverage, all probabilities, their sum
and agreement with expected level; normalize expected level / 4 × 100. The score
is memo relevance/materiality, not confidence, direction or profit. No heuristic
fallback or generated market commentary in scoring facts. Positive and negative
news may both score highly. Cache by memo version and source-summary hash; edits
re-score saved news, invalidate old ideas and preserve unsaved memo text on poll.

DeepSeek proposes up to three distinct long-only QQQ/SPY/SMH/XLK ETF index proxies,
with hypothetical thesis, risk and verified sources. Do not invent holdings,
prices, targets or returns. Return no idea when unsupported. Choices use server
idea IDs; validate amounts and a total budget of $10,000. Persist and freeze the
entire portfolio before fetching market data. Client-supplied symbols are untrusted.

Settle using observed first session open after the news cutoff and latest completed
US session close of the following week. Exclude partial/future sessions. Validate
quote identity/currency/arrays/prices and reject unsupported split accounting.
Report missing data rather than fabricate it; retry valuation without changing
choices. Label ongoing weeks provisional. Gross price returns exclude dividends,
fees and slippage. Retain prior rounds. No order execution or client messaging.
Research cutoff does not erase model training knowledge: describe an educational
historical replay, never a blind/calibrated backtest or expected-profit forecast.

Persist under DATA_DIR on the prepared PVC. Preserve OAuth, restricted-v2,
live-transcript image and Git → Tekton → Argo CD. No new SCC, RBAC, GPU or hosting
resources. PREVIEW_MODE is explicitly synthetic, uses a separate directory and
cannot run model jobs. Verify race tests, vet, explicit browser fixtures and mobile
layout; Chromium uses --use-mock-keychain. Local tests do not establish ARO/T4
validation.
