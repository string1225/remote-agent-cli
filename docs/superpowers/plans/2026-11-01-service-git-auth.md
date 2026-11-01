# Windows Service Git Authorization Implementation Plan

> **For agentic workers:** Use superpowers:executing-plans to implement this plan in this session. The user has approved the design and requires development directly on main, preserving existing work.

**Goal:** Let Windows service tasks authenticate GitHub independently and remind users during client initialization.

**Architecture:** A bounded local pipe routes fixed GCM commands to the LocalService business worker. GCM uses a private DPAPI directory; only service process environment selects it. GitHub helper overrides preserve desktop files and other Git settings.

**Tech Stack:** Go, go-winio named pipes, Windows access tokens/DACL, Git for Windows and GCM.

---

### Task 1: GCM command and environment boundary

Files: `internal/gitauth/manager.go`, `environment.go`, matching tests and platform command helpers.

- [x] Write tests for managed environment preserving unrelated Git settings, strict actions/accounts, streamed device-code output, failed authorization and cancellation using a subprocess fixture.
- [x] Run `go test ./internal/gitauth` and observe the missing implementation failure.
- [x] Implement fixed login/status/logout arguments, GCM discovery, DPAPI environment and noninteractive task settings, including failing askpass guidance.
- [x] Re-run tests; verify effective GitHub helper using an isolated real Git config.

### Task 2: Protected local transport and service integration

Files: `internal/gitauth/protocol.go`, `windows.go`, matching tests; `internal/agent/git_auth_windows.go`, `git_auth_other.go`; `internal/agent/agent.go`.

- [x] Add request-boundary tests using temporary connections: unauthorized callers, unknown fields, concurrent mutation, cancellation and bounded output.
- [x] Add Windows tests for fake SID authorization decisions, protected directory ACL and real local pipe with current-user fixture identities.
- [x] Implement protocol version 1, per-request deadlines and cancellation; protect the store with inherited service-SID ACL and OWNER RIGHTS. Use identification-level pipe tokens and SCM worker identity rather than cross-user process token queries.
- [x] Start the listener and install task environment only for the verified RemoteAgent LocalService worker; restore process environment on shutdown. Keep an explicit failure diagnostic without stopping unrelated client functionality.
- [x] Run `go test ./internal/gitauth ./internal/agent`.

### Task 3: CLI and initialization reminders

Files: `cmd/agent/git.go`, `git_test.go`, `main.go`, `connect.go`, `service.go`; localized messages; `web/index.html`; `docs/windows-service-msi.md`.

- [x] Add tests for parsing, logout account requirements and reminders preserving paths with spaces/custom configuration and no-start guidance. Non-Windows builds retain the normal user authentication flow.
- [x] Route `remote-agent git` to the local service and print help without reading credentials.
- [x] Add successful Windows initialization reminders, including repaired bindings, and a Windows-specific onboarding note in the webpage and documentation.
- [x] Run `go test ./cmd/agent ./internal/i18n` and JS syntax checks.

### Task 4: Verify, review and publish

- [x] Run gofmt on changed Go files, `go test ./...` and `go vet ./...`.
- [x] Build Windows amd64/arm64 and Linux amd64 core to a temporary directory; verify installed GCM DPAPI roundtrips using only temporary fake credentials.
- [x] Independently review secret exposure, transport identity and recovery behavior. Verify Windows-only onboarding visibility and mobile layout in Chrome/Playwright without real user data.

Publication follows verification: stage only this task's hunks and files, commit directly to main and push origin/main without force. The final response records the resulting commit and unperformed real LocalService/login/sign-out/reboot acceptance. The reparse-directory test skips when Windows denies symbolic-link creation.
