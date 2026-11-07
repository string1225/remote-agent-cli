# Global Remote Agent skill lifecycle

Remote Agent already exposes explicit IM notification and temporary file upload through its local MCP server; upload also has a CLI command. Other local agents need discovery instructions, not copies of device/model/bot credentials or fabricated session identities.

Use the official user-level shared location `~/.agents/skills/remote-agent`. Keep the embedded SKILL.md as the release source, track its version/hash in a small manifest, and migrate existing real `remote-agent` skill folders out of legacy Codex skills roots into timestamped backups outside discovery paths. Preserve all old resources and user edits; reject symlinks/concurrent installs. No permanent MCP/model configuration is changed.

- [x] Add explicit-home sync/status APIs, isolated tests for canonical installation, legacy migration, version changes, idempotence and failure preservation.
- [x] Add `remote-agent skill install/status` independent of binding/Codex installation. Reuse it from Codex setup and installation.
- [x] Sync only on activated Windows workers and Unix CLI daemon startup, never during worker preparation/probes. Use the owning user's explicit profile; log skill failures without preventing the client from connecting.
- [x] Call skill installation in Unix installer templates and after successful Windows MSI installation in the invoking user context. No MSI installation test runs on this developer host.
- [x] Update the skill/docs to explain use by standalone local agents, explicit IM targets, public temporary file links, global installation and version synchronization.
- [ ] Run full Go tests/vet, platform and MSI package/static checks; review, commit/push, publish runtime/MSI and deploy changed server installer templates with existing backup/rollback procedures.
- [ ] Update Windows and 139 clients; verify automatic migration/version markers, discovery and MCP tool listing without sending messages or uploading private files. Preserve configuration, sessions, recordings and unrelated dirty files.
