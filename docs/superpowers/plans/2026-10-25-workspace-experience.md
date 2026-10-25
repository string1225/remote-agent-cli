# Workspace experience implementation plan

> **For agentic workers:** use focused parallel workers for the independent binding and IM subsystems; the primary worker integrates and verifies the interface. Develop directly on main as required by AGENTS.md.

**Goal:** make device binding, conversations, channels and account credentials work as one understandable workflow.

**Architecture:** retain the Go client/server and embedded vanilla web application. Device authentication determines the account used for model credentials. Each IM conversation keeps its own persistent session and notification route. Preserve existing identities, history, recordings and credentials.

**Design direction:** white conversation canvas, slate navigation, blue selection/action color, green online status with explicit text and outlined gray offline status. Use 13–15 px controls and readable Markdown, recognizable OS logos, restrained borders and generous message spacing. The composer is a non-scrolling footer. Device actions live in a keyboard-accessible context menu; account Tokens use aligned right-hand actions. No transport implementation labels in the workspace. Image generation is unavailable in this session; implement the design with native HTML/CSS/SVG and inspect desktop/mobile renders using bundled Playwright (Browser plugin unavailable).

## Binding and execution

- [ ] Make device-token provisioning idempotent and account-scoped, including retry after partial setup.
- [ ] Show credential account identity without printing secrets; default to the enrolled device for `codex setup`.
- [ ] Apply full authorization to Codex/Qoder, Web/IM/SaySo and compatible old configuration reads.
- [ ] Register a project's first Web session on the client; return it to the UI and avoid duplicates on retry.
- [ ] Validate using temporary configuration and mock HTTP/CLI fixtures; never modify real service state in tests.

## Channels

- [ ] Diagnose local Feishu using sanitized logs/configuration and read-only official API checks.
- [ ] Accept all users by default; infer private/group target from incoming messages and preserve separate sessions and names.
- [ ] Send Markdown and preserve exact notification targets through nested/cross-device tasks.
- [ ] Add/remove Feishu THINKING through full reply lifecycle; persist cleanup for interrupted work and expose unsupported DingTalk capability honestly.
- [ ] Surface connection stages and readable errors; keep a success modal open until dismissed.
- [ ] Verify target isolation, old history preservation, cancellation, failed delivery/recovery and reaction cleanup with mocks.

## Workspace and account

- [ ] Add owner-only device rename API and an accessible device menu with rename/repair actions.
- [ ] Make repair explain installation/update, reuse of local binding, and loss-of-config recovery without discarding old data.
- [ ] Render sanitized Markdown, including code and tables; constrain scroll containers and keep composer visible on desktop/mobile.
- [ ] Enter sends, Shift+Enter inserts newline, input-method composition must not submit.
- [ ] Remove manual Token import and align Copy/Revoke actions. Explain install → bind → account-scoped Token setup → launch Codex.
- [ ] Replace OS symbols with recognizable licensed vector logos and strengthen device status contrast.
- [ ] Remove disconnect button and relay/WebSocket jargon from product screens.

## Integration and delivery

- [ ] Run Go tests/vet, JS syntax, targeted browser interactions, desktop/mobile screenshots, Markdown security checks and platform builds.
- [ ] Review changes, update current capability documentation, commit only this task's files and push origin/main.
- [ ] Publish a consistent server/client/MSI release; back up database and encryption key together and health-check activation.
- [ ] Report release/hash, evidence and remaining real-device/IM acceptance.
