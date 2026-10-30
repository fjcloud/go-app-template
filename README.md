# Signal Desk

A simple paper-investment game, in English.

1. **Set your fund memo**, starting with semiconductors and components.
2. **Load the repository’s dated Techmeme news** and read the relevance ranking /100.
3. **Find company stock opportunities** for an interesting story: the agent compares archived evidence, looks up the direct company and possible suppliers/customers/competitors, and explains the catalyst and risk.
4. **Allocate $10,000 of virtual money**, then compare with the latest market price.

Choices freeze before market prices are fetched. Entry uses the first session open on each company’s exchange after the fixed news cutoff; valuation uses the latest available regular-market quote, including intraday prices. Show its actual timestamp and disclose delays. Export the collected news and coverage gaps as JSON.

d3-mini ranks memo relevance; DeepSeek proposes individual company shares with verified tickers and exchanges, plus sourced relationship hypotheses. ETFs and indices are excluded. Foreign shares include FX conversion into the USD paper fund. A score is not confidence, BUY/SELL or predicted profit. Returns exclude dividends, fees and slippage. This is an educational replay, not a blind backtest. No orders are executed.

Start with three news cards, choose a company and reveal the paper result. Full reasoning and market calculations stay in expandable sections. View MLflow trace opens the actual agent/tool graph. The bundled finance method is adapted from Anthropic Financial Services idea-generation.

Read [APP_CONTRACT.md](APP_CONTRACT.md) for the app behavior and fixed model APIs.

Keep the prepared HTTPS Route, restricted-v2, storage and Git → Tekton → Argo CD contract in AGENTS.md. The working Go reference and local setup are in [the workshop repository](https://github.com/fjcloud/aro-ai-app-platform/tree/main/examples/briefing-desk).
