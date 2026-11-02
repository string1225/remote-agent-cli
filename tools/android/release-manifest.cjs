const fs = require('node:fs');
const crypto = require('node:crypto');
const path = require('node:path');
const code = Number(process.env.ANDROID_VERSION_CODE), version = process.env.ANDROID_VERSION_NAME;
if (!Number.isSafeInteger(code) || code < 1 || code > 2100000000 || !/^\d+\.\d+\.\d+\.\d+$/.test(version || '')) throw Error('Invalid Android version');
const repository = process.env.GITHUB_REPOSITORY || 'string1225/remote-agent-cli';
if (repository !== 'string1225/remote-agent-cli') throw Error('Update channel repository differs from the Android client');
const output = 'dist/android'; fs.mkdirSync(output, {recursive:true});
const apk = fs.readFileSync('android/app/build/outputs/apk/release/app-release.apk');
const hash = crypto.createHash('sha256').update(apk).digest('hex');
fs.writeFileSync(path.join(output, 'remote-agent.apk'), apk);
fs.writeFileSync(path.join(output, 'remote-agent.apk.sha256'), hash + '  remote-agent.apk\n');
fs.writeFileSync(path.join(output, 'android-update.json'), JSON.stringify({versionCode:code, versionName:version, sha256:hash,
  apkUrl:`https://gh-proxy.com/https://github.com/${repository}/releases/download/android-v${version}/remote-agent.apk`,
  commit:process.env.GITHUB_SHA || ''}, null, 2) + '\n');
