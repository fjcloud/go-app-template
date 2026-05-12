# AGENTS.md — Go App Platform Conventions

You are running inside an OpenShift DevSpace on ROSA. `oc` is authenticated.
`$GIT_SERVER` and `$APP_NAME` are set in the terminal.

## Files YOU must generate

For every new application, generate these files:

### main.go
Entry point. HTTP server on port 8080. Must expose:
- `GET /` — the application UI or API response
- `GET /healthz` — returns `{"status":"ok"}` (used by Kubernetes probes)

### go.mod
```
module <app-name>
go 1.22
```

### Dockerfile
Use exactly this two-stage template — adapt only the `COPY` and `RUN` lines if the app has additional source files:

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

Key constraints:
- Build in `/tmp/build` (go-toolset runs as UID 1001, cannot write to `/app`)
- `-buildvcs=false` required (Tekton git-clone workspace lacks full git metadata)
- `USER 1001` in image; OpenShift overrides at runtime with a namespace UID

## Files already in the repository (do not regenerate)

```
deploy/base/          # Kustomize manifests for the app — only update image: field
pipeline/base/        # Kustomize manifests for the Tekton build pipeline
gitops/base/          # Kustomize manifests for the developer Argo CD instance
scripts/
  git-push.yml        # Ansible: creates personal Git repo + pushes code
  build-image.yml     # Ansible: applies pipeline/base + Tekton PipelineRun
  gitops-deploy.yml   # Ansible: applies gitops/base + ArgoCD Application
```

## Build verification

After generating all files, always run:
```bash
CGO_ENABLED=0 go build -buildvcs=false -o /dev/null .
```
Fix any compile errors before proceeding.

## Kubernetes / OpenShift requirements
Every Deployment MUST have:
- resources.requests: cpu 50m, memory 64Mi
- resources.limits: cpu 200m, memory 128Mi
- livenessProbe + readinessProbe on /healthz port 8080
- Label app.kubernetes.io/name: <app-name>
- Route with tls.termination: edge, insecureEdgeTerminationPolicy: Redirect
  Do NOT set spec.host — leave it empty so OpenShift auto-generates the hostname.
- Do NOT set runAsUser — OpenShift assigns a UID from the namespace range (restricted-v2 SCC).
- Pod securityContext:
    runAsNonRoot: true
    seccompProfile:
      type: RuntimeDefault
- Container securityContext:
    allowPrivilegeEscalation: false
    capabilities:
      drop: ["ALL"]
    runAsNonRoot: true

## Deploy workflow — run the Ansible playbooks

Once the build is verified, run these playbooks in order using your bash tool:

```bash
ansible-playbook scripts/git-push.yml
ansible-playbook scripts/build-image.yml      # takes 3-5 min — Tekton PipelineRun
ansible-playbook scripts/gitops-deploy.yml    # takes 2-3 min — ArgoCD startup
```

Each playbook prints named task output. Wait for it to succeed before running the next.
If a task fails, the error is shown inline — read it and fix the root cause.

## In-cluster LLM service
- Base URL : http://qwen36-predictor.llm-inference.svc.cluster.local:8080/v1
- Model ID  : qwen36
- API       : OpenAI-compatible
- Always add `"chat_template_kwargs": {"enable_thinking": false}` to every request body
