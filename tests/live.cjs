// Real integration, after deployment. No fixture responses or model credentials.
const assert = require('node:assert/strict');
const {expect} = require('@playwright/test');
const connectBrowser = require('./browser.cjs');
const base = process.env.APP_URL?.replace(/\/$/, '');
const upstream = (process.env.CARD_SCOUT_URL || 'http://card-scout-platform.card-scout-dev.svc.cluster.local:8080').replace(/\/$/, '');
assert.ok(base, 'Set APP_URL to the deployed HTTPS Route');
(async () => {
  const get = async url => { const r = await fetch(url); assert.equal(r.status, 200, url); return r.json(); };
  assert.deepEqual(await get(base + '/healthz'), {status: 'ok'});
  const [actual, expected] = await Promise.all([get(base + '/api/batch'), get(upstream + '/api/batch')]);
  assert.deepEqual(actual.dataset, expected.dataset, 'Use the real platform dataset');
  assert.deepEqual(actual.items.map(i => i.id).sort(), expected.items.map(i => i.id).sort());
  const browser = await connectBrowser();
  try {
    const page = await browser.newPage(); await page.goto(base);
    await expect(page.getByTestId('card')).toHaveCount(actual.dataset.items);
    if (actual.completed !== actual.dataset.items || actual.running || actual.failed) {
      await page.getByTestId('analyze').click();
      await expect.poll(async () => { const b = await get(base + '/api/batch'); return !b.running && !b.failed && b.completed === b.dataset.items; }, {timeout: 600000}).toBe(true);
    }
    const batch = await get(base + '/api/batch');
    const eligible = batch.items.filter(i => i.status === 'done' && i.kind === 'pokemon_tcg');
    const target = eligible.find(i => !i.research) || eligible[0]; assert.ok(target);
    const reply = page.waitForResponse(r => r.url() === base + '/api/cards/' + target.id + '/research' && r.request().method() === 'POST', {timeout: 240000});
    await page.locator('[data-testid=card][data-item-id=' + target.id + ']').getByRole('button', {name: 'Check price', exact: true}).click();
    const response = await reply; assert.equal(response.status(), 200); const data = await response.json();
    await expect(page.getByTestId('price')).toContainText(data.card.name);
    const trend = (await page.getByTestId('trend').innerText()).trim();
    if (data.trend === null) assert.equal(trend, 'Unavailable'); else assert.equal(Number(trend), data.trend);
    for (const [name, url] of [['Source', data.source_url], ['View MLflow trace', data.trace_url]]) {
      assert.ok(url?.startsWith('https://')); await expect(page.getByTestId('price').getByRole('link', {name, exact: true})).toHaveAttribute('href', url);
    }
    console.log(JSON.stringify({passed: true,scope: 'real platform and deployed app',completed: batch.completed,card: data.card.name,trace_url: data.trace_url}));
  } finally { await browser.close(); }
})().catch(e => {console.error(e.message); process.exitCode = 1;});
