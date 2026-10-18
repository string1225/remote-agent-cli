#!/usr/bin/env bash
set -euo pipefail
umask 077
server='__SERVER__'
mode='__MODE__'
setup_codex='__SETUP_CODEX__'
no_start=false
for option in "$@"; do case "$option" in --no-start) no_start=true ;; --skip-codex) setup_codex=false ;; *) echo "Unknown option: $option" >&2; exit 1 ;; esac; done
if [[ "$(uname -s)" != Darwin ]]; then echo 'This installer requires macOS.' >&2; exit 1; fi
root="$HOME/.remote-agent"
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
export PATH="$root/bin:$PATH"
path_line='export PATH="$HOME/.remote-agent/bin:$PATH"'
if [[ "$setup_codex" == true || "$mode" == cli || "$mode" == codex ]]; then
for rc in "$HOME/.profile" "$HOME/.zshrc"; do
    if ! grep -Fqx "$path_line" "$rc" 2>/dev/null; then printf '\n%s\n' "$path_line" >> "$rc"; fi
done
fi
if [[ "$mode" == cli || "$mode" == codex ]]; then
    if [[ -e "$root/config.json" ]]; then
        "$temp/agent" update --installer --preserve-start --server "$server" --config "$root/config.json"
    else
        mv "$temp/agent" "$root/bin/remote-agent"
    fi
    if [[ "$mode" == cli ]]; then echo 'Local CLI installed. No remote service enabled. Run remote-agent --help in a new terminal.'; exit 0; fi
fi
if [[ "$mode" == codex ]]; then
    IFS= read -r -s -p 'Paste a personal proxy Token from /agents (input hidden): ' proxy_token </dev/tty
    printf '\n' >/dev/tty
    printf '%s' "$proxy_token" | "$root/bin/remote-agent" codex setup --server "$server" --token-stdin
    unset proxy_token
    echo 'Codex ready. Open a new terminal to use codex. Restart Codex App and create a new local chat.'
    exit 0
fi
if [[ -e "$root/config.json" ]]; then
    if [[ "$setup_codex" == true ]]; then "$temp/agent" codex setup --config "$root/config.json" --server "$server"; fi
    "$temp/agent" update --installer --preserve-start="$no_start" --server "$server" --config "$root/config.json"
    if [[ "$no_start" == true ]]; then "$root/bin/remote-agent" stop --config "$root/config.json"; else "$root/bin/remote-agent" autostart install --config "$root/config.json"; fi
    echo 'Existing agent upgraded. Its account binding, projects, permissions, and conversations are retained. Reconnect in the web console.'
    exit 0
fi
mv "$temp/agent" "$root/bin/remote-agent"
printf '%s' '__TOKEN__' | "$root/bin/remote-agent" enroll --server "$server" --config "$root/config.json" --token-stdin
if [[ "$setup_codex" == true ]]; then "$root/bin/remote-agent" codex setup --config "$root/config.json"; fi
if [[ "$no_start" == true ]]; then echo 'CLI installed and bound; remote service is stopped. Use remote-agent start when needed.'; exit 0; fi
"$root/bin/remote-agent" autostart install --config "$root/config.json"
echo 'Codex and Agent ready. Return to the web console, connect this device and add a project. Qoder remains optional.'
