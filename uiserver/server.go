// httpserver
package uiserver

import (
	"github.com/a526757124/polaris/uiserver/router"
	"github.com/devfeel/middleware/cors"
	"github.com/devfeel/polaris/util/logx"

	"fmt"
	"strconv"

	"github.com/devfeel/dotweb"
)

func StartServer(logPath string) error {
	//初始化DotServer
	dotweb := dotweb.Classic(logPath)
	dotweb.HttpServer.SetEnabledAutoOPTIONS(true)
	dotweb.Use(cors.DefaultMiddleware())
	//初始化Router
	router.InitRoute(dotweb)

	httpPort := 10001

	logger.UIServerLogger.Debug("开始启动监听" + strconv.Itoa(httpPort) + "端口...")
	err := dotweb.StartServer(httpPort)
	if err != nil {
		logger.UIServerLogger.Error("UIServer[" + strconv.Itoa(httpPort) + "]启动失败:" + err.Error())
		fmt.Println("UIServer[" + strconv.Itoa(httpPort) + "]启动失败:" + err.Error())
	}
	return err
}

//注册路由
func InitRoute(dotweb *dotweb.DotWeb) {
	//应用组
	//appGroup := dotweb.HttpServer.Group("/App")
	//dotweb.HttpServer.Any("/App/Add", appInfoHandler.Add)
	//appGroup.OPTIONS("/Add", appInfoHandler.Add)
	//appGroup.POST("/Delete", appInfoHandler.Delete)
	//dotweb.HttpServer.Any("/App/GetList", appInfoHandler.GetList)
	//dotweb.HttpServer.Any("/GetList", appInfoHandler.GetList)
	//appGroup.OPTIONS("/GetList", appInfoHandler.GetList)
}
