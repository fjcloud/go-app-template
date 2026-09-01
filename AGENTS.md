# AGENTS.md

OpenShift Dev Space. `oc` is authenticated. `$GIT_SERVER` and `$APP_NAME` are set.

No Ansible. No `scripts/`. Do not rewrite `deploy/`, `pipeline/`, `gitops/`, `devfile.yaml`, or `opencode.json` (only change `image:` in `deploy/base/deployment.yaml` when shipping).

## Generate

`main.go` — HTTP `:8080`, `GET /`, `GET /healthz` → `{"status":"ok"}`
`go.mod` — `module <app-name>` / `go 1.22`

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

Image: `image-registry.openshift-image-registry.svc:5000/${APP_NAME}-build/${APP_NAME}:latest`

1. `gitpop init --host "$GIT_SERVER" --name "$APP_NAME"`, commit, push `main`.
2. Namespace `${APP_NAME}-build`: privileged SCC + registry-editor on SA `pipeline`, `oc apply -k pipeline/base`, PipelineRun of `build-app` (params `git-url` = origin, `image` = above; workspace PVC `build-ws`). Wait Succeeded.
3. Namespace `${APP_NAME}-dev`: image-puller for `default` from the build ns, set Deployment `image:`, push, `oc apply -k gitops/base`, Argo CD Application (repo origin, path `deploy/base`, auto-sync). Wait Healthy. Print app + `argocd-server` Route URLs.

Later change: local build, git push, new PipelineRun, `oc rollout restart deployment/$APP_NAME -n ${APP_NAME}-dev`.

## OpenShift

Deployment: requests 50m/64Mi, limits 200m/128Mi, probes `/healthz:8080`, `runAsNonRoot` + drop ALL, no `runAsUser`, no Route `spec.host`, TLS edge + Redirect.
