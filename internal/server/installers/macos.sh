#!/usr/bin/env bash
set -euo pipefail
umask 077
server='__SERVER__'
if [[ "$(uname -s)" != Darwin ]]; then echo 'This installer requires macOS.' >&2; exit 1; fi
root="$HOME/.remote-agent"
if [[ -e "$root/config.json" ]]; then echo 'An agent is already bound. Revoke it and remove its config before reinstalling.' >&2; exit 1; fi
case "$(uname -m)" in arm64) arch=arm64 ;; x86_64) arch=amd64 ;; *) echo 'Unsupported architecture.' >&2; exit 1 ;; esac
mkdir -p "$root/bin"
temp="$(mktemp -d "$root/install.XXXXXX")"
trap 'rm -f "$temp/agent" "$temp/checksum"; rmdir "$temp"' EXIT
name="remote-agent-darwin-$arch"
curl --fail --silent --show-error --connect-timeout 20 "$server/downloads/$name" -o "$temp/agent"
curl --fail --silent --show-error --connect-timeout 20 "$server/downloads/$name.sha256" -o "$temp/checksum"
expected="$(awk '{print $1}' "$temp/checksum")"
actual="$(shasum -a 256 "$temp/agent" | awk '{print $1}')"
if [[ ! "$expected" =~ ^[a-fA-F0-9]{64}$ || "$expected" != "$actual" ]]; then echo 'Download checksum mismatch.' >&2; exit 1; fi
chmod 700 "$temp/agent"
mv "$temp/agent" "$root/bin/remote-agent"
printf '%s' '__TOKEN__' | "$root/bin/remote-agent" enroll --server "$server" --config "$root/config.json" --token-stdin
"$root/bin/remote-agent" autostart install --config "$root/config.json"
echo 'Agent installed and started. Return to the web console to register Codex or Qoder.'
