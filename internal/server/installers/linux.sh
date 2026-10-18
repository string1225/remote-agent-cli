#!/usr/bin/env bash
set -euo pipefail
umask 077
server='__SERVER__'
mode='__MODE__'
setup_codex='__SETUP_CODEX__'
no_start=false
for option in "$@"; do case "$option" in --no-start) no_start=true ;; --skip-codex) setup_codex=false ;; *) echo "Unknown option: $option" >&2; exit 1 ;; esac; done
if [[ "$(uname -s)" != Linux ]]; then echo 'This installer requires Linux.' >&2; exit 1; fi
case "$(uname -m)" in aarch64|arm64) arch=arm64 ;; x86_64|amd64) arch=amd64 ;; *) echo 'Unsupported architecture.' >&2; exit 1 ;; esac
for dependency in curl sha256sum; do
    if ! command -v "$dependency" >/dev/null; then echo "Required command not found: $dependency" >&2; exit 1; fi
done
if [[ "$mode" != codex && "$mode" != cli && "$no_start" != true ]] && ! systemctl --user show-environment >/dev/null; then
    echo 'A systemd user session is required. Run this command as the target user in a normal login session, without sudo.' >&2
    exit 1
fi
root="$HOME/.remote-agent"
mkdir -p "$root/bin"
chmod 700 "$root" "$root/bin"
temp="$(mktemp -d "$root/install.XXXXXX")"
trap 'rm -f "$temp/agent" "$temp/checksum"; rmdir "$temp"' EXIT
name="remote-agent-linux-$arch"
curl --fail --silent --show-error --connect-timeout 20 --max-time 180 "$server/downloads/$name" -o "$temp/agent"
curl --fail --silent --show-error --connect-timeout 20 --max-time 30 "$server/downloads/$name.sha256" -o "$temp/checksum"
expected="$(awk '{print $1}' "$temp/checksum")"
actual="$(sha256sum "$temp/agent" | awk '{print $1}')"
if [[ ! "$expected" =~ ^[a-fA-F0-9]{64}$ || "$expected" != "$actual" ]]; then echo 'Download checksum mismatch.' >&2; exit 1; fi
chmod 700 "$temp/agent"
export PATH="$root/bin:$PATH"
path_line='export PATH="$HOME/.remote-agent/bin:$PATH"'
if [[ "$setup_codex" == true || "$mode" == cli || "$mode" == codex ]]; then
for rc in "$HOME/.profile" "$HOME/.bashrc"; do
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
exe="$root/bin/remote-agent"
if [[ -e "$root/config.json" ]]; then
    if [[ "$setup_codex" == true ]]; then "$temp/agent" codex setup --config "$root/config.json" --server "$server"; fi
    "$temp/agent" update --installer --preserve-start="$no_start" --server "$server" --config "$root/config.json"
    echo 'Existing account binding, projects, permissions, and conversations retained.'
else
    mv "$temp/agent" "$exe"
    printf '%s' '__TOKEN__' | "$exe" enroll --server "$server" --config "$root/config.json" --token-stdin
    if [[ "$setup_codex" == true ]]; then "$exe" codex setup --config "$root/config.json"; fi
fi
if [[ "$no_start" == true ]]; then
    if [[ -e "${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user/remote-agent.service" ]]; then "$exe" stop --config "$root/config.json"; fi
    echo 'CLI installed and bound; remote service is stopped. Use remote-agent start when needed.'
    exit 0
fi
"$exe" autostart install --config "$root/config.json"
# Keep the user manager alive after SSH logout and start it at boot. Do not run
# the installer as root; only this optional loginctl operation may use sudo.
uid="$(id -u)"
if command -v loginctl >/dev/null && [[ "$(loginctl show-user "$uid" -p Linger --value 2>/dev/null || true)" != yes ]]; then
    if loginctl --no-ask-password enable-linger "$uid" >/dev/null 2>&1; then
        echo 'Enabled systemd user lingering for boot and logout recovery.'
    elif command -v sudo >/dev/null && sudo -n loginctl enable-linger "$uid"; then
        echo 'Enabled systemd user lingering for boot and logout recovery.'
    else
        echo "Agent is running for this login. To keep it running after logout and at boot, run: sudo loginctl enable-linger $uid" >&2
    fi
fi
echo 'Codex and Agent ready with systemd --user. Return to the web console, connect this device and add a project. Qoder remains optional.'
