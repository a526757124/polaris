// httpserver
package httpserver

import (
	"github.com/devfeel/polaris/config"
	"github.com/devfeel/polaris/util/logx"
	"TechPlat/apigateway/httpserver/route"

	"fmt"
	"strconv"

	"github.com/devfeel/dotweb"
)

func StartServer(logPath string) error {
	//初始化DotServer
	dotweb := dotweb.Classic(logPath)

	//初始化Router
	httproute.InitRoute(dotweb)

	httpPort := config.CurrentConfig.HttpServer.HttpPort

	logger.DefaultLogger.Debug("开始启动监听"+strconv.Itoa(httpPort)+"端口...")
	err := dotweb.StartServer(httpPort)
	if err != nil {
		logger.DefaultLogger.Error("HttpServer["+strconv.Itoa(httpPort)+"]启动失败:"+err.Error())
		fmt.Println("HttpServer[" + strconv.Itoa(httpPort) + "]启动失败:" + err.Error())
	}
	return err
}
