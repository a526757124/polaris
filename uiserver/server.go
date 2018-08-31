// httpserver
package httpserver

import (
	"github.com/devfeel/polaris/config"
	"github.com/devfeel/polaris/util/logx"

	"fmt"
	"strconv"

	"github.com/devfeel/dotweb"
)

func StartServer(logPath string) error {
	//初始化DotServer
	dotweb := dotweb.Classic(logPath)

	//初始化Router
	InitRoute(dotweb)

	httpPort := config.CurrentConfig.Server.UIPort

	logger.UIServerLogger.Debug("开始启动监听"+strconv.Itoa(httpPort)+"端口...")
	err := dotweb.StartServer(httpPort)
	if err != nil {
		logger.UIServerLogger.Error("UIServer["+strconv.Itoa(httpPort)+"]启动失败:"+err.Error())
		fmt.Println("UIServer[" + strconv.Itoa(httpPort) + "]启动失败:" + err.Error())
	}
	return err
}

func InitRoute(dotweb *dotweb.DotWeb) {

}
