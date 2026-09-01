# AGENTS.md — Go App Platform Conventions

You are in an OpenShift Dev Space on ROSA. `oc` is already authenticated.
`$GIT_SERVER` and `$APP_NAME` are set (default `APP_NAME=fortune-cookie`).

Do **not** create Ansible playbooks or a `scripts/` directory. Follow the commands in this file with your bash tool. Wait for each phase to succeed before the next. If a command fails, read the error, fix the cause, retry that phase.

## Files YOU must generate

### main.go
HTTP server on port 8080:
- `GET /` — application UI or API
- `GET /healthz` — `{"status":"ok"}` (Kubernetes probes)

### go.mod
```
module <app-name>
go 1.22
```

### Dockerfile
Use this two-stage template. Change only `COPY` / `RUN` if there are extra source files:

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

- Build in `/tmp/build` (go-toolset is UID 1001; cannot write to `/app`)
- `-buildvcs=false` (Tekton git-clone workspace has no full git metadata)
- `USER 1001` in the image; OpenShift assigns a namespace UID at runtime

## Files already in the repository (do not rewrite)

```
deploy/base/      # app manifests — only change the image: field when deploying
pipeline/base/    # Tekton Pipeline + PVC
gitops/base/      # developer Argo CD instance + AppProject
devfile.yaml      # Dev Spaces workspace
```

## Verify the binary

After generating files, always:

```bash
CGO_ENABLED=0 go build -buildvcs=false -o /dev/null .
```

Fix compile errors before deploying.

## OpenShift constraints (`deploy/base`)

Every Deployment MUST have:
- resources.requests: cpu 50m, memory 64Mi
- resources.limits: cpu 200m, memory 128Mi
- livenessProbe + readinessProbe: HTTP GET `/healthz` on port 8080
- label `app.kubernetes.io/name: <app>`
- Route: `tls.termination: edge`, `insecureEdgeTerminationPolicy: Redirect`
- Do NOT set `spec.host` — OpenShift generates the hostname
- Do NOT set `runAsUser` — restricted-v2 SCC assigns the UID
- Pod securityContext: `runAsNonRoot: true`, `seccompProfile.type: RuntimeDefault`
- Container securityContext: `allowPrivilegeEscalation: false`, `capabilities.drop: ["ALL"]`, `runAsNonRoot: true`

## Deploy — three phases, git + oc only

### Phase 1 — Push to the personal Git server

```bash
git remote remove origin 2>/dev/null || true
gitpop init --host "$GIT_SERVER" --name "$APP_NAME"
git config user.email "${GIT_EMAIL:-dev@workshop.local}"
git config user.name "${GIT_NAME:-Developer}"
git add -A
git diff --staged --quiet || git commit -m "feat: $APP_NAME initial implementation"
git push -u origin main
git remote get-url origin
```

### Phase 2 — Build the image (Tekton)

```bash
BUILD_NS="${APP_NAME}-build"
IMAGE="image-registry.openshift-image-registry.svc:5000/${APP_NAME}-build/${APP_NAME}:latest"
REPO=$(git remote get-url origin)

oc new-project "$BUILD_NS" 2>/dev/null || true
oc project "$BUILD_NS"
oc adm policy add-scc-to-user privileged -z pipeline -n "$BUILD_NS"
oc create rolebinding pipeline-registry-editor \
  --clusterrole=registry-editor \
  --serviceaccount="${BUILD_NS}:pipeline" \
  -n "$BUILD_NS" 2>/dev/null || true
oc apply -k pipeline/base -n "$BUILD_NS"

PR=$(oc create -n "$BUILD_NS" -o name -f - <<EOF
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  generateName: build-app-
spec:
  pipelineRef:
    name: build-app
  params:
    - name: git-url
      value: ${REPO}
    - name: image
      value: ${IMAGE}
  workspaces:
    - name: source
      persistentVolumeClaim:
        claimName: build-ws
EOF
)
echo "Waiting for $PR ..."
oc wait -n "$BUILD_NS" "$PR" --for=condition=Succeeded --timeout=15m
```

### Phase 3 — Deploy with your Argo CD instance

```bash
DEV_NS="${APP_NAME}-dev"
BUILD_NS="${APP_NAME}-build"
IMAGE="image-registry.openshift-image-registry.svc:5000/${APP_NAME}-build/${APP_NAME}:latest"
REPO=$(git remote get-url origin)

oc new-project "$DEV_NS" 2>/dev/null || true
oc create rolebinding image-puller \
  --clusterrole=system:image-puller \
  --serviceaccount="${DEV_NS}:default" \
  -n "$BUILD_NS" 2>/dev/null || true

sed -i "s|image: .*|image: ${IMAGE}|" deploy/base/deployment.yaml
git add deploy/base/deployment.yaml
git diff --staged --quiet || git commit -m "ci: update image to ${APP_NAME}:latest"
git push

oc apply -k gitops/base -n "$DEV_NS"
oc wait -n "$DEV_NS" argocd/argocd --for=jsonpath='{.status.phase}'=Available --timeout=300s

oc apply -n "$DEV_NS" -f - <<EOF
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: ${APP_NAME}
  namespace: ${DEV_NS}
spec:
  project: default
  source:
    repoURL: ${REPO}
    targetRevision: main
    path: deploy/base
  destination:
    server: https://kubernetes.default.svc
    namespace: ${DEV_NS}
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
EOF

oc wait -n "$DEV_NS" "application.argoproj.io/${APP_NAME}" \
  --for=jsonpath='{.status.health.status}'=Healthy --timeout=300s

echo "App:    https://$(oc get route $APP_NAME -n $DEV_NS -o jsonpath='{.spec.host}')"
echo "ArgoCD: https://$(oc get route argocd-server -n $DEV_NS -o jsonpath='{.spec.host}')"
```

## Ship a later change

1. `CGO_ENABLED=0 go build -buildvcs=false -o /dev/null .`
2. `git add -A && git commit -m "feat: update $APP_NAME" && git push`
3. Repeat Phase 2 (new PipelineRun)
4. `oc rollout restart deployment/$APP_NAME -n ${APP_NAME}-dev` so pods pull `:latest`
