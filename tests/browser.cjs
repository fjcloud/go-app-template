const {chromium} = require('@playwright/test');
module.exports = async function browser() {
  if (process.env.PLAYWRIGHT_WS_ENDPOINT) {
    return chromium.connect(process.env.PLAYWRIGHT_WS_ENDPOINT, {timeout: 10000});
  }
  return chromium.launch({headless: true, args: ['--use-mock-keychain']});
};
