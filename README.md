# Signal Desk

A simple paper-investment game, in English.

1. **Set your fund memo**, starting with semiconductors and components.
2. **Load last week's Techmeme news** and read the relevance ranking /100.
3. **Find index ideas** for an interesting story: the agent compares archived evidence and explains the risk.
4. **Allocate $10,000 of virtual money**, then reveal this week's actual result.

Choices freeze before market prices are fetched. Entry uses the first US session open after the news week; valuation uses the latest completed session close. A running week is clearly marked provisional. Export the collected news and coverage gaps as JSON.

d3-mini ranks memo relevance; DeepSeek proposes hypothetical ETF exposures. A score is not confidence, BUY/SELL or predicted profit. Returns exclude dividends, fees and slippage. This is an educational replay, not a blind backtest. No orders are executed.

Read [APP_CONTRACT.md](APP_CONTRACT.md) for the app behavior and fixed model APIs.

Keep the prepared OAuth, restricted-v2, storage and Git → Tekton → Argo CD contract in AGENTS.md. The working Go reference and local setup are in [the workshop repository](https://github.com/fjcloud/aro-ai-app-platform/tree/main/examples/briefing-desk).
