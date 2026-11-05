// Acceptance criteria for the generated app. These fixtures never enter production.
const {test, before, after} = require('node:test');
const assert = require('node:assert/strict');
const http = require('node:http');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const {spawn, spawnSync} = require('node:child_process');
const {expect} = require('@playwright/test');
const connectBrowser = require('./browser.cjs');
const image = Buffer.from('<svg xmlns="http://www.w3.org/2000/svg" width="200" height="300"><rect width="200" height="300" fill="gold"/></svg><!--' + 'x'.repeat(2 * 1024 * 1024) + '-->');
const price = {status: 'ok', card: {name: 'Charizard'}, trend: 0,
  provider_updated_at: '2026-11-01T12:00:00Z', source_url: 'https://example.org/card',
  trace_url: 'https://example.org/trace', note: 'Catalogue aggregate, not a verified 1st Edition quote.'};
const cards = [
  {id: 'item-104', name: 'Charizard', kind: 'pokemon_tcg', rarity: 'rare_or_higher'},
  {id: 'item-197', name: 'Energy', kind: 'pokemon_tcg', rarity: 'common'},
  {id: 'item-203', name: 'Pikachu artwork', kind: 'not_a_card', rarity: 'not_applicable'}
].map(c => ({...c, image_url: '/assets/' + c.id, status: 'done', research: null}));
let fixture, app, browser, base, tmp, state, logs = '';
function reset(options = {}) {
  state = {reads: 0, starts: 0, analyzed: false, polls: 0, batchFailures: 0,
    priceFailures: 0, quote: structuredClone(price), ...options};
}
function reply(res, code, data) { res.writeHead(code, {'Content-Type': 'application/json'}); res.end(JSON.stringify(data)); }
function upstream(req, res) {
  if (req.url === '/api/batch') {
    state.reads++;
    if (state.batchFailures-- > 0) return reply(res, 503, {error: 'Fixture unavailable'});
    const running = state.analyzed && state.polls++ === 0;
    return reply(res, 200, {dataset: {name: 'Acceptance fixture', items: 3}, running,
      completed: state.analyzed ? (running ? 1 : 3) : 0, failed: 0,
      items: cards.map(c => state.analyzed ? c : {...c, status: 'pending', kind: null, rarity: null})});
  }
  if (req.method === 'POST' && req.url === '/api/batch/analyze') {
    state.starts++; state.analyzed = true; return reply(res, 202, {running: true});
  }
  if (req.url === '/assets/item-104' || req.url === '/assets/item-197' || req.url === '/assets/item-203') {
    res.writeHead(200, {'Content-Type': 'image/svg+xml'}); return res.end(image);
  }
  if (req.method === 'POST' && req.url === '/api/cards/item-104/research') {
    if (state.priceFailures-- > 0) return reply(res, 502, {error: 'Fixture price unavailable'});
    return reply(res, 200, state.quote);
  }
  reply(res, 404, {error: 'Not found'});
}
async function listen(server) {
  await new Promise(r => server.listen(0, '127.0.0.1', r)); return server.address().port;
}
before(async () => {
  for (const args of [['test', '-race', './...'], ['vet', './...']]) {
    const r = spawnSync('go', args, {encoding: 'utf8', env: {...process.env, CGO_ENABLED: '1'}});
    assert.equal(r.status, 0, r.stdout + r.stderr);
  }
  tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'card-scout-'));
  const binary = path.join(tmp, 'app');
  const built = spawnSync('go', ['build', '-buildvcs=false', '-o', binary, '.'], {encoding: 'utf8'});
  assert.equal(built.status, 0, built.stdout + built.stderr);
  reset(); fixture = http.createServer(upstream);
  const fixturePort = await listen(fixture), reservation = http.createServer();
  const appPort = await listen(reservation); await new Promise(r => reservation.close(r));
  base = 'http://127.0.0.1:' + appPort;
  app = spawn(binary, [], {env: {PATH: process.env.PATH, HOME: process.env.HOME,
    CARD_SCOUT_URL: 'http://127.0.0.1:' + fixturePort, LISTEN_ADDR: '127.0.0.1:' + appPort}, stdio: ['ignore', 'pipe', 'pipe']});
  app.stdout.on('data', b => logs += b); app.stderr.on('data', b => logs += b);
  await expect.poll(async () => { try { return (await fetch(base + '/healthz')).status; } catch { return 0; } }, {timeout: 10000, message: 'The actual Go app must start: ' + logs}).toBe(200);
  browser = await connectBrowser();
});
after(async () => {
  if (browser) await browser.close();
  if (app) { const exited = new Promise(r => app.once('exit', r)); app.kill(); await exited; }
  if (fixture) await new Promise(r => fixture.close(r));
  if (tmp) fs.rmSync(tmp, {recursive: true, force: true});
});
async function pageFor(t) {
  const context = await browser.newContext(); t.after(() => context.close());
  return context.newPage();
}
test('real Go server: health, fixed routes, complete images and upstream errors', async () => {
  reset(); assert.deepEqual(await (await fetch(base + '/healthz')).json(), {status: 'ok'});
  for (const route of ['/unknown', '/proxy?url=https://example.org', '/assets/item-bad']) {
    assert.equal((await fetch(base + route)).status, 404, 'Reject ' + route);
  }
  const r = await fetch(base + '/assets/item-104');
  assert.equal(r.headers.get('content-type'), 'image/svg+xml');
  assert.deepEqual(Buffer.from(await r.arrayBuffer()), image, 'Do not truncate large card images');
  state.priceFailures = 1;
  const error = await fetch(base + '/api/cards/item-104/research', {method: 'POST'});
  assert.equal(error.status, 502); assert.deepEqual(await error.json(), {error: 'Fixture price unavailable'});
});
test('load → analyze → price, reload and full cards on mobile', async t => {
  reset(); const p = await pageFor(t); await p.goto(base);
  await expect(p.getByRole('heading', {name: 'Card Scout', exact: true})).toBeVisible();
  await expect(p.getByTestId('card')).toHaveCount(3);
  await expect(p.getByTestId('progress')).toContainText('0/3');
  await p.getByTestId('analyze').click(); await expect(p.getByTestId('analyze')).toBeDisabled();
  await expect(p.getByTestId('progress')).toContainText('3/3', {timeout: 10000});
  assert.equal(state.starts, 1);
  const control = p.locator('[data-testid=card][data-item-id=item-203]');
  await expect(control.getByRole('button', {name: 'Check price', exact: true, disabled: false})).toHaveCount(0);
  const card = p.locator('[data-testid=card][data-item-id=item-104]');
  await expect(card.getByTestId('rarity')).toHaveText('Rare');
  await card.getByRole('button', {name: 'Check price', exact: true}).click();
  const panel = p.getByTestId('price'); await expect(panel).toContainText('Charizard');
  await expect(panel.getByTestId('trend')).toHaveText(/^0(?:\.0+)?$/);
  await expect(panel).toContainText('2026-11-01'); await expect(panel).toContainText(price.note);
  await expect(panel.getByRole('link', {name: 'Source', exact: true})).toHaveAttribute('href', price.source_url);
  await expect(panel.getByRole('link', {name: 'View MLflow trace', exact: true})).toHaveAttribute('href', price.trace_url);
  await p.reload(); await expect(p.getByTestId('progress')).toContainText('3/3');
  await p.setViewportSize({width: 390, height: 844});
  const img = p.getByTestId('card-image').first();
  await expect.poll(() => img.evaluate(i => i.complete && i.naturalWidth > 0)).toBe(true);
  const g = await img.evaluate(i => ({fit: getComputedStyle(i).objectFit, width: i.width,
    ratio: i.width / i.height, natural: i.naturalWidth / i.naturalHeight}));
  assert.ok(g.width > 100 && g.fit !== 'cover' && (g.fit === 'contain' || Math.abs(g.ratio / g.natural - 1) < .03), 'Show the whole card');
  assert.ok(await p.evaluate(() => document.documentElement.scrollWidth <= innerWidth), 'No mobile overflow');
});
test('two initial load failures can be retried without losing the page', async t => {
  reset({batchFailures: 2}); const p = await pageFor(t); await p.goto(base);
  await expect(p.getByRole('alert')).toBeVisible();
  await p.getByRole('button', {name: 'Retry', exact: true}).click();
  await expect.poll(() => state.reads).toBe(2); await expect(p.getByRole('alert')).toBeVisible();
  await p.getByRole('button', {name: 'Retry', exact: true}).click();
  await expect(p.getByTestId('card')).toHaveCount(3); await expect(p.getByRole('alert')).toBeHidden();
});
test('price failure keeps the card; Retry handles unavailable prices and unsafe links', async t => {
  reset({analyzed: true, polls: 1, priceFailures: 1,
    quote: {...price, trend: null, source_url: 'javascript:alert(1)', trace_url: 'data:text/html,test'}});
  const p = await pageFor(t); await p.goto(base);
  await p.locator('[data-testid=card][data-item-id=item-104]').getByRole('button', {name: 'Check price', exact: true}).click();
  await expect(p.getByRole('alert')).toBeVisible(); await expect(p.getByTestId('price')).toContainText('Charizard');
  await p.getByRole('button', {name: 'Retry', exact: true}).click();
  await expect(p.getByTestId('trend')).toHaveText('Unavailable');
  await expect(p.getByTestId('price').locator('a')).toHaveCount(0);
});
