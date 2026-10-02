# Community native WebSocket boundary

Vendored from the crates.io release `tauri-plugin-websocket` **2.3.0**.
Upstream repository: https://github.com/tauri-apps/plugins-workspace
Upstream commit recorded in the published crate: `ad1726627365e44c1841129be1871fb9509f5eb3`, path `plugins/websocket`.
Unmodified `src/lib.rs` SHA-256: `2096a8955a63f7fc837bd97e3a5873e62ea85f18ecdfdcd65521a7f1a2e2b29a`.

The upstream Apache-2.0 and MIT licenses and copyright notices are retained.

The community patch permits only `ws://127.0.0.1:817/api/database/v1`.
Legacy `localhost` callers are canonicalized to the literal loopback IP before connecting.
It refuses user credentials, fragments, other ports, schemes and paths, and caller-supplied headers.
Incoming frames and messages and the outgoing buffer are capped at 8 MiB; caller configuration cannot disable the caps.
At most eight sessions may connect concurrently, and connection/write operations expire after ten seconds. EOF and errors release connection resources.
TLS features are disabled by the application because this bridge only needs the local plain WebSocket.

Run `cargo test --workspace --locked` from the application directory to exercise the application and these boundary tests.
When upgrading this vendored code, reapply and review these checks and retain the provenance and licenses.
