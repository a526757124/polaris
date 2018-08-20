// httpserver
package httpserver

import (
	"github.com/devfeel/polaris/config"
	"github.com/devfeel/polaris/util/logx"

	"fmt"
	"strconv"

	"github.com/devfeel/dotweb"
	"github.com/devfeel/polaris/gateway/httpserver/handlers"
	"github.com/devfeel/polaris/gateway/httpserver/middlewares"
)

func StartServer(logPath string) error {
	//初始化DotServer
	dotweb := dotweb.Classic(logPath)

	//初始化Router
	InitRoute(dotweb)

	httpPort := config.CurrentConfig.HttpServer.HttpPort

	logger.DefaultLogger.Debug("开始启动监听"+strconv.Itoa(httpPort)+"端口...")
	err := dotweb.StartServer(httpPort)
	if err != nil {
		logger.DefaultLogger.Error("HttpServer["+strconv.Itoa(httpPort)+"]启动失败:"+err.Error())
		fmt.Println("HttpServer[" + strconv.Itoa(httpPort) + "]启动失败:" + err.Error())
	}
	return err
}

func InitRoute(dotweb *dotweb.DotWeb) {
	dotweb.HttpServer.Router().GET("/api/:module/:version/:apikey", handlers.OneProxy).Use(middlewares.NewValidateMiddleware())
	dotweb.HttpServer.Router().POST("/api/:module/:version/:apikey", handlers.OneProxy).Use(middlewares.NewValidateMiddleware())
	dotweb.HttpServer.Router().GET("/", handlers.Index)
	dotweb.HttpServer.Router().GET("/monitor", handlers.Monitor)
	dotweb.HttpServer.Router().GET("/pprof/:module", handlers.Watch)
	dotweb.HttpServer.Router().GET("/info/:infotype", handlers.Info)
	dotweb.HttpServer.Router().GET("/version", handlers.Version)
}
