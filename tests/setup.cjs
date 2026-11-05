const {spawnSync} = require('node:child_process');
const browser = require('./browser.cjs');
(async () => {
  if (!process.env.PLAYWRIGHT_WS_ENDPOINT) {
    const r = spawnSync('npx', ['--no-install', 'playwright', 'install', 'chromium'], {stdio: 'inherit'});
    if (r.status !== 0) process.exit(r.status || 1);
  }
  let last;
  for (let i = 0; i < 60; i++) {
    try { const b = await browser(); await b.close(); console.log('Test browser ready.'); return; }
    catch (e) { last = e; await new Promise(r => setTimeout(r, 1000)); }
  }
  throw last;
})().catch(e => { console.error(e.message); process.exitCode = 1; });
