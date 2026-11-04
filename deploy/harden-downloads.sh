#!/usr/bin/env bash
# Repair permissions of the active download directory without changing packages.
set -Eeuo pipefail
test "$(id -u)" = 0
exec 9>/run/lock/remote-agent-publish.lock
flock -n 9
dist=$(readlink -f /opt/remote-agent/current/dist)
[[ "$dist" == /opt/remote-agent/releases/* || "$dist" == /opt/remote-agent/client-releases/* ]]
backup="/var/backups/remote-agent-clients/download-permissions-$(date +%Y%m%d%H%M%S)-$$"
install -d -m 0700 "$backup"
python3 - "$dist" "$backup" <<'PY'
import json, os, pathlib, stat, subprocess, sys
root, backup = map(pathlib.Path, sys.argv[1:])
paths = [root] + sorted(root.rglob('*'), key=lambda p: len(p.parts))
entries = []
for path in paths:
    info = path.lstat()
    if stat.S_ISLNK(info.st_mode):
        raise RuntimeError('Unexpected symlink in active download directory')
    if not (stat.S_ISDIR(info.st_mode) or stat.S_ISREG(info.st_mode)) or info.st_uid != 0:
        raise RuntimeError('Download files and directories must be root owned')
    entries.append({'path': str(path), 'mode': stat.S_IMODE(info.st_mode)})
(backup / 'modes.json').write_text(json.dumps(entries, indent=2))
(backup / 'restore.py').write_text('''import json, os, pathlib
entries = json.loads((pathlib.Path(__file__).parent / 'modes.json').read_text())
for entry in entries:
    os.chmod(entry['path'], entry['mode'])
print('Original download permissions restored')
''')
try:
    for entry in entries:
        os.chmod(entry['path'], entry['mode'] & ~0o022)
    for entry in entries:
        assert not stat.S_IMODE(os.stat(entry['path']).st_mode) & 0o022
    raw = subprocess.check_output(['curl', '-fsS', '--max-time', '15', 'https://codex.sunny-string.cn/agents/healthz'])
    assert json.loads(raw)['ok'] is True
except BaseException:
    for entry in entries:
        os.chmod(entry['path'], entry['mode'])
    raise
print(f'Group/other write permissions removed from {len(entries)} root-owned entries')
print(f'Permission backup: {backup}; rollback: python3 {backup}/restore.py')
PY
