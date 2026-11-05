# Card Scout

Build the minimal app from scratch using APP_CONTRACT.md. Go standard library,
embedded HTML/CSS/JavaScript. No existing app source, framework or model client.

- Only edit application files. Keep the prepared tests, package manifests,
  contract, devfile, deploy/, pipeline/ and gitops/ unchanged.
- Prepare the test tools once: `npm ci && npm run setup`.
- Run `npm test`. Fix the application and repeat until every test passes.
  Do not weaken tests or substitute fake data in the application.
- When asked to ship, follow DELIVERY.md through Git, Tekton and Argo CD.
  Then run `APP_URL=https://your-route npm run test:live` and fix any app failures.
- Finish only with passing tests and the exact deployed Git revision/image digest.
  Report the URL and actual results. If platform access or tooling is unavailable,
  report the blocker instead of claiming success.
- Never print credentials, grant permissions/SCCs or use an administrator login.
  The app only calls CARD_SCOUT_URL; it needs no model key or Kubernetes token.
