# Signal Desk

An English news intelligence desk for a hedge-fund workshop.

1. Select a fund mandate and **Scan news**.
2. Read English summaries and financial relevance from the local T4 model.
3. Select **Investigate** on a relevant story. Follow the OpenShift AI agent's
   tool activity and inspect the source-backed research draft.
4. Export a summary, research note or financial briefing as Markdown.

The facilitator prepares the existing agent, local model, runtime key and storage
with `prepare-briefing-desk.py` from the workshop repository before deployment.
The MaaS key lasts 24 hours; rerun preparation and restart the app to renew it.

The app uses native OpenShift AI Responses and internal read-only MCP tools,
public BBC/France 24 RSS, SEC submission metadata and ECB reference rates.
Current-feed search is bounded; it is not exhaustive web research. Source facts,
financial hypotheses and missing evidence remain separate. No trades are executed.

See AGENTS.md for the application, security and delivery contract.
