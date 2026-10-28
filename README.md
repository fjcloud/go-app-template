# Signal Desk / Techmeme terminal

A compact English research terminal for the ARO workshop.

1. **Fetch Techmeme** to retrieve and summarize technology news.
2. Read **BUY / SHORT / NO SIGNAL** rankings from the local T4 decision model.
3. **Investigate** a directional story to examine an index-proxy thesis,
   timestamped reference price and conditional target/stop P&L.
4. Edit scenario assumptions or export the research note as Markdown.

The app uses native OpenShift AI Responses and read-only tools. Entry prices come
from quote evidence, while targets/stops are hypothetical assumptions. There is no
calibrated expected-return estimate or order execution.

The facilitator prepares the existing model/agent, 24h MaaS key and persistent
storage. The public service stays behind OAuth under restricted-v2. Keep the
live-transcript image and Git → Tekton → Argo CD contract in AGENTS.md.

The working Go reference and clearly labeled local preview are in
[the workshop repository](https://github.com/fjcloud/aro-ai-app-platform/tree/feature/briefing-desk/examples/briefing-desk).
