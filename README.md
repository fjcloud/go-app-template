# Go App Template — ARO Orbit Checkpoint experiment

Inspect passport and cargo-permit images, apply the current shift directive and
stamp **Clear**, **Inspect** or **Reject** before local d3-mini reveals its choice.
Play a 12-courier shift against the AI, or watch it run the checkpoint.

The Go server owns documents, rules and scoring. Both players see the same image;
the model never receives extracted fields or the answer key. Its mistakes remain
visible. A failed model call leaves the courier waiting, without a fallback verdict.

The coding assistant still uses DeepSeek through OpenShift AI MaaS. Game inference
uses the existing T4 and internal System One endpoint. Keep the technical
`live-transcript` app name, existing `decision-tetris` runtime identity/config,
OpenShift OAuth and personal Git → Tekton → Argo CD delivery path.

From the matching [workshop branch](https://github.com/fjcloud/aro-ai-app-platform/tree/feature/orbit-checkpoint),
the facilitator prepares the local runtime and native MLflow experiment:

```sh
python3 scripts/prepare-decision-tetris.py --reuse-runtime --game orbit-checkpoint
```

Omit `--reuse-runtime` when the custom runtime image has not been built yet.
Copy the complete [Go reference](https://github.com/fjcloud/aro-ai-app-platform/tree/feature/orbit-checkpoint/examples/orbit-checkpoint),
then follow `AGENTS.md` to build and ship. Inspect `orbit-checkpoint` traces in
native MLflow. The community runtime uses FP16 on T4; this is an experiment,
not a Red Hat validated model/runtime combination.

The stable `aro-transcription` guide and ROSA `main` remain unchanged. The earlier
Tetris prototype remains on `feature/decision-tetris`.

```sh
python3 -m pip install -r tests/requirements.txt
python3 -m unittest discover -s tests
```
