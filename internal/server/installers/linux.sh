#!/usr/bin/env bash
set -euo pipefail
umask 077
server='__SERVER__'
if [[ "$(uname -s)" != Linux ]]; then echo 'This installer requires Linux.' >&2; exit 1; fi
case "$(uname -m)" in aarch64|arm64) arch=arm64 ;; x86_64|amd64) arch=amd64 ;; *) echo 'Unsupported architecture.' >&2; exit 1 ;; esac
for dependency in curl sha256sum systemctl; do
    if ! command -v "$dependency" >/dev/null; then echo "Required command not found: $dependency" >&2; exit 1; fi
done
if ! systemctl --user show-environment >/dev/null; then
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
exe="$root/bin/remote-agent"
if [[ -e "$root/config.json" ]]; then
    "$temp/agent" update --installer --server "$server" --config "$root/config.json"
    echo 'Existing account binding, projects, permissions, and conversations retained.'
else
    mv "$temp/agent" "$exe"
    printf '%s' '__TOKEN__' | "$exe" enroll --server "$server" --config "$root/config.json" --token-stdin
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
echo 'Agent installed and started with systemd --user. Return to the web console to connect this device. Codex and Qoder are optional.'
