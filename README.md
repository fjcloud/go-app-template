# Go App Template — AI-Powered Developer Platform on ROSA

This is the **template repository** for the *AI-Powered Developer Platform on ROSA* workshop.

Developers clone this template into their OpenShift DevSpace, then use **OpenCode** (an AI coding assistant backed by the in-cluster Qwen3.6 LLM) to generate a Fortune Cookie Go application and deploy it to their own namespace.

## What's in this repo

| Path | Description |
|------|-------------|
| `AGENTS.md` | Platform conventions read by OpenCode before every session |
| `opencode.json` | LLM configuration pointing at the in-cluster Qwen3.6 service |
| `devfile.yaml` | DevSpaces workspace definition (installs tools, exposes tasks) |
| `deploy/base/` | Kustomize manifests for the application (Deployment, Service, Route) |
| `pipeline/base/` | Kustomize manifests for the Tekton build pipeline |
| `gitops/base/` | Kustomize manifests for the developer-owned Argo CD instance |
| `scripts/git-push.yml` | Ansible: create personal repo and push code |
| `scripts/build-image.yml` | Ansible: run Tekton PipelineRun (git-clone + buildah) |
| `scripts/gitops-deploy.yml` | Ansible: spin up Argo CD instance and sync the application |

## How it works

1. **Platform Engineer** initializes this template and publishes it to the Git server
2. **Developer** launches a DevSpace from this template URL
3. **OpenCode** reads `AGENTS.md` and generates `main.go`, `go.mod`, and `Dockerfile`
4. **OpenCode** runs the three Ansible playbooks sequentially to build and deploy the app

## Prerequisites

- Access to an OpenShift cluster with ROSA
- OpenShift AI (RHOAI) with Qwen3.6 model deployed in `llm-inference` namespace
- OpenShift Pipelines and OpenShift GitOps operators installed
- OpenShift Dev Spaces operator installed

## Workshop

Full workshop instructions: https://fjcloud.github.io/rosa-ai-app-platform/
