# Signal Desk

A simple paper-investment game, in English.

1. **Set your fund memo**, starting with semiconductors and components.
2. **Load the repository’s dated Techmeme news** and read the relevance ranking /100.
3. **Find index ideas** for an interesting story: the agent compares archived evidence and explains the risk.
4. **Allocate $10,000 of virtual money**, then compare with the latest market price.

Choices freeze before market prices are fetched. Entry uses the first US session open after the fixed news cutoff; valuation uses the latest available regular-market quote, including intraday prices. Show its actual timestamp and disclose delays. Export the collected news and coverage gaps as JSON.

d3-mini ranks memo relevance; DeepSeek proposes hypothetical ETF exposures. A score is not confidence, BUY/SELL or predicted profit. Returns exclude dividends, fees and slippage. This is an educational replay, not a blind backtest. No orders are executed.

Read [APP_CONTRACT.md](APP_CONTRACT.md) for the app behavior and fixed model APIs.

Keep the prepared HTTPS Route, restricted-v2, storage and Git → Tekton → Argo CD contract in AGENTS.md. The working Go reference and local setup are in [the workshop repository](https://github.com/fjcloud/aro-ai-app-platform/tree/main/examples/briefing-desk).
