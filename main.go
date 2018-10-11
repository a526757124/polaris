// main
package main

import (
	_ "net/http/pprof"
	"os"

	"github.com/devfeel/polaris/const"
	"github.com/devfeel/polaris/util/logx"

	"github.com/a526757124/polaris/core/exception"
	"github.com/a526757124/polaris/uiserver"
	"github.com/a526757124/polaris/util/filex"
)

var (
	RunEnv string
)

const (
	RunEnv_Flag       = "RunEnv"
	RunEnv_Develop    = "develop"
	RunEnv_Test       = "test"
	RunEnv_Production = "production"
	RunEnv_UAT        = "uat"
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			ex := exception.CatchError(_const.ProjectName+":main", err)
			logger.DefaultLogger.Error(ex.GetDefaultLogString())
			os.Stdout.Write([]byte(ex.GetDefaultLogString()))
		}
	}()

	// RunEnv = os.Getenv(RunEnv_Flag)
	// if RunEnv == "" {
	// 	RunEnv = RunEnv_Develop
	// }

	currentBaseDir := filex.GetCurrentDirectory()
	//var configFile string
	//flag.StringVar(&configFile, "config", "", "配置文件路径")
	//for docker
	//if configFile == "" {
	//	configFile = currentBaseDir + "/conf/" + RunEnv + "/app.conf"
	//}
	//for local run
	//if configFile == "" {
	//	configFile = currentBaseDir + "/gateway.conf"
	//}

	//加载xml配置文件
	//config.InitConfig(configFile)

	//设置基本目录
	//config.SetBaseDir(currentBaseDir)

	//启动Api计数日志
	//metrics.StartApiCountHandler()
	err := uiserver.StartServer(currentBaseDir)
	if err != nil {
		//fmt.Println(err.Error())
		logger.DefaultLogger.Debug("HttpServer.StartServer失败 " + err.Error())
	}
}
