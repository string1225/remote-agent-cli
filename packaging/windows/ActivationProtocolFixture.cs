using System;
using System.Collections.Generic;
using System.Text;
using System.Web.Script.Serialization;
public static class ActivationProtocolFixture {
    public static void Main() {
        Console.InputEncoding=Encoding.UTF8;Console.OutputEncoding=new UTF8Encoding(false);
        var row=new JavaScriptSerializer().Deserialize<Dictionary<string,object>>(Console.In.ReadToEnd());
        if((string)row["bindingCode"]!="fixture ' 中文 code" || (bool)row["autoStart"])Environment.Exit(2);
        Console.Error.Write(new string('x',100000));
        Console.WriteLine("{\"type\":\"progress\",\"message\":\"准备中\"}");
        Console.WriteLine("{\"type\":\"result\",\"success\":true}");
    }
}
