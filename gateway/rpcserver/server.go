package rpcserver

import (
	"net"
	"fmt"
	"net/rpc/jsonrpc"
	"net/rpc"
	"github.com/devfeel/polaris/util/logx"
	. "github.com/devfeel/polaris/gateway/models"
	. "github.com/devfeel/polaris/gateway/const"
	"strconv"
	"github.com/devfeel/polaris/gateway/request"
	"github.com/devfeel/polaris/models"
	"github.com/gin-gonic/gin/json"
)

var(
	gatewayLogger = logger.GatewayLogger
	port = 1789
)

type RequestJson struct {
	GateAppID string
	GateEncrypt string
	Module  string
	ApiKey string
	ApiVersion string
	Body interface{}
}


func StartServer(){
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return
	}
	defer lis.Close()

	srv := rpc.NewServer()
	if err := srv.RegisterName("Request", new(RequestJson)); err != nil {
		logger.DefaultLogger.Error("RpcServer["+strconv.Itoa(port)+"] start failed:"+err.Error())
		fmt.Println("RpcServer[" + strconv.Itoa(port) + "] start failed:" + err.Error())
		return
	}

	logger.DefaultLogger.Debug("RpcServer[" + strconv.Itoa(port) + "] starting...")

	for {
		conn, err := lis.Accept()
		if err != nil {
			logger.DefaultLogger.Error("RpcServer["+strconv.Itoa(port)+"] Accept failed:"+err.Error())
			fmt.Println("RpcServer[" + strconv.Itoa(port) + "] Accept failed:" + err.Error())
			continue
		}
		go srv.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}


func (self *RequestJson) Request(args RequestJson, result *ProxyResponse) error {
	apiContext := &ApiContext{
		RetCode:      0,
		RetMsg:       "",
		RealTargetApi: nil,
		ApiInfo:&models.GatewayApiInfo{},
		AppInfo:&models.AppInfo{},
	}

	apiContext.ApiModule = args.Module
	apiContext.ApiName = args.ApiKey
	apiContext.ApiVersion = args.ApiVersion
	apiContext.ApiUrlKey = args.Module + "/" + args.ApiKey + "/" + args.ApiVersion
	apiContext.RemoteIP = ""

	//解析api请求目录
	if args.Module == "" || args.ApiKey == "" || args.ApiVersion == "" {
		apiContext.RetMsg = "Not supported Api(QueryPath resolve error) => " + args.String()
		apiContext.RetCode = RetCode_Validate_ResolveApiPathError
	}


	if apiContext.RetCode == RetCode_OK {
		if args.GateAppID == "" {
			apiContext.RetMsg = "unable to resolve query:lost gate_appid"
			apiContext.RetCode = RetCode_Validate_NotExistsAppID
		}
	}

	if apiContext.RetCode == RetCode_OK {
		request.DoValidate(apiContext)
	}

	//TODO validate md5 sign

	return nil
}


func (self *RequestJson) String() string {
	str, _ := json.Marshal(self)
	if str == nil{
		return ""
	}
	return string(str)
}


