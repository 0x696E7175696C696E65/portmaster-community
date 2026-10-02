# GLib 0.18.5 compatibility backport

This directory contains the authentic crates.io `glib` 0.18.5 source with its
MIT license and copyright retained. Its version is intentionally unchanged.

- Original archive: https://static.crates.io/crates/glib/glib-0.18.5.crate
- Archive SHA-256: `233daaf6e83ae6a12a52055f568f9d7cf4671dabb78ff9560ab6da230ce00ee5`
- Published source commit: `42b9caf98e03ded086362d9653ca58fe94dc8658`, directory `glib`
- Advisory: https://rustsec.org/advisories/RUSTSEC-2024-0429.html
- Reviewed upstream fix: https://github.com/gtk-rs/gtk-rs-core/pull/1343
- Fix commit: `b5a4071e439bef2b5eea76c3aa25e5ae84839e34`
- Upstream merge commit: `05dff0ee696f9bcd8617cd48c4b812d046d440cb`

The only upstream source changes are the two reviewed edits in
`src/variant_iter.rs`: declare the C output pointer mutable and pass `&mut p`
to `g_variant_get_child`. All other upstream files are preserved. The registry's
generated `.cargo-ok` extraction marker is omitted.

The main native manifest patches crates.io GLib to this local source while
preserving compatibility with GTK 0.18. Linux regression tests live in
`../../tests/glib-backport`, with a separate committed lockfile. From the
repository root, after installing `libglib2.0-dev` and Rust 1.98.1, run:

```sh
cargo +1.98.1 test --locked --release --manifest-path desktop/tauri/src-tauri/tests/glib-backport/Cargo.toml
```

These optimized tests exercise all five affected iterator methods, ordinary
strings, empty strings, Unicode, exhaustion, and mixed forward/backward access.
The standalone workspace avoids imposing Linux system libraries on Windows
application tests. The source is still 0.18.5, so version-based RustSec tools may
continue to report RUSTSEC-2024-0429. This backport does not hide that warning or
claim an upstream 0.20 release.

`cargo audit` skips this local path package and therefore stops emitting its
original GLib warning. That scanner behavior is not evidence of remediation;
the advisory, authentic archive checksum, reviewed source change and optimized
regressions above provide the separate verification for this backport.
