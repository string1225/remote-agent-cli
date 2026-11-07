# Web proxy split routing and direct traffic isolation

The regression is reproduced: 32 ordinary CONNECT tunnels exhaust the local relay quota and cause unrelated HTTP/CONNECT requests to receive 429. Windows currently sends all browser traffic to that listener through a manual global proxy.

The authorized behavior is unchanged: only the existing GitHub/ChatGPT/OpenAI/Google domain list and subdomains relay through 139; all other requests connect from this machine through the local listener. Apply the 32-device/256-server relay quota only to relay destinations. Retain cancellation and finite connection lifetimes for direct connections. The user explicitly chose local forwarding and requested Git proxy synchronization.

- [x] Reproduce saturation using isolated loopback upstreams and mock binding credentials.
- [x] Keep the existing local HTTP proxy/system proxy architecture and domain list.
- [x] Synchronize the invoking user's Git HTTP proxy with on/off and the actual port, privately back up/restore original proxy values, and preserve external changes. Status identifies Git proxy ownership.
- [x] Isolate direct HTTP/CONNECT from relay quota and retain shutdown cancellation.
- [x] Test saturation and 139 failure, Git proxy switching/restoration/external edits, and existing proxy behavior using temporary Git configuration. Review regressions also cover failed port-switch journal recovery and effective included/URL-specific Git overrides.
- [x] Run full Go tests/vet, six-platform builds and independent review; all checks passed and review findings were resolved.
- [x] Commit/push only intended files (`24e1bba`).
- [x] Publish the existing runtime channel with backup/rollback, update this machine and 139 client, and verify real direct/relay requests under load, Git clone without per-command proxy flags, off/on restoration and unchanged device/session data. See `docs/deployments/2026-11-07-web-proxy-direct-git.md`.

Windows on remains explicit and default off. Only the Git HTTP proxy setting is managed alongside it; other Git settings and authentication remain unchanged. No PAC, persistent environment variable, certificate setting, relay domain expansion, new service or scheduled task is introduced.
