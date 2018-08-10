// monitorHandler
package handlers

import (
	"TechPlat/apigateway/config"
	"TechPlat/apigateway/httpserver/monitor"
	"encoding/json"
	"runtime/pprof"
	"html/template"

	"github.com/devfeel/dotweb"
	"TechPlat/apigateway/const"
	"github.com/devfeel/polaris/util/httpx"
)

func Monitor(ctx dotweb.Context) error{
	bytes, _ := json.Marshal(monitor.Current)
	ctx.WriteString(string(bytes))
	return nil
}

func Watch(ctx dotweb.Context) error {
	ctx.Response().SetContentType(dotweb.MIMETextPlain)
	module := ctx.GetRouterName("module")
	p := pprof.Lookup("goroutine")
	if module != "" {
		p = pprof.Lookup(module)
	}
	if p != nil {
		p.WriteTo(ctx.Response().Writer(), 1)
	} else {
		ctx.WriteString("no such pprofmodule")
	}
	return nil
}


func Index(ctx dotweb.Context) error {
	ctx.WriteString("welcome to " + _const.ProjectName + " [" + _const.ProjectVersion+"]!")
	return nil
}

func Info(ctx dotweb.Context) error{
	ctx.Response().SetContentType(dotweb.MIMETextHTML)

	ctx.WriteString("LastLoadApisTime : " + config.LastLoadApisTime.Format(_const.DefaultTimeLayout))
	ctx.WriteString("<br>")
	ctx.WriteString("<br>")

	infoType := ctx.GetRouterName("infotype")
	if infoType == "app" {
		for _, app := range config.GetAppList() {
			app.AppKey = ""
			jsons, _ := json.Marshal(app)
			ctx.WriteString(string(jsons) + "<br>")
		}
	}
	if infoType == "api" {
		for _, api := range config.GetApiList() {
			jsons, _ := json.Marshal(api)
			ctx.WriteString(string(jsons) + "<br>")
		}
	}
	if infoType == "relation" {
		for _, relation := range config.GetRelationList() {
			jsons, _ := json.Marshal(relation)
			ctx.WriteString(string(jsons) + "<br>")
		}
	}
	return nil
}

func Version(ctx dotweb.Context) error{
	filePath, errFile := httpx.GetCurrentDirectory()
	if errFile != nil {
		ctx.WriteString("version template file error => " + errFile.Error())
		return nil
	}
	filePath = filePath + "/views/version.html"
	tmpl, err := template.New("version.html").ParseFiles(filePath)
	if err != nil {
		ctx.WriteString("version template Parse error => " + err.Error())
		return nil
	}
	viewdata := make(map[string]string)
	viewdata["version"] = _const.ProjectVersion
	viewdata["versiondesc"] = "启用RawResponseFlag判断，若正常调用，直接返回原始内容，若非正常调用，按照API网关统一返回类型返回"
	ctx.Response().SetContentType("text/html; charset=utf8")
	err = tmpl.Execute(ctx.Response().Writer(), viewdata)
	if err != nil {
		ctx.WriteString("version template Execute error => " + err.Error())
		return nil
	}

	return nil
}
