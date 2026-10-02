const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('../desktop/angular/node_modules/typescript');
const root = path.resolve(__dirname, '..');
const source = fs.readFileSync(path.join(root, 'desktop/angular/src/app/shared/netquery/line-chart/chart-ticks.ts'), 'utf8');
const output = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText;
const sandbox = { exports: {}, Date };
vm.runInNewContext(output, sandbox);
const { chartTimeTicks, bandwidthTimeLabel } = sandbox.exports;

test('bandwidth labels use minutes instead of treating milliseconds as seconds', () => {
  const now = 1000000;
  assert.equal(bandwidthTimeLabel(new Date(now - 9 * 60000 - 14000), now), '9m ago');
  assert.equal(bandwidthTimeLabel(new Date(now - 14000), now), '<1m ago');
  assert.equal(bandwidthTimeLabel(new Date(now + 14000), now), '<1m ago');
});
test('ticks preserve time range while reserving label space at typical dashboard widths', () => {
  for (const width of [280, 480, 1024]) {
    const ticks = chartTimeTicks(new Date(0), new Date(600000), width);
    assert.equal(ticks[0].getTime(), 0);
    assert.equal(ticks[ticks.length - 1].getTime(), 600000);
    assert.ok(ticks.length <= 6);
    assert.ok(width / (ticks.length - 1) >= 110);
  }
});
test('help links all point to the fork and legacy ticket routes are redirected', () => {
  const pages = fs.readFileSync(path.join(root, 'desktop/angular/src/app/pages/support/pages.ts'), 'utf8');
  const routes = fs.readFileSync(path.join(root, 'desktop/angular/src/app/app-routing.module.ts'), 'utf8');
  assert.match(pages, /https:\/\/git\.frxst\.org\/bytefrxst\/portmaster/);
  assert.doesNotMatch(pages, /https?:\/\/(?:[^/]*safing\.|github\.com|twitter\.com|fosstodon\.org)/);
  assert.match(routes, /path: 'support\/:id',\s+redirectTo: 'support'/);
});
