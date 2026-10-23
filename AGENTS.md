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

When asked to ship, complete Git, Tekton and Argo CD and wait for the matching
Git revision to be Synced / Healthy. A successful local build alone is not shipping.

Before deploying, `oc auth can-i '*' '*' --all-namespaces` must return yes.
If not, stop and ask the participant to log the terminal into the assigned lab
administrator using the console's Copy login command. Never search for or print
credentials, tokens or other kubeconfigs. Browser SSO alone does not log in oc.

Image: `image-registry.openshift-image-registry.svc:5000/${APP_NAME}-build/${APP_NAME}:latest`

The Deployment already has this exact image and `imagePullPolicy: Always`.
Keep it: never use `placeholder`, a bare app name, Docker Hub or an invented tag.
Before the first commit, verify the Deployment image matches IMAGE. For an older
checkout containing `image: placeholder`, replace that value with IMAGE, then
commit and push BEFORE starting Tekton. The pipeline checks the rendered deployment
against its image parameter before buildah; do not bypass this validation.

Create namespaces `${APP_NAME}-build` and `${APP_NAME}-dev`. Never use namespace `default` for the app, Argo CD, or the Application.

1. If `git config user.email` is empty: `git config user.email "dev@workshop.local"` and `git config user.name "developer"`. Then `gitpop init --host "$GIT_SERVER" --name "$APP_NAME"`, run `git branch -M main` for the personal repository, commit, push `main`.
2. In `${APP_NAME}-build`: ensure SA `pipeline` exists, grant its build SCC with `oc adm policy add-scc-to-user privileged -z pipeline -n ${APP_NAME}-build`, then `oc apply -k pipeline/base -n ${APP_NAME}-build`. Start `build-app` with `oc create -f -` (do **not** use `oc create pipelinerun --param` — this `oc` rejects it):

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

Wait Succeeded. If validate-image fails, fix the image in Git, commit/push and rerun.
Before creating the Argo CD Application, also verify the pushed tag exists:
`oc get imagestreamtag "${APP_NAME}:latest" -n "${APP_NAME}-build" -o name`.
Never start Argo CD delivery before the build and this registry check succeed.

3. In `${APP_NAME}-build`: `oc policy add-role-to-user system:image-puller system:serviceaccount:${APP_NAME}-dev:default -n ${APP_NAME}-build` (creates a namespaced RoleBinding to the existing ClusterRole `system:image-puller`; do not use `image-puller` without the `system:` prefix). The Deployment image is already correct; verify it before pushing any manifest change. Then in `${APP_NAME}-dev`: `oc apply -k gitops/base -n ${APP_NAME}-dev`. Create the Argo CD Application **in ${APP_NAME}-dev** (destination namespace `${APP_NAME}-dev`, repo origin, path `deploy/base`, auto-sync). Wait for the matching Git revision to be Synced / Healthy and the Deployment rollout to finish. Verify its running image equals IMAGE. Only then print app + `argocd-server` Route URLs.

Later change: local build, git push, new PipelineRun (same `oc create -f -`), `oc rollout restart deployment/$APP_NAME -n ${APP_NAME}-dev`.

## OpenShift

Deployment: requests 50m/64Mi, limits 200m/128Mi, probes `/healthz:8080`, `runAsNonRoot` + drop ALL, no `runAsUser`, no Route `spec.host`, TLS edge + Redirect.
