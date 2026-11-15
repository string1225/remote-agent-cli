using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.ComponentModel;
using System.IO;
using System.Security.Principal;
using System.Text;
using System.Web.Script.Serialization;
using Microsoft.Deployment.WindowsInstaller;

// Immediate UI actions run as the installing user. Only the existing service
// configuration helper elevates; tokens travel in stdin, never MSI action data.
public static class ActivationActions {
    static readonly JavaScriptSerializer Json = new JavaScriptSerializer();
    static bool Flag(IDictionary<string,object> row,string key) {return row.ContainsKey(key) && row[key] is bool && (bool)row[key];}
    static string Text(IDictionary<string,object> row,string key) {return row.ContainsKey(key) ? Convert.ToString(row[key]) : "";}
    static void ShowStatus(Session session,string message) {
        session["ACTIVATION_MESSAGE"]=message;
        using(var record=new Record(3)) {
            record[1]="DeviceActivation";record[2]=message;record[3]="";
            session.Message(InstallMessage.ActionStart,record);
        }
    }
    static IDictionary<string,object> Call(Session session,string args,object request) {
        if (WindowsIdentity.GetCurrent().User.Value != session["UserSID"])
            throw new InvalidOperationException("请从目标 Windows 用户的桌面运行安装程序，避免配置到其他用户目录。");
        return RunProcess(session.Format("[#SeedCore]"),args,request,delegate(string message) {ShowStatus(session,message);},session["UserSID"]);
    }
    public static IDictionary<string,object> RunProcess(string executable,string args,object request,Action<string> progress,string owner=null) {
        var start = new ProcessStartInfo(executable,args) {
            UseShellExecute=false,CreateNoWindow=true,RedirectStandardInput=true,
            RedirectStandardOutput=true,RedirectStandardError=true,
            StandardOutputEncoding=Encoding.UTF8,StandardErrorEncoding=Encoding.UTF8
        };
        if(owner!=null)start.EnvironmentVariables["REMOTE_AGENT_ACTIVATION_OWNER"]=owner;
        using (var process = Process.Start(start)) {
            // Never copy stderr into MSI logs: arbitrary dependencies may print
            // credentials. The structured result is the only display channel.
            process.ErrorDataReceived += delegate {};
            process.BeginErrorReadLine();
            if (request != null) {
                var bytes=Encoding.UTF8.GetBytes(Json.Serialize(request));
                process.StandardInput.BaseStream.Write(bytes,0,bytes.Length);
                process.StandardInput.BaseStream.Flush();
            }
            process.StandardInput.Close();
            IDictionary<string,object> result=null;
            string line;
            while ((line=process.StandardOutput.ReadLine())!=null) {
                IDictionary<string,object> item;
                try {item=Json.Deserialize<Dictionary<string,object>>(line);} catch {continue;}
                if(item==null)continue;
                if (Text(item,"type")=="progress") {
                    progress(Text(item,"message"));
                } else if(Text(item,"type")=="result" || Text(item,"type")=="inspect") result=item;
            }
            process.WaitForExit();
            if (process.ExitCode!=0) throw new InvalidOperationException("激活程序退出（退出码 "+process.ExitCode+"）。安装和原绑定已保留，请在新终端使用 CLI 重试。");
            if (result==null) throw new InvalidOperationException("激活进程未返回结果（退出码 "+process.ExitCode+"）。安装已保留，可在新终端运行 remote-agent connect --token-prompt 或 remote-agent codex setup --server https://codex.sunny-string.cn/agents --token-prompt 重试。");
            return result;
        }
    }
    [CustomAction]
    public static ActionResult InitializeActivation(Session session) {
        try {
            var result=Call(session,"activation --inspect",null);
            if(Text(result,"type")=="result") {session["ACTIVATION_MESSAGE"]=Text(result,"message");return ActionResult.Success;}
            session["AGENT_BOUND"]=Flag(result,"bound")?"1":"";
            session["ACTIVATION_MESSAGE"]=Flag(result,"bound")?"已保留设备「"+Text(result,"name")+"」的绑定。可直接完成，或调整下方选项后应用。":"填写设备绑定码或指引中的 Endpoint Token，再点击激活。";
            var options=(IDictionary<string,object>)result["options"];
            if (String.IsNullOrEmpty(session["BINDING_CODE"])) session["AGENT_SERVER"]=Text(options,"server");
            session["AUTO_START"]=Flag(options,"autoStart")?"1":"";
            session["WEB_PROXY"]=Flag(options,"webProxy")?"1":"";
            session["CONFIGURE_CODEX"]=Flag(options,"configureCodex")?"1":"";
            session["INSTALL_CODEX"]=Flag(options,"installCodex")?"1":"";
            session["INSTALL_CODEX_APP"]="";
        } catch (Exception) {session["ACTIVATION_MESSAGE"]="安装已完成。激活状态暂时无法读取，可以填写绑定码后重试。";}
        return ActionResult.Success;
    }
    [CustomAction]
    public static ActionResult ActivateDevice(Session session) {
        session["ACTIVATION_OK"]="";
        try {
            string code=session["BINDING_CODE"].Trim();
            bool personal=code.StartsWith("ra_codex_",StringComparison.Ordinal);
            var request=new Dictionary<string,object> {
                {"server",session["AGENT_SERVER"]},{"bindingCode",personal?"":code},{"apiToken",personal?code:""},
                {"autoStart",session["AUTO_START"]=="1"},{"webProxy",session["WEB_PROXY"]=="1"},
                {"configureCodex",session["CONFIGURE_CODEX"]=="1"},{"installCodex",session["INSTALL_CODEX"]=="1"},
                {"installApp",session["INSTALL_CODEX_APP"]=="1"}
            };
            var result=Call(session,"activation --request-stdin",request);
            session["ACTIVATION_OK"]=Flag(result,"success")?"1":"";
            if(!personal) session["AGENT_BOUND"]=Flag(result,"bound")?"1":"";
            ShowStatus(session,Text(result,"message"));
            if (Flag(result,"bound") || Flag(result,"success")) session["BINDING_CODE"]="";
        } catch(Exception e) {
            var win32=e as Win32Exception;
            ShowStatus(session,win32!=null?"无法启动激活程序（Windows 错误 "+win32.NativeErrorCode+"）。安装已保留，请在新终端使用 CLI 重试。":e is InvalidOperationException?e.Message:"激活暂未完成（"+e.GetType().Name+"）。安装和已有绑定已保留，请检查网络或授权后重试。");
        }
        return ActionResult.Success;
    }
}
