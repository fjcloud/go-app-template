# Signal Desk

News ranked for an investment memo, starting with a semiconductor fund.

1. **Edit the memo** to describe the fund’s focus.
2. **Refresh news** to fetch and summarize Techmeme stories.
3. Read **opportunity scores /100** and their memo-specific explanations.
4. Open a story to investigate its implications or export the research note.

The UI is a memo and a ranked list. d3-mini scores research relevance/materiality;
positive and negative news can both rank highly. The score is neither BUY/SELL,
confidence nor an expected return. Changing the memo re-scores existing summaries
and hides old research.

The app uses OpenShift AI and read-only tools. Keep the prepared OAuth,
restricted-v2, storage and Git → Tekton → Argo CD deployment contract in AGENTS.md.
Detailed research can include sourced reference quotes and hypothetical scenarios;
no orders are executed.

The working Go reference, live local setup and labeled preview are in
[the workshop repository](https://github.com/fjcloud/aro-ai-app-platform/tree/feature/briefing-desk/examples/briefing-desk).
