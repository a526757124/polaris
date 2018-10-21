package router

import (
	"github.com/a526757124/polaris/uiserver/handlers"
	"github.com/devfeel/dotweb"
)

//应用信息处理
var appInfoHandler *handlers.AppInfoHandler
var apiGroupHandler *handlers.APIGroupHandler
var userHandler *handlers.UserHandler

func init() {
	appInfoHandler = new(handlers.AppInfoHandler)
	apiGroupHandler = new(handlers.APIGroupHandler)
	userHandler = new(handlers.UserHandler)
}

//注册路由
func InitRoute(dotweb *dotweb.DotWeb) {
	//登录
	dotweb.HttpServer.POST("/User/Login", userHandler.Login)
	//退出
	dotweb.HttpServer.POST("/User/LoginOut", userHandler.Login)

	//用户
	dotweb.HttpServer.POST("/User/Add", userHandler.Add)
	dotweb.HttpServer.POST("/User/Update", userHandler.Update)
	dotweb.HttpServer.POST("/User/UpdatePwd", userHandler.UpdatePwd)
	dotweb.HttpServer.POST("/User/Delete", userHandler.Delete)
	dotweb.HttpServer.POST("/User/GetList", userHandler.GetList)

	//应用组
	dotweb.HttpServer.POST("/App/Add", appInfoHandler.Add)
	dotweb.HttpServer.POST("/App/Update", appInfoHandler.Update)
	dotweb.HttpServer.POST("/App/Delete", appInfoHandler.Delete)
	dotweb.HttpServer.POST("/App/GetList", appInfoHandler.GetList)

	//apigroup
	dotweb.HttpServer.POST("/APIGroup/Add", apiGroupHandler.Add)
	dotweb.HttpServer.POST("/APIGroup/Update", apiGroupHandler.Update)
	dotweb.HttpServer.POST("/APIGroup/Delete", apiGroupHandler.Delete)
	dotweb.HttpServer.POST("/APIGroup/GetList", apiGroupHandler.GetList)
}
