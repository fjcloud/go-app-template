# Card Scout — ARO template

Generate one English Go page from scratch: show the cards, analyze them, check a
price with its source and MLflow trace. The platform owns the models and data.

1. Complete the [ARO platform workshop](https://fjcloud.github.io/aro-ai-app-platform/).
2. Open this branch in Dev Spaces and ask OpenCode to follow [AGENTS.md](AGENTS.md).
3. It runs the provided `npm test`, fixes app failures, then ships using
   [DELIVERY.md](DELIVERY.md) and runs `APP_URL=https://your-route npm run test:live`.

The short [contract](APP_CONTRACT.md) defines the app. Tests contain only fixture
API data and acceptance checks; no completed app is bundled. Keep tests unchanged.
`npm test` builds and runs the actual Go app against a local fixture API and a
browser; `test:live` verifies the deployed app against the real platform.

Dev Spaces includes a local-only test-browser component. For local development,
run `npm ci && npm run setup` to install Chromium instead. The app image remains
Go only. No application credentials, filters, pagination or extra services to build.

The main branch belongs to ROSA. This ARO template uses `aro-card-scout`.
