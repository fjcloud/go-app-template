# Go App Template — AI-Powered Developer Platform on ROSA

This is the **template repository** for the *AI-Powered Developer Platform on ROSA* workshop.

Developers launch this template in OpenShift Dev Spaces. **OpenCode** is injected by the platform (Dev Spaces AI tool registry) and talks to the in-cluster **Qwen3.8** LLM. The developer generates a Fortune Cookie Go application and deploys it to their own namespace.

## What's in this repo

| Path | Description |
|------|-------------|
| `AGENTS.md` | Project instructions and delivery steps for OpenCode |
| `devfile.yaml` | Dev Spaces workspace (installs gitpop on start). OpenCode is not installed here. |
| `deploy/base/` | Kustomize manifests for the application (Deployment, Service, Route) |
| `pipeline/base/` | Kustomize manifests for the Tekton build pipeline |
| `gitops/base/` | Kustomize manifests for the developer-owned Argo CD instance |
| `opencode.json` | Pins OpenCode to in-cluster Qwen3.8 (MaaS). API key comes from `$OPENAI_API_KEY`. |

## How it works

1. **Platform Engineer** publishes this template and registers OpenCode + Qwen3.8 in Dev Spaces
2. **Developer** launches a Dev Space from the template URL (`?ai-provider=opencodeai/opencode`)
3. **OpenCode** (Qwen3.8) reads `AGENTS.md` and generates `main.go`, `go.mod`, and `Dockerfile`
4. **Developer** authenticates `oc` inside the workspace; **OpenCode** follows `AGENTS.md` to push to ephemeral Git, build with Tekton and deploy through Argo CD
5. **Developer** verifies the matching Git revision, live app and platform conventions

## Prerequisites

- Access to the assigned ROSA HCP cluster on OpenShift 4.22 or later, including administrator access in the workspace terminal
- OpenShift AI (RHOAI) with Qwen3.8 model deployed in `llm-serving` namespace
- OpenShift Pipelines and OpenShift GitOps operators installed
- OpenShift Dev Spaces 3.29+ with the workshop AI tool registry applied

## Workshop

Full workshop instructions: https://fjcloud.github.io/rosa-ai-app-platform/

The model is used for coding; Fortune Cookie serves local strings without runtime inference. The template pushes source to an external ephemeral Git service. Export work before the RHDP assignment expires. For ARO Live Transcript, use the [`aro-transcription` branch](https://github.com/fjcloud/go-app-template/tree/aro-transcription) and the ARO guide.
