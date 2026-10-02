const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');
const Fuse = require('fuse.js');

test('matched overview searches preserve raw profile labels for text bindings', () => {
  const serviceSource = fs.readFileSync(path.join(__dirname, '../src/app/shared/fuzzySearch/fuse.service.ts'), 'utf8');
  const sandbox = {
    exports: {},
    require(module) {
      if (module === '@angular/core') return { Injectable: () => target => target };
      if (module === '@safing/portmaster-api') return { deepClone: structuredClone };
      if (module === 'fuse.js') return Fuse;
      throw new Error(`Unexpected service dependency ${module}`);
    },
  };
  vm.runInNewContext(ts.transpileModule(serviceSource, { compilerOptions: { module: ts.ModuleKind.CommonJS, esModuleInterop: true, experimentalDecorators: true } }).outputText, sandbox);
  const overview = fs.readFileSync(path.join(__dirname, '../src/app/pages/app-view/overview.ts'), 'utf8');
  const optionsSource = overview.match(/\.searchList\(profiles,\s*searchTerm,\s*({[\s\S]*?})\)/)?.[1];
  assert.ok(optionsSource, 'overview search options are present');
  const options = vm.runInNewContext(`(${optionsSource})`);
  const profile = { Name: 'Zen <img src=x onerror=alert(1)> browser', PresentationPath: 'C:/Apps/Zen/browser.exe' };
  const service = new sandbox.exports.FuzzySearchService();
  const results = service.searchList([profile], 'browser', options);
  assert.equal(results.length, 1);
  assert.ok(results[0].matches.length > 0, 'fixture matches the search term');
  assert.equal(results[0].item.Name, profile.Name);
  assert.equal(results[0].item.PresentationPath, profile.PresentationPath);
});
