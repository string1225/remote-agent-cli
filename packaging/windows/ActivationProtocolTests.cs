using System;
using System.Collections.Generic;

// Pure child-process protocol test. No MSI execution, service, binding, user
// configuration, system proxy, or network operation is performed.
public static class ActivationProtocolTests {
    public static void Main(string[] args) {
        try { Run(args); } catch(Exception e) {Console.WriteLine("FAIL: "+e.GetType().Name+" "+e.Message);Environment.Exit(1);}
    }
    static void Run(string[] args) {
        int progress=0;
        var request=new Dictionary<string,object>{{"bindingCode","fixture ' 中文 code"},{"autoStart",false},{"installApp",true}};
        var result=ActivationActions.RunProcess(args[0],"fixture",request,delegate(string message){if(message!="准备中")throw new Exception("Unicode progress lost");progress++;});
        if(progress!=1 || !(bool)result["success"])throw new Exception("Activation protocol failed");
        foreach(string mode in new[]{"missing","failed"}) {
            bool rejected=false;
            try {ActivationActions.RunProcess(args[0],mode,null,delegate(string message){});}
            catch(InvalidOperationException e){
                if(e.Message.Contains("fixture-secret"))throw new Exception("Dependency output exposed");
                if(mode=="failed" && !e.Message.Contains("23"))throw new Exception("Exit code lost");
                rejected=true;
            }
            if(!rejected)throw new Exception("Invalid activation result accepted: "+mode);
        }
        Console.WriteLine("PASS: hidden process stdin, Unicode progress/result, stderr draining; no installation or service changes");
    }
}
