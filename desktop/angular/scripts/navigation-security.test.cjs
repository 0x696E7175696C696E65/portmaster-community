const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');
const source = fs.readFileSync(path.join(__dirname, '../src/app/integration/navigation.ts'), 'utf8');
const sandbox = { exports: {}, URL };
vm.runInNewContext(ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText, sandbox);
const classify = href => sandbox.exports.navigationTarget(href, 'http://127.0.0.1:817/ui/');

test('native opener rejects executable, file and unknown URI schemes', () => {
  for (const href of ['javascript:alert(1)', ' JAVASCRIPT:alert(1)', 'data:text/html,test', 'file:///C:/Windows/system32/cmd.exe', 'ms-settings:display', 'mailto:test@example.com', '//user:secret@evil.example/', 'http://[invalid']) {
    assert.equal(classify(href).kind, 'blocked', href);
  }
});
test('same-origin links and local blob downloads stay in the application', () => {
  for (const href of ['/ui/#/settings', '#help', './support', 'http://127.0.0.1:817/ui/', 'blob:http://127.0.0.1:817/123']) {
    assert.equal(classify(href).kind, 'internal', href);
  }
  assert.equal(classify('blob:https://evil.example/123').kind, 'blocked');
});
test('origin comparison does not trust host substrings or different ports', () => {
  for (const href of ['https://evil.example/?host=127.0.0.1', 'http://127.0.0.1.evil.example/', 'http://127.0.0.1:818/', '//github.com/0x696E7175696C696E65/portmaster-community']) {
    const result = classify(href);
    assert.equal(result.kind, 'external', href);
    assert.match(result.url, /^https?:\/\//);
  }
});
