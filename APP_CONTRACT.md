# Signal Desk — app contract

Build one simple English Go web app: **last week's news → memo relevance → index ideas → virtual allocation → this week's result**. Reuse the complete [Go reference](https://github.com/fjcloud/aro-ai-app-platform/tree/feature/briefing-desk/examples/briefing-desk), including its tests and web assets. AGENTS.md supplies delivery rules; preserve the prepared scaffold.

## User flow

1. Edit a fund memo (default: semiconductors and electronic components).
2. Load the previous Monday–Sunday from Techmeme archives, rank every collected primary story and show the imported/ranked counts. Export all collected news and coverage gaps as JSON. Initially show 30 stories with Show more.
   If cloud requests are blocked, offer **Import news** for a same-week Signal Desk JSON export. Validate dates/source URLs and label manual provenance. Ignore uploaded scores, research and portfolio data; score archived facts using the server model.
3. Open a story and find up to three hypothetical long-only ETF ideas (QQQ, SPY, SMH or XLK), with a short thesis, risk and actual source evidence. No supported exposure is a valid outcome.
4. Allocate a virtual **$10,000** across selected ideas. No result or future price is shown yet.
5. Reveal: persist and freeze the entire portfolio before requesting historical prices. Entry is the first available US session open after the news week; exit is the latest completed US session close this week. Show dates, source prices, position P&L and total fund value including uninvested cash.
6. Label unfinished weeks **Week in progress**. Refresh valuation without altering choices. Play again creates a fresh budget and retains previous rounds.

Keep one column: collapsible memo, portfolio, ranked news. Show loading/progress and errors; preserve unsaved memo edits during polling. No ticker dashboard, microphone or real trades.

## Fixed model APIs — backend only

The prepared ConfigMap supplies all endpoint/model values. Never invent a URL or put credentials in the browser.

| Purpose | Endpoint and model | Contract |
| --- | --- | --- |
| Relevance | `DECISION_URL=http://d3-mini.decision-serving.svc.cluster.local:8080/v1/systemone`, `DECISION_MODEL=vllm-sr/d3-mini` | Native System One JSON. `state` contains memo and archived source facts; question `opportunity` has `type:score`, five ordered criteria 0–4, `require_full_input:true`. Validate complete coverage, the five probabilities and expected score; display `score / 4 × 100`. No generated explanation replaces this result. |
| Research agent | `AGENT_URL=http://lsd-genai-playground-service.live-transcript-dev.svc.cluster.local:8321/v1`, `AGENT_MODEL=maas-vllm-inference-1/deepseekv4pro` | Native OpenShift AI `POST /responses`, `store:false`, bounded function-tool loop. Read `AGENT_KEY_FILE` and send `X-OGX-Provider-Data: {"vllm_api_token":"<runtime MaaS key>"}`. Use the reference client; no direct Azure call or provider key. |

The developer preflight uses `https://d3-api.decision-serving.svc.cluster.local:8443/v1/systemone` with a scoped OpenShift Bearer token. That endpoint adds authorization to the same native protocol; it is not MaaS. The runtime table above intentionally uses the existing internal app path. Preserve the platform ConfigMap; never hardcode a developer token or send the MaaS key to D3.

D3 is a community decision model, not a RedHatAI or supported Red Hat AI Inference model/runtime. Its score is **relevance/materiality**, not confidence, bullishness, probability of profit or expected return. Both positive and negative catalysts can score highly. Never replace failed inference with a keyword rule.

## Evidence and replay rules

- Archive week/cutoff uses `America/New_York`. Use archived Techmeme snapshots and backfill missing primary post IDs; deduplicate publisher URLs. Disclose unavailable IDs and excerpt-only coverage. Do not fetch today's publisher article to enrich past facts.
- Game research tools: `read_archive(source_id)`, `search_archive(query)`, `index_universe()`. Resolve source IDs server-side, validate citations and cap calls/turns. No current-web, arbitrary URL, shell, price, return or order tool is available to this agent.
- Historical Yahoo daily prices are fetched by deterministic Go code after commitment. Validate symbol/currency, sessions, ordering and positive prices; exclude current incomplete/future bars. Report missing prices or unsupported split accounting instead of inventing data. Retrying a valuation cannot change the frozen portfolio.
- Gross price-only P&L: `amount × (exit / entry − 1)`; fund value is `$10,000 + sum(P&L)`. Exclude dividends, fees and slippage explicitly. No price targets or calibrated return forecasts.
- This is an educational historical replay. A tool cutoff does not erase model training knowledge: never call it a blind backtest.
- Cache scores/ideas by memo version and source-summary hash. Memo edits re-score saved news and invalidate old ideas. Persist documents, snapshots and rounds atomically under `DATA_DIR`.

## Small app API

Reuse the reference JSON contracts:

| Route | Purpose |
| --- | --- |
| `GET /healthz` | `{"status":"ok"}` |
| `GET /api/game` | Week, memo, news, ranking progress, ideas and current round |
| `POST /api/game/load` | Start/resume archive import and scoring |
| `POST /api/game/import` | Import a bounded same-week JSON news export and rank its archived facts |
| `GET /api/game/export` | Full collected corpus and disclosed coverage gaps |
| `POST /api/profiles` | Save fund memo and re-score |
| `POST /api/game/ideas` | Research a scored `article_id` |
| `POST /api/game/play` | Validate `round_id`, `memo_version`, positions with `article_id`, `idea_id`, `amount`; freeze before prices |
| `POST /api/game/settle` | Refresh frozen valuation |
| `POST /api/game/reset` | Retain prior round, start again |
| `/mcp` | Authenticated read-only Streamable HTTP: `list_news`, `read_archive`, `search_archive`, `index_universe`. No prices, positions, mutation or credentials. |

## Platform and acceptance

Keep technical APP_NAME `live-transcript`, the `signal-desk` ServiceAccount/ConfigMap, mounted model key, data PVC, OpenShift OAuth proxy, loopback listener and service CA. The public app requires OpenShift login. Traces go to the prepared MLflow `signal-desk` experiment; include model/tool timing and usage, never keys, full news text or portfolio contents. Show the research trace ID.

Run reference Go race/vet tests. Verify archived-week scope, memo re-scoring, real tool evidence, budget checks, frozen choices before price fetch, completed-session valuation and correct fund arithmetic. Test errors and missing data. Verify the English page on mobile. Distinguish fixture tests from actual model/cluster tests. Follow AGENTS.md for Git → Tekton → Argo CD and verify the matching pushed revision, not just an older healthy app.
