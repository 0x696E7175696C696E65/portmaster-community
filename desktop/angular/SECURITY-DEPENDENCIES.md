# Frontend dependency maintenance

The community frontend uses Angular 21 and compatible CDK, ng-zorro, ngx-markdown,
Font Awesome and build tools. Angular 16 is unsupported and has unpatched
sanitizer advisories. Keep all framework packages on the same patched version.
Use a Node version listed in `package.json` and the
[Angular compatibility table](https://angular.dev/reference/versions).

The 2026-10-02 audit reduced npm's full dependency findings from 90 (including 4
critical and 45 high) to zero. The production dependency tree went from 23
(including 10 high) to zero. These numbers describe published package advisories,
not a guarantee that every application behavior is secure. Re-run audits whenever
the lockfile changes and before releasing.

```powershell
npm ci --ignore-scripts
npm audit --audit-level=low
npm audit --omit=dev --audit-level=low
npm run test:security
npm run lint:security
npm run build-libs:dev
ng build --configuration production --base-href /ui/modules/portmaster/
ng build --configuration production tauri-builtin
```

Use `node_modules/.bin/ng.cmd` for the Angular commands on Windows. Build the main
frontend before `tauri-builtin`: their outputs share `dist`, and the main builder
cleans that directory. Native builds require the built Tauri frontend.

Two transitive overrides are deliberate and covered by regression tests:

- `piscina` 5.3.2 fixes
  [inherited execution options](https://github.com/piscinajs/piscina/security/advisories/GHSA-67c8-pqhq-4rmx)
  in the worker pools used by the Angular builders and ng-packagr. This keeps their
  existing major version. Remove the override after all parent packages require a
  patched version.
- SockJS uses `uuid.v4()` through CommonJS. `uuid` 11.1.1 preserves that interface
  while fixing [buffer bounds validation](https://github.com/uuidjs/uuid/security/advisories/GHSA-w5hq-g745-h8pq).
  A local websocket session/echo test verifies compatibility. Remove this override
  when SockJS requires a patched dependency itself.

The obsolete Protractor scaffold only checked the generated "app is running"
message and depended on unmaintained vulnerable packages; its files, command and
builder were removed. `test:security` is a regression suite, not an end-to-end
suite. Use the repository release checklist for application smoke tests. The
browser extension build target remains available; development updates require a
manual reload in the browser after removing its unmaintained hot-reloader.

The migration retains the existing NgModule architecture (`standalone: false`)
and zone-based change detection explicitly. Markdown remains sanitized by Angular;
do not introduce sanitizer bypasses for API responses, notifications or help text.
`lint:security` rejects calls to Angular's sanitizer bypass API across application
and library TypeScript sources. It runs independently of pre-existing general
lint issues and cannot be disabled by inline ESLint comments.
