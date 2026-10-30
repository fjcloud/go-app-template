# Signal Desk — app contract

Build one simple English Go web app: **repository news snapshot → memo relevance → company stock opportunities → virtual allocation → latest market result**. Reuse the complete [Go reference](https://github.com/fjcloud/aro-ai-app-platform/tree/main/examples/briefing-desk), including its tests and web assets. AGENTS.md supplies delivery rules; preserve the prepared scaffold.

## User flow

1. Edit a fund memo (default: semiconductors and electronic components).
2. Load `news/techmeme.json`, bundled from the Go reference with `go:embed`. It contains 161 dated Techmeme headlines/excerpts from 26–28 October 2026, captured on 29 October. Rank every saved article and show loaded/ranked counts. Export the corpus and provenance as JSON. Initially show 30 stories with Show more. The normal flow must not contact Techmeme or require manual import. Preserve the recorded information cutoff and source links.
3. Open a story and research up to three individual company stocks: the named business and, where supported, its suppliers, customers or competitors. Verify each ticker, company name and exchange with `search_companies`; ETFs, funds and indices are excluded. Show the relationship, specific catalyst, archived evidence and concrete risk. Related stocks do not automatically move together. An empty idea list is valid when evidence is insufficient.
4. Allocate a virtual **$10,000** across selected ideas. No result or future price is shown yet.
5. Reveal: persist and freeze the entire portfolio before requesting historical prices. Entry is each stock exchange’s first session open after the fixed snapshot cutoff (29 October 2026 at 00:05 America/New_York); valuation is the latest available regular-market quote, including intraday or delayed quotes. Keep the news fixed across calendar weeks. Show dates, source prices, position P&L and total fund value including uninvested cash.
6. Show the actual quote timestamp and **Latest available market quote**. Update result fetches a fresh quote without altering choices. Play again creates a fresh budget and retains previous rounds.

Keep one column: collapsible memo, portfolio, ranked news. Show loading/progress and errors; preserve unsaved memo edits during polling. No ticker dashboard, microphone or real trades.

## Fixed model APIs — backend only

The prepared ConfigMap supplies all endpoint/model values. Never invent a URL or put credentials in the browser.

| Purpose | Endpoint and model | Contract |
| --- | --- | --- |
| Relevance | `DECISION_URL=http://d3-mini.decision-serving.svc.cluster.local:8080/v1/systemone`, `DECISION_MODEL=vllm-sr/d3-mini` | Native System One JSON. `state` contains memo and archived source facts; question `opportunity` has `type:score`, five ordered criteria 0–4, `require_full_input:true`. Validate complete coverage, the five probabilities and expected score; display `score / 4 × 100`. No generated explanation replaces this result. |
| Research agent | `AGENT_URL=http://lsd-genai-playground-service.signal-desk-dev.svc.cluster.local:8321/v1`, `AGENT_MODEL=maas-vllm-inference-1/deepseekv4pro` | Native OpenShift AI `POST /responses`, `store:false`, bounded function-tool loop. Read `AGENT_KEY_FILE` and send `X-OGX-Provider-Data: {"vllm_api_token":"<runtime MaaS key>"}`. Use the reference client; no direct Azure call or provider key. |

The developer and runtime both call the internal D3 service directly, with no API key or authentication proxy. DeepSeek uses its separate MaaS credential. Preserve the model endpoints supplied by the platform ConfigMap.

D3 is a decision model from vLLM Semantic Router. Its score is **relevance/materiality**, not confidence, bullishness, probability of profit or expected return. Both positive and negative catalysts can score highly. Never replace failed inference with a keyword rule.

## Evidence and replay rules

- Use the embedded, dated repository snapshot with its `America/New_York` cutoff. Do not roll it to another week or fetch publisher pages. Retain collection provenance, deduplicated publisher links and excerpt-only coverage. Copy `news/` with the Go source; never invent a replacement dataset.
- Game research tools: `read_archive(source_id)`, `search_archive(query)`, `search_companies(query)`. Look up the direct business and another possible related company with at least two distinct company-name queries. Resolve source IDs server-side; accept only returned individual equities, bind the final schema to those symbols and require archived source citations for every idea. Company lookup returns current identity metadata only, without prices or historical fundamentals. Keep the six-tool budget and bounded turns. No live publisher pages, arbitrary URL, shell, price, return or order tool is available to this agent.
- Yahoo entry-session opens and latest regular-market quotes are fetched by deterministic Go code after commitment. Validate symbol/currency, ordered sessions, positive prices and quote timestamps. Entry must follow the news cutoff. Never substitute an incomplete daily close for the latest quote, or invent a future quote. Report missing prices or unsupported split accounting instead of inventing data. Retrying a valuation cannot change the frozen portfolio.
- Keep the paper budget and P&L in USD. For foreign shares, show local prices/currency, the stock exchange timezone and separate FX source/timestamp. Use the last daily opening FX reference available by stock entry, then the latest FX quote at valuation. USD P&L is `amount × ((exit_local × exit_USD_per_unit) / (entry_local × entry_USD_per_unit) − 1)`; fund value is `$10,000 + sum(P&L)`. Normalize pence quotes to their base currency. Missing or invalid FX prevents valuation. Exclude dividends, fees and slippage; no price targets or calibrated return forecasts.
- This is an educational historical replay. A tool cutoff does not erase model training knowledge: never call it a blind backtest.
- Cache scores/ideas by memo version and source-summary hash. Memo edits re-score saved news and invalidate old ideas. Persist documents, snapshots and rounds atomically under `DATA_DIR`. Version the equities game state so earlier ETF proposals and portfolios cannot be reused as company-share ideas; retain the earlier files.

## Small app API

Reuse the reference JSON contracts:

| Route | Purpose |
| --- | --- |
| `GET /healthz` | `{"status":"ok"}` |
| `GET /api/game` | Week, memo, news, ranking progress, ideas and current round |
| `POST /api/game/load` | Load the bundled repository news and start/resume scoring |
| `POST /api/game/import` | Import a bounded same-week JSON news export and rank its archived facts |
| `GET /api/game/export` | Full collected corpus and disclosed coverage gaps |
| `POST /api/profiles` | Save fund memo and re-score |
| `POST /api/game/ideas` | Research a scored `article_id` |
| `POST /api/game/play` | Validate `round_id`, `memo_version`, positions with `article_id`, `idea_id`, `amount`; freeze before prices |
| `POST /api/game/settle` | Refresh frozen valuation |
| `POST /api/game/reset` | Retain prior round, start again |
| `/mcp` | Public read-only Streamable HTTP: `list_news`, `read_archive`, `search_archive`, `search_companies`. No prices, positions, mutation or credentials. |

## Platform and acceptance

Keep technical APP_NAME `signal-desk`, the `signal-desk` ServiceAccount/ConfigMap, mounted model key, data PVC and service CA. Expose the app on `:8080` through its Service and HTTPS edge Route; the app and MCP connector require no login or token. All participants share the same workshop state. Traces go to the prepared MLflow `signal-desk` experiment; include model/tool timing and usage, never keys, full news text or portfolio contents. Show the research trace ID.

Run reference Go race/vet tests. Verify archived-week scope, memo re-scoring, real tool evidence, budget checks, frozen choices before price fetch, timestamped latest-quote valuation and correct fund arithmetic. Test errors and missing data. Verify the English page on mobile. Distinguish fixture tests from actual model/cluster tests. Follow AGENTS.md for Git → Tekton → Argo CD and verify the matching pushed revision, not just an older healthy app.
