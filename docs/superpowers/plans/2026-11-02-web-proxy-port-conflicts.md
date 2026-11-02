# Web proxy port conflicts

**Goal:** `web-proxy on` detects an unavailable port, tells the user, and automatically enables a usable listener.

**Design:** Keep selection in the invoking CLI, which owns the current user's system proxy. Reuse the configured address when `--listen` is omitted. An existing healthy listener must belong to this client. Otherwise check binding on the requested literal loopback IP; if unavailable, ask the OS for a free port on that same IP. Persist the selected address using existing backups. Keep the current identity and relay checks before changing system settings. Leave other applications running.

**Boundary:** Foreground `web-proxy run` still binds the explicitly requested address. This change does not add a service, change device configuration, expand relay domains, or change TLS verification. Choosing a port cannot reserve it across the CLI/service boundary, so activation rechecks and retries a competing bind within a bounded context.

## Implementation

- [x] Add isolated regression tests using occupied loopback listeners, another client's health endpoint, a healthy same-client listener, invalid addresses, cancellation, and a simulated bind race. No real system proxy or credentials in tests.
- [x] Add address selection to `internal/agent`, retaining exact binding behavior of `ListenWebProxy`.
- [x] Wire selection, warning output, and bounded retry into `cmd/agent/web_proxy.go`; preserve the configured listener when no `--listen` is supplied.
- [x] Update `docs/web-proxy.md` and the embedded remote-agent skill with automatic selection and explicit proxy lookup instructions.
- [x] Run focused tests, full Go tests/vet, formatting, and six-platform builds; independent code review found no actionable issues.
- [x] Review only intended changes; commit and push `main` (`d1c509f`).
- [x] Upgrade this client's signed core through the existing updater; verify a conflict with the real occupied port, a repeated plain `on`, and system-proxy HTTP access. Preserve the installed service, client identity, project/session data, and unrelated repository edits.

## 2026-11-02 acceptance

Six platform binaries and signed update manifests were built from the isolated source snapshot containing only this change, and published as `20261102-port-auto`, sequence `2026110201`. The 139 server program was not restarted; its previous download directory is retained at `/var/backups/remote-agent/20261102-port-auto-dist`.

The installed Windows client updated through the existing signed core updater with no force option. Both existing automatic services remained installed and running; the device configuration SHA-256 was unchanged. After `off`, plain `on` detected the existing `0dcloudCore.exe` listener on `17891`, printed the conflict, selected `127.0.0.1:53920`, and enabled the current user's system proxy. The other application's listener remained running. Repeated plain `on` reused `53920`.

Windows WinINET PRECONFIG requests, without a per-request proxy, returned HTTP 200 for Google Play, GitHub, and Baidu. The installed remote-agent skill was updated using the existing installer with its normal backup. No MSI installation, service registration, real IM message, or cold boot test was performed; only this Windows client's runtime was explicitly upgraded. Project/session data was not migrated by this change.
