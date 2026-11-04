#!/usr/bin/env bash
# Backend/web deployment; retain the currently published signed client packages.
set -Eeuo pipefail
label=${1:?release label}
bundle_hash=${2:?bundle SHA-256}
[[ "$label" =~ ^[0-9a-f]{40}-[0-9]{14}$ ]]
[[ "$bundle_hash" =~ ^[0-9a-f]{64}$ ]]
test "$(id -u)" -eq 0
stage="/tmp/remote-agent-$label"
release="/opt/remote-agent/releases/$label"
backup="/var/backups/remote-agent/$label"
config=/etc/nginx/conf.d/remote-agent.conf
snippet=/etc/nginx/remote-agent-shared-files.conf
exec 9>/run/lock/remote-agent-publish.lock
flock -n 9
previous=$(readlink -f /opt/remote-agent/current)
[[ "$previous" == /opt/remote-agent/releases/* ]]
test -x "$previous/remote-server"
test -d "$previous/dist"
test ! -e "$release"
test ! -e "$backup"
printf '%s  %s\n' "$bundle_hash" "$stage/server.tar.gz" | sha256sum -c -
install -d -m 0755 "$release"
tar -xzf "$stage/server.tar.gz" -C "$release"
(cd "$release" && sha256sum -c remote-server.sha256)
test -f "$release/SOURCE_REVISION"
test "$(cat "$release/SOURCE_REVISION")" = "${label:0:40}"
chown -R root:root "$release"
chmod -R u+rwX,go+rX,go-w "$release"
chmod 0755 "$release/remote-server"
ln -s "$previous/dist" "$release/dist"
install -d -m 0700 "$backup"
printf '%s\n' "$previous" > "$backup/previous-release"
cp -a /etc/remote-agent "$backup/config"
cp -a /etc/systemd/system/remote-agent.service "$backup/remote-agent.service"
cp -L --preserve=mode,ownership,timestamps "$config" "$backup/nginx.conf"
if test -e "$snippet"; then cp -a "$snippet" "$backup/shared-files.conf"; fi
data_root=$(readlink -f /var/lib/remote-agent)
[[ "$data_root" = /var/lib/private/remote-agent || "$data_root" = /var/lib/remote-agent ]]
backup_ready=0
rollback() {
    trap - ERR HUP INT TERM
    set +e
    systemctl stop remote-agent
    if test "$backup_ready" = 1; then
        mv "$data_root" "$backup/failed-candidate-data"
        cp -a "$backup/data" "$data_root"
    fi
    cat "$backup/nginx.conf" > "$config"
    if test -e "$backup/shared-files.conf"; then
        cp -a "$backup/shared-files.conf" "$snippet"
    else
        rm -f /etc/nginx/remote-agent-shared-files.conf
    fi
    nginx -t && systemctl reload nginx
    ln -sfn "$previous" /opt/remote-agent/current.rollback
    mv -Tf /opt/remote-agent/current.rollback /opt/remote-agent/current
    systemctl start remote-agent
    echo "Deployment failed; previous program/data restored. Recovery files: $backup" >&2
    exit 1
}
trap rollback ERR HUP INT TERM
install -m 0644 "$release/shared-files-location.conf" "$snippet"
python3 - "$config" <<'PY'
import pathlib, sys
path = pathlib.Path(sys.argv[1])
text = path.read_text()
include = '    include /etc/nginx/remote-agent-shared-files.conf;'
if include not in text:
    anchor = '    location = /agents { return 308 /agents/; }'
    if text.count(anchor) != 1:
        raise SystemExit('Unexpected Nginx layout; refusing to rewrite existing vhosts')
    text = text.replace(anchor, anchor + '\n' + include)
    path.write_text(text)
PY
nginx -t
systemctl stop remote-agent
cp -a "$data_root" "$backup/data"
backup_ready=1
ln -s "$release" /opt/remote-agent/current.next
mv -Tf /opt/remote-agent/current.next /opt/remote-agent/current
systemctl start remote-agent
healthy=0
for attempt in $(seq 1 20); do
    if curl -fsS --max-time 3 http://127.0.0.1:8086/agents/healthz; then healthy=1; break; fi
    sleep 1
done
test "$healthy" = 1
# An anonymous binary upload must reach the new authenticated route (401).
test "$(curl -sS -o /dev/null -w '%{http_code}' --data-binary '' -H 'Content-Type: application/octet-stream' http://127.0.0.1:8086/agents/api/agent/files)" = 401
systemctl reload nginx
curl -fsS --max-time 10 --resolve codex.sunny-string.cn:443:127.0.0.1 https://codex.sunny-string.cn/agents/healthz
test "$(curl -sS --max-time 10 -o /dev/null -w '%{http_code}' --resolve codex.sunny-string.cn:443:127.0.0.1 --data-binary '' -H 'Content-Type: application/octet-stream' https://codex.sunny-string.cn/agents/api/agent/files)" = 401
systemctl is-active remote-agent nginx
pid=$(systemctl show remote-agent -p MainPID --value)
test "$(readlink -f "/proc/$pid/exe")" = "$release/remote-server"
trap - ERR HUP INT TERM
printf '\nPublished %s\nBackup: %s\n' "$label" "$backup"
