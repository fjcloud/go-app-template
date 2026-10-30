# Card Scout — ARO workshop template

Generate a small Go app with OpenCode and DeepSeek: analyze a mixed-era card
collection, browse the rarity ranking, then ask an OpenShift AI agent for dated
market reference prices.

The platform provides the dataset, D3-mini on one T4, the DeepSeek agent, its
read-only MCP tool and MLflow traces. The app only needs a UI and a same-origin
proxy to `CARD_SCOUT_URL`.

1. Complete the [ARO platform workshop](https://fjcloud.github.io/aro-ai-app-platform/).
2. Launch this template in Dev Spaces using the `aro-signal-desk` branch.
3. Read [APP_CONTRACT.md](APP_CONTRACT.md), then ask OpenCode to build the app
   from scratch and ship it using [AGENTS.md](AGENTS.md).

The existing delivery scaffold uses the technical name `signal-desk`. Keep its
namespaces, internal registry image, restricted SCC and HTTPS Route. The generated
application has no model key, Kubernetes token or persistent volume.

The main branch of this repository belongs to the ROSA workshop. This ARO branch
contains the Card Scout contract and delivery scaffold; no completed app is bundled.
