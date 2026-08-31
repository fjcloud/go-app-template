# Go App Template — AI-Powered Developer Platform on ROSA

This is the **template repository** for the *AI-Powered Developer Platform on ROSA* workshop.

Developers launch this template in OpenShift Dev Spaces. **OpenCode** is injected by the platform (Dev Spaces AI tool registry) and talks to the in-cluster Qwen3.8 LLM. The developer generates a Fortune Cookie Go application and deploys it to their own namespace.

## What's in this repo

| Path | Description |
|------|-------------|
| `AGENTS.md` | Platform conventions read by OpenCode before every session |
| `devfile.yaml` | Dev Spaces workspace (gitpop + Ansible + one-click tasks). OpenCode is not installed here. |
| `deploy/base/` | Kustomize manifests for the application (Deployment, Service, Route) |
| `pipeline/base/` | Kustomize manifests for the Tekton build pipeline |
| `gitops/base/` | Kustomize manifests for the developer-owned Argo CD instance |
| `scripts/git-push.yml` | Ansible: create personal repo and push code |
| `scripts/build-image.yml` | Ansible: run Tekton PipelineRun (git-clone + buildah) |
| `scripts/gitops-deploy.yml` | Ansible: spin up Argo CD instance and sync the application |

## How it works

1. **Platform Engineer** publishes this template and registers OpenCode + Qwen3.8 in Dev Spaces
2. **Developer** launches a Dev Space from the template URL (`?ai-provider=opencodeai/opencode`)
3. **OpenCode** reads `AGENTS.md` and generates `main.go`, `go.mod`, and `Dockerfile`
4. **OpenCode** runs the three Ansible playbooks sequentially to build and deploy the app

## Prerequisites

- Access to an OpenShift cluster with ROSA
- OpenShift AI (RHOAI) with Qwen3.8 model deployed in `llm-serving` namespace
- OpenShift Pipelines and OpenShift GitOps operators installed
- OpenShift Dev Spaces 3.29+ with the workshop AI tool registry applied

## Workshop

Full workshop instructions: https://fjcloud.github.io/rosa-ai-app-platform/
