#!/usr/bin/env bash
# Publish signed runtime packages while retaining MSI installers and the backend.
set -Eeuo pipefail
label=${1:?release label}
bundle_hash=${2:?bundle SHA-256}
[[ "$label" =~ ^runtime-[0-9a-f]{40}-[0-9]+$ ]]
[[ "$bundle_hash" =~ ^[0-9a-f]{64}$ ]]
test "$(id -u)" = 0
stage="/tmp/remote-agent-$label"
target="/opt/remote-agent/client-releases/$label"
backup="/var/backups/remote-agent-clients/$label"
exec 9>/run/lock/remote-agent-publish.lock
flock -n 9
release=$(readlink -f /opt/remote-agent/current)
[[ "$release" == /opt/remote-agent/releases/* ]]
previous=$(readlink -f "$release/dist")
[[ "$previous" == /opt/remote-agent/* ]]
test -d "$previous"
test ! -e "$target"
test ! -e "$backup"
printf '%s  %s\n' "$bundle_hash" "$stage/runtime.tar.gz" | sha256sum -c -
install -d -m 0755 "$target"
cp -a "$previous/." "$target/"
tar -xzf "$stage/runtime.tar.gz" -C "$target"
(cd "$target" && sha256sum -c files.sha256)
python3 - "$target" "$previous" "${label:8:40}" <<'PY'
import hashlib, json, pathlib, sys
root, old = map(pathlib.Path, sys.argv[1:3])
provenance = json.loads((root / 'RUNTIME_RELEASE.json').read_text())
assert provenance['schemaVersion'] == 1 and provenance['revision'] == sys.argv[3]
for platform in ('windows', 'darwin', 'linux'):
    for arch in ('amd64', 'arm64'):
        relative = pathlib.Path('updates/v1') / platform / (arch + '.json')
        meta = json.loads((root / relative).read_text())
        assert meta['version'] == provenance['version'] and meta['sequence'] == provenance['sequence']
        assert meta['os'] == platform and meta['arch'] == arch and meta['keyId'] == 'release-2026' and meta['signature']
        artifact = ('remote-agent-core-windows-' + arch + '.exe') if platform == 'windows' else ('remote-agent-' + platform + '-' + arch)
        raw = (root / artifact).read_bytes()
        assert len(raw) == meta['size'] and hashlib.sha256(raw).hexdigest() == meta['sha256']
        if (old / relative).exists():
            assert json.loads((old / relative).read_text())['sequence'] < meta['sequence'], 'Release sequence must increase'
print('Six runtime packages, provenance and monotonic release sequences verified')
PY
chown -R root:root "$target"
# Archives built on Windows may carry 0777 directories or 0666 files.
# Grant public read access while removing every group/other write permission.
chmod -R u+rwX,go+rX,go-w "$target"
install -d -m 0700 "$backup"
printf '%s\n' "$previous" > "$backup/previous-dist"
printf '%s\n' "$release" > "$backup/server-release"
cat > "$backup/rollback.sh" <<'ROLLBACK'
#!/usr/bin/env bash
set -Eeuo pipefail
test "$(id -u)" = 0
backup=$(cd -- "$(dirname -- "$0")" && pwd)
exec 9>/run/lock/remote-agent-publish.lock
flock -n 9
release=$(cat "$backup/server-release")
previous=$(cat "$backup/previous-dist")
[[ "$release" == /opt/remote-agent/releases/* && "$previous" == /opt/remote-agent/* ]]
test "$(readlink -f /opt/remote-agent/current)" = "$release"
test -d "$previous"
temporary="$release/dist.rollback-$$"
ln -s "$previous" "$temporary"
mv -Tf "$temporary" "$release/dist"
echo 'Previous update channel restored. Client runtime rollback is a separate explicit operation.'
ROLLBACK
chmod 0700 "$backup/rollback.sh"
published=0
rollback() {
    trap - ERR HUP INT TERM
    if test "$published" = 1; then
        ln -s "$previous" "$release/dist.rollback-$$"
        mv -Tf "$release/dist.rollback-$$" "$release/dist"
    fi
    echo "Runtime publication failed; previous packages retained. Backup: $backup" >&2
    exit 1
}
trap rollback ERR HUP INT TERM
ln -s "$target" "$release/dist.next-$$"
# Arm recovery before the atomic switch, including termination during mv.
published=1
mv -Tf "$release/dist.next-$$" "$release/dist"
python3 - "$target" <<'PY'
import hashlib, json, pathlib, subprocess, sys
root = pathlib.Path(sys.argv[1])
base = 'https://codex.sunny-string.cn/agents'
def fetch(path):
    return subprocess.check_output(['curl', '-fsS', '--max-time', '30', '--resolve', 'codex.sunny-string.cn:443:127.0.0.1', base + path])
for platform in ('windows', 'darwin', 'linux'):
    for arch in ('amd64', 'arm64'):
        relative = f'updates/v1/{platform}/{arch}.json'
        expected = json.loads((root / relative).read_text())
        assert json.loads(fetch('/' + relative)) == expected
        artifact = f'remote-agent-core-windows-{arch}.exe' if platform == 'windows' else f'remote-agent-{platform}-{arch}'
        raw = fetch('/downloads/' + artifact)
        assert len(raw) == expected['size'] and hashlib.sha256(raw).hexdigest() == expected['sha256']
assert json.loads(fetch('/healthz'))['ok'] is True
print('Public update manifests, all six downloads and server health verified')
PY
trap - ERR HUP INT TERM
printf '\nPublished signed runtime packages: %s\nRollback: bash %s/rollback.sh\n' "$label" "$backup"
