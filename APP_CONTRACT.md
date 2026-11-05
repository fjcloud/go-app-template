# Card Scout — minimal app

Generate this English Go web app **from scratch**. The platform already owns
images, D3 classification, the OpenShift AI agent, prices, state and MLflow.
Write only a small Go proxy and an embedded web page. Keep APP_NAME=card-scout.

## One page, three actions

1. On opening **Card Scout**, load the batch and show its progress as
   `completed/dataset.items`. Display all cards in the API's ranking order:
   name, full image and rarity (Common, Uncommon, Rare, Uncertain or N/A).
   Use a responsive grid and lazy-loaded images; never crop the cards.
2. **Analyze batch** starts analysis. Disable the button while running and poll
   about once a second. Reloading must recover the platform's current results.
3. **Check price** is available only for `status=done`, `kind=pokemon_tcg`.
   Show the selected card name, Cardmarket trend in EUR, provider update date,
   the returned note, **Source** and **View MLflow trace** links when supplied.
   Read numeric evidence from the API: zero is 0; null is Unavailable.

Show a visible error (`role=alert`) and a working **Retry** button on failure.
Keep loaded cards and the selected card name. Retries must survive repeated
errors. Make one price request at a time. Safely display text; accept only HTTPS
source/trace links and fixed same-origin `/assets/item-NNN` image paths.

No filters, pagination, localStorage, uploads, login, extra screens or features.

## Backend contract

`CARD_SCOUT_URL` is the only upstream setting (provided by the platform).
`LISTEN_ADDR` defaults to `:8080`. The browser calls the Go app only.

| App route | Behavior |
| --- | --- |
| GET / | Embedded page; other unknown paths return 404 |
| GET /healthz | Local JSON `{"status":"ok"}` |
| GET /api/batch | Proxy unchanged |
| POST /api/batch/analyze | Proxy unchanged (202) |
| GET /assets/item-NNN | Proxy unchanged; exactly three digits |
| POST /api/cards/item-NNN/research | Proxy unchanged |

Preserve status, content type and complete bodies, including images over 2 MiB.
Do not follow upstream redirects or proxy arbitrary URLs. Allow up to 210 seconds
for research, then return a visible failure. Never invent classifications/prices.

Batch JSON: `dataset.items`, `running`, `completed`, `failed`, `items`.
Each item has `id`, `name`, `image_url`, `status`, `kind`, `rarity`, `research`.
Rarity values: `common`, `uncommon`, `rare_or_higher`, `uncertain`, `not_applicable`.
Research JSON: `card.name`, `trend`, `provider_updated_at`, `note`, `source_url`,
`trace_url`. Errors are `{"error":"..."}`. The provided tests contain examples.
Prices are catalogue aggregates, not verified 1st Edition valuations: show the note.

## Stable test hooks

Use `data-testid`: `analyze` (button), `progress`, `card` (each card container),
`card-image` (img), `rarity`, `price` (selected-card result), `trend` (value only).
Each card container also has `data-item-id` with its API ID.
Button/link text is the bold text above. Layout and styling are up to you.

## Acceptance

Run `npm test` against the actual Go binary with a local fixture API and Chromium.
Fix app failures without editing tests. Then follow DELIVERY.md and run
`APP_URL=https://your-route npm run test:live` against the real platform.
The fixture data belongs only to the test runner, never to the generated app.
