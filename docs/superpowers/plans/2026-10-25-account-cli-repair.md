# Account tokens and client repair

Goal: copy saved personal tokens, choose Chinese/English CLI messages from the MSI installation language, and reopen binding/repair instructions for existing clients. Diagnose Windows elevation failures without losing the original binding.

- [x] Add encrypted token storage with a private sibling key file. Keep SHA-256 authentication and metadata-only lists. Owner-only reveal/import endpoints must reject other users, disabled accounts and revoked tokens. Legacy hashes remain valid; importing a known token verifies the hash without rotating it. Test reopen, missing keys and ownership.
- [x] Add copy/import controls to My Tokens, clear secrets on close/logout, and guard asynchronous responses against account/dialog changes. Exercise the controls in a local browser with fixture credentials.
- [x] Store CLI language in the MSI registry component using the Windows system language (Chinese primary language -> zh, otherwise en). Localize CLI help, prompts, progress and actionable errors, keeping JSON/MCP output intact. Test language selection and format strings; compile x64/ARM64 MSI and Unix agents.
- [x] Add a repair button beside existing devices. Reissue pending enrollment instructions for the same device; enrolled devices reuse their existing local configuration without issuing replacement credentials. Preserve identity, projects, histories and recordings.
- [x] Persist elevated service setup diagnostics and distinguish UAC cancellation from a child-process failure. Include stage and exit status, retain logs on failure, and keep tokens out of logs. Cover the handoff using temporary files and mocked commands; never install/control a real service in ordinary tests.
- [x] Run gofmt, go test ./..., go vet ./..., both workspace JS syntax checks, browser interaction checks and platform builds. Review the intended diff; the final commit/push and release activation follow these checks.

Observed live failure (read-only investigation, 2026-10-25): the local edf73ba MSI service is running but unconfigured; no ProgramData/RemoteAgent/service.json exists. connect preserved the user's old binding. The old elevation launcher discarded the child error, so its exact failure cannot be reconstructed from existing logs. Local runtime agent.log has no recent setup entry. The 139 journal is healthy; Nginx records older clients receiving 426 and is not evidence of this local setup failure.

Verified outcome: real local service repair completed on the original binding. The home-directory ACL grant took 3m31.337s, exceeding the old 60s timeout. service status now reports configured=true/running/automatic; the original device is online in the real 139 catalog. No MSI installation test was run on the development machine.

Validation: go test ./..., go vet ./..., JS checks, CLI language subprocess checks, native PowerShell/fixture handoff with UAC mocked, AES-GCM/ownership/legacy-import tests, x64/ARM64 MSI builds, and Linux/macOS amd64/arm64 builds passed. Local Edge/Playwright exercised copy after reload, clipboard, legacy import, secret clearing and pending/enrolled repair at 1440x960 and 390x844. The Browser plugin was unavailable; bundled Playwright was used. Only the expected anonymous /api/me 401 occurred. Screenshots are temporary artifacts outside the repository.

Remaining target acceptance: multilingual MSI installation/upgrade on disposable administrator runners, native ARM64, full machine cold boot/logout, and Linux race suite. The actual fault repair above does not replace those platform acceptance checks.
