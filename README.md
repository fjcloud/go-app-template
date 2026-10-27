# Go App Template — ARO Decision Tetris experiment

This branch prepares the existing `live-transcript` scaffold for **Decision Tetris**:
Go game engine → legal placements → local d3-mini on one NVIDIA T4 → typed decision
and probabilities. The browser shows the game, single steps, autoplay, a manual
player and a clearly labelled heuristic baseline.

The coding assistant still uses DeepSeek through OpenShift AI MaaS. Game inference
uses the internal System One endpoint, without a hosted model or runtime API key.
The technical app name, Tekton image guard, personal Git repository, Argo CD and
OpenShift OAuth flow remain the same. The runtime ServiceAccount has image-pull,
OAuth delegation and namespace-scoped MLflow access; it needs no monitoring access.

The facilitator first runs `scripts/prepare-decision-tetris.py` from the matching
[workshop branch](https://github.com/fjcloud/aro-ai-app-platform/tree/feature/decision-tetris).
It installs an experimental custom OpenShift AI runtime on the single T4. This is
a community model/runtime combination with FP16 numerics, not a Red Hat-validated
model. Whisper is stopped to free the GPU; its model/cache remain available.

Copy the complete [tested Go reference](https://github.com/fjcloud/aro-ai-app-platform/tree/feature/decision-tetris/examples/decision-tetris),
then follow `AGENTS.md` to commit, push, build with Tekton and verify Argo Synced/Healthy
at the matching revision. Inspect `decision-tetris` traces in native MLflow.

The published `aro-transcription` branch remains Cluster Voice during this experiment;
ROSA `main` remains unchanged. This template uses fixed namespaces and an ephemeral
Git service: use one assigned cluster per participant and export code before expiry.

```sh
python3 -m pip install -r tests/requirements.txt
python3 -m unittest discover -s tests
```
