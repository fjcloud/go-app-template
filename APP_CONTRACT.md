# Card Scout — app contract

Build a small English Go web application **from scratch**. The platform is already
prepared. Your job is the UI and a small backend proxy, not model orchestration.
Keep the delivery scaffold and technical `APP_NAME=signal-desk`.

## Experience

A mixed box of cards → Analyze batch → Ranked cards → Check price.

- Title: **Card Scout**. Show the dataset description and analysis progress.
- One primary **Analyze batch** button. Start analysis, then poll the batch API
  about once a second while running. Disable duplicate starts. Completed results
  are cached by the platform; refresh/reload must recover the current state.
- A responsive card grid: image, catalogue name, set/year, D3 rarity badge and
  **Check price**. Show 12 cards initially and a **Show more** button.
- Simple filters: **All**, **Rare**, **Needs review**, **Other**. Filter the full
  batch before pagination. Show the matching count and a useful empty state.
  Rare means `kind=pokemon_tcg` and `rarity=rare_or_higher`; Needs review means
  `needs_review=true`; Other means `other_card` or `not_a_card`.
- Use the platform's ranking order. D3 predicts card kind and printed rarity;
  its probabilities are not a price, market value or calibrated certainty.
- Enable Check price only for successfully analyzed Pokémon items. Make one
  request at a time, show progress, and preserve the selected card on errors.
- Show a compact selected-card result: exact name, set, collector number,
  **catalogue rarity**, Cardmarket EUR **Trend**, **7-day average**, provider
  update date, **Source**, and **View MLflow trace** when a URL is supplied.
  Display zero as zero and null as **Unavailable**. Retain the source timestamp;
  it is not necessarily today's price. Preserve a differing D3 prediction.
- Numeric fields must come from the platform's verified tool evidence. Do not
  infer prices from the agent prose; keep that prose out of the price display.
- Keep loading, partial results, failed items, retry and empty states visible.
  Do not turn model/network failures into fabricated classifications or prices.
- No login, microphone, uploads, portfolio, charts, trading, news feed or API-key
  form. Use plain HTML/CSS/JavaScript, no frontend framework or CDN dependency.

## Prepared dataset and platform

The platform contains 48 English Pokémon cards across eight sets (1999–2023),
plus a Yu-Gi-Oh! card, a receipt and a standalone Pokémon illustration. Images
are preloaded. D3 receives image bytes only; held-out catalogue labels are not
sent to it. Catalogue IDs are known from this curated dataset: do not claim
that the model recognized the exact edition of an arbitrary uploaded photo.

The platform runs D3 on the T4, serves the dataset, persists progress, calls the
native OpenShift AI Responses agent and hosts its read-only price MCP tool.
It owns all model credentials and MLflow tracing. **Do not implement any of
those functions or copy their source into this application.**

The only backend setting is `CARD_SCOUT_URL`, supplied by `signal-desk-config`:

```
http://card-scout-platform.signal-desk-dev.svc.cluster.local:8080
```

Keep this URL and all platform calls server-side. The browser talks only to the
same-origin Go app. No model key, OpenShift token or Azure credential is needed
by the generated app. Use `LISTEN_ADDR` (default `:8080`).

## API to proxy

Preserve upstream HTTP status, JSON errors and image content types. Use bounded
requests/timeouts. Only proxy these fixed routes to the configured platform;
never accept an arbitrary caller URL or create a public general-purpose proxy.

| Application route | Prepared platform route |
| --- | --- |
| `GET /` | Embedded UI |
| `GET /healthz` | Local `{"status":"ok"}` |
| `GET /api/batch` | `GET /api/batch` |
| `POST /api/batch/analyze` | `POST /api/batch/analyze` (202) |
| `GET /assets/item-NNN` | `GET /assets/item-NNN` |
| `POST /api/cards/item-NNN/research` | Same POST path |

`GET /api/batch` returns:

```json
{
  "dataset": {"name":"Mixed-era collection","items":51,"pokemon_cards":48,"controls":3,"years":"1999–2023","snapshot_at":"2026-10-30T15:30:47.498665+00:00","identity_method":"Known catalogue references in the prepared dataset"},
  "running": false, "completed": 1, "failed": 0,
  "trace_id": "tr-example", "trace_url": "",
  "items": [{"id":"item-039","image_url":"/assets/item-039","name":"Toedscool","set_name":"Scarlet & Violet","year":"2023","status":"done","error":null,"kind":"pokemon_tcg","rarity":"rare_or_higher","needs_review":false,"probabilities":{},"research":null}]
}
```

Items are already ranked. `status` is `pending`, `analyzing`, `done` or `error`.
Before analysis, `kind` and `rarity` are null. `kind` can be `pokemon_tcg`,
`other_card`, `not_a_card` or `uncertain`. Rarity can be `common`, `uncommon`,
`rare_or_higher`, `not_applicable` or `uncertain`. Read dataset counts from the
response. A later GET includes saved successful price research per item.

A successful research response contains:

```json
{
  "status":"ok",
  "card":{"name":"Toedscool","set":"Scarlet & Violet","set_id":"sv01","number":"201","rarity":"Illustration rare"},
  "provider":"Cardmarket","currency":"EUR","trend":4.09,"average_7d":4.26,
  "provider_updated_at":"2026-10-29T22:54:33.169Z",
  "retrieved_at":"2026-10-30T15:40:00Z",
  "source_url":"https://api.tcgdex.net/v2/en/cards/sv01-201",
  "trace_id":"tr-example","trace_url":"",
  "note":"Market reference; language, finish and condition can change the value."
}
```

These numbers illustrate the schema; **never hardcode them or seed fake results**.
Research can return 409 if the item is not analyzed as Pokémon, 404 for an
unknown ID, or 502 when evidence cannot be verified. Errors are `{"error":"..."}`.
A successful response may have null prices. Show the model and tool traces using
only server-supplied HTTPS URLs. Do not construct MLflow links in the browser.

## Build, test and ship

Write the application source yourself from this contract. Do not clone, fetch,
copy or adapt an existing application or workshop reference implementation.
Keep the scaffold; do not add Python, MCP servers, databases or model clients.

Use `go:embed` for the UI. Prefer the Go standard library and a small source
surface. Run `go test -race ./...`, `go vet ./...`, and a local build. Meaningful
httptest cases must cover proxy status/errors, images and rejection of arbitrary
routes. Check the UI at mobile width and keyboard access when a browser is
available. Chromium must use `--use-mock-keychain`.

Follow AGENTS.md for Git → Tekton → Argo CD. Verify the exact pushed revision and
running image digest. Report actual model/browser checks separately from fixtures.
