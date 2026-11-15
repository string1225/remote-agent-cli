using System;
using System.Collections.Generic;
using System.Text;
using System.Web.Script.Serialization;
public static class ActivationProtocolFixture {
    public static void Main(string[] args) {
        Console.InputEncoding=Encoding.UTF8;Console.OutputEncoding=new UTF8Encoding(false);
        if(args.Length>0 && args[0]=="missing"){Console.WriteLine("{\"unexpected\":\"fixture-secret\"}");return;}
        if(args.Length>0 && args[0]=="failed"){Console.Error.WriteLine("fixture-secret");Environment.Exit(23);}
        var row=new JavaScriptSerializer().Deserialize<Dictionary<string,object>>(Console.In.ReadToEnd());
        if((string)row["bindingCode"]!="fixture ' 中文 code" || (bool)row["autoStart"] || !(bool)row["installApp"])Environment.Exit(2);
        Console.Error.Write(new string('x',100000));
        Console.WriteLine("{\"type\":\"progress\",\"message\":\"准备中\"}");
        Console.WriteLine("{\"type\":\"result\",\"success\":true}");
        Console.WriteLine("null");
        Console.WriteLine("{\"unexpected\":\"fixture-secret\"}");
    }
}
