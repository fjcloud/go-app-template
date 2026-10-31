# Card Scout

A small English Go UI and same-origin proxy for the prepared ARO Card Scout
platform: analyze a mixed-era collection, browse predicted rarity, and request
verified, dated market reference prices through the native OpenShift AI agent.

The application was generated from scratch by OpenCode 1.18.14 with
`deepseek/deepseekv4pro`. The first two commits contain the generated application
and its review corrections. A maintainer follow-up fixes hidden controls, exact
item-ID lengths and a redirect-sensitive test.

This branch preserves the generated validation result. Workshop participants
start from the [ARO template contract](https://github.com/fjcloud/go-app-template/blob/aro-signal-desk/APP_CONTRACT.md)
and generate their own implementation.

```sh
go test -race ./...
go vet ./...
CARD_SCOUT_URL=http://card-scout-platform.signal-desk-dev.svc.cluster.local:8080 go run .
```

The [ARO workshop](https://fjcloud.github.io/aro-ai-app-platform/) prepares that
service, models, MCP tools, state and MLflow. The generated app needs no model
key or Kubernetes token. It defaults to HTTP port 8080; `LISTEN_ADDR` overrides it.

The deployment scaffold keeps the technical name `signal-desk`, its restricted
security context, internal registry image, Tekton and Argo CD delivery path.
