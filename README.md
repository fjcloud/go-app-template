# Go App Template — ARO Sky Control experiment

Watch local d3-mini run a fictional airport: assign landing runways, release
departures or keep aircraft in holding patterns. The radar shows live trajectories
and each applied/rejected decision. Increase traffic and simulation speed until
the decision model cannot keep pace.

A Go engine owns movement, runway slots and separation. Inference runs
asynchronously on the existing T4. Overdue, unsafe or stale commands do not take
effect; an explicit safety shield prevents crashes. Its interventions, airport
backlog and model deadline misses have separate counters. No fallback AI player.

Keep the technical live-transcript app, decision-tetris runtime identity/config,
OpenShift OAuth and personal Git → Tekton → Argo CD flow. Coding uses DeepSeek/MaaS;
game decisions use only the internal local model, without another key or GPU.

The facilitator prepares the matching [workshop branch](https://github.com/fjcloud/aro-ai-app-platform/tree/feature/sky-control):

```sh
python3 scripts/prepare-decision-tetris.py --reuse-runtime --game sky-control
```

Copy the [complete Go reference](https://github.com/fjcloud/aro-ai-app-platform/tree/feature/sky-control/examples/sky-control),
then follow AGENTS.md to ship. Native MLflow's sky-control experiment contains
atc.cycle and decision.choose spans. The community runtime uses approximate FP16
numerics; this is a game experiment, not a Red Hat validated model/runtime or
an operational aviation system. The stable guide and earlier prototypes remain
on their existing branches; ROSA main is unchanged.

```sh
python3 -m unittest discover -s tests
```
