/* Run from the repository root. Signing material stays in ignored .local/;
 * --upload installs it as Actions secrets using the existing Git/GH login. */
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const cp = require('node:child_process');
const directory = path.resolve('.local/android-signing');
fs.mkdirSync(directory, {recursive:true,mode:0o700});
const settings = path.join(directory,'signing.json'), keystore = path.join(directory,'release.jks');
if (!fs.existsSync(settings)) {
  if (fs.existsSync(keystore)) throw Error('Signing settings are missing. Recover the original password; never replace an existing key.');
  fs.writeFileSync(settings,JSON.stringify({password:crypto.randomBytes(32).toString('hex'),alias:'remote-agent'},null,2),{mode:0o600,flag:'wx'});
}
const config = JSON.parse(fs.readFileSync(settings,'utf8'));
const env = {...process.env,ANDROID_KEYSTORE_PATH:keystore,ANDROID_KEYSTORE_PASSWORD:config.password,ANDROID_KEY_ALIAS:config.alias,ANDROID_KEY_PASSWORD:config.password};
if (!fs.existsSync(keystore)) {
  const keytool = process.env.JAVA_HOME ? path.join(process.env.JAVA_HOME,'bin',process.platform==='win32'?'keytool.exe':'keytool') : 'keytool';
  cp.execFileSync(keytool,['-genkeypair','-keystore',keystore,'-storetype','PKCS12','-alias',config.alias,'-keyalg','RSA','-keysize','3072','-validity','10000','-dname','CN=Remote Agent Android','-storepass:env','ANDROID_KEYSTORE_PASSWORD','-keypass:env','ANDROID_KEY_PASSWORD'],{env,stdio:['ignore','ignore','pipe']});
}
if (process.argv.includes('--upload')) {
  if (!env.GH_TOKEN && !env.GITHUB_TOKEN) {
    try {
      const result = cp.execFileSync('git',['-c','safe.directory='+path.resolve('.').replaceAll('\\','/'),'credential','fill'],{input:'protocol=https\nhost=github.com\n\n',encoding:'utf8',stdio:['pipe','pipe','pipe'],timeout:20000});
      const password = result.split(/\r?\n/).find(line=>line.startsWith('password=')); if (password) env.GH_TOKEN=password.slice(9);
    } catch {}
  }
  let names;
  try { names=JSON.parse(cp.execFileSync('gh',['secret','list','--repo','string1225/remote-agent-cli','--json','name'],{env,encoding:'utf8',stdio:['ignore','pipe','pipe']})).map(s=>s.name); }
  catch { throw Error('GitHub authentication is unavailable or lacks repository secret access. Signing material has been retained locally.'); }
  const secrets = {ANDROID_KEYSTORE_BASE64:fs.readFileSync(keystore).toString('base64'),ANDROID_KEYSTORE_PASSWORD:config.password,ANDROID_KEY_ALIAS:config.alias,ANDROID_KEY_PASSWORD:config.password};
  if (names.some(name=>name in secrets) && !process.argv.includes('--restore')) throw Error('Android signing secrets already exist. Reuse the original signing material. Pass --restore only with that original backup.');
  for (const [name,value] of Object.entries(secrets)) cp.execFileSync('gh',['secret','set',name,'--repo','string1225/remote-agent-cli'],{env,input:value,stdio:['pipe','ignore','pipe']});
  console.log('Installed the four Android release signing secrets. Back up .local/android-signing securely.');
} else console.log('Release signing identity is ready in .local/android-signing. Back up this directory securely.');
