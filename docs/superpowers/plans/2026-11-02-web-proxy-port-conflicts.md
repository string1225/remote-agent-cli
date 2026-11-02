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
- [ ] Review only intended changes; commit and push `main`.
- [ ] Upgrade this client's signed core through the existing updater; verify a conflict with the real occupied port, a repeated plain `on`, and system-proxy HTTP access. Preserve the installed service, client identity, project/session data, and unrelated repository edits.
