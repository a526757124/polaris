package rpcserver

import (
	"net"
	"fmt"
	"net/rpc/jsonrpc"
	"github.com/devfeel/polaris/util/logx"
	. "github.com/devfeel/polaris/gateway/models"
	. "github.com/devfeel/polaris/gateway/const"
	"strconv"
	"github.com/devfeel/polaris/gateway/request"
	"github.com/devfeel/polaris/models"
	"github.com/devfeel/polaris/config"
	"github.com/devfeel/polaris/gateway/balance"
	"sync"
	"github.com/devfeel/polaris/const"
	"encoding/json"
	"net/rpc"
	"github.com/devfeel/polaris/control/metrics"
)

var(
	rpcLogger = logger.JsonRpcLogger
)

type RequestJson struct {
	GateAppID string
	GateEncrypt string
	Module  string
	ApiKey string
	ApiVersion string
	Body []byte
}


func StartServer(){
	port := config.CurrentConfig.Server.JsonRpcPort
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

	proxyLog := &ProxyLog{}
	resJson := &ProxyResponse{
		RetCode : 0,
		RetMsg : "ok",
	}

	apiContext := &ApiContext{
		RetCode:      0,
		RetMsg:       "",
		RealTargetApi: nil,
		ApiInfo:&models.GatewayApiInfo{},
		AppInfo:&models.AppInfo{},
	}

	//resolve api path, from args to api context
	resolveApiPath(&args, apiContext)

	if args.Module == "" || args.ApiKey == "" || args.ApiVersion == "" {
		apiContext.RetMsg = "Not supported Api(QueryPath resolve error) => " + args.String()
		apiContext.RetCode = RetCode_Validate_ResolveApiPathError
	}

	if apiContext.RetCode == RetCode_OK {
		if args.GateAppID == ""{
			apiContext.RetMsg = "unable to resolve query:lost gate_appid"
			apiContext.RetCode = RetCode_Validate_NotExistsAppID
		}
	}

	if apiContext.RetCode == RetCode_OK {
		request.DoValidate(apiContext)
	}
	resJson.LoadConfigTime = config.LoadConfigTime

	//TODO validate md5 sign

	if resJson.RetCode == 0 {
		apiContext.PostBody = args.Body
		if apiContext.ApiInfo.ApiType == _const.ApiType_Balance {
			resJson.RetCode, resJson.RetMsg, apiContext.RealTargetApi = request.DoBalanceTargetApi(apiContext)
			if resJson.RetCode ==0 {
				body, contentType, intervalTime, err := request.DoRequestTarget(apiContext)
				if err != nil {
					resJson.RetCode = -209999
					resJson.RetMsg = body
					resJson.Message = err.Error()
					balance.SetError(apiContext.ApiInfo, apiContext.RealTargetApi.TargetKey)
				} else {
					resJson.RetCode = 0
					resJson.RetMsg = "ok"
					resJson.Message = body
				}
				resJson.IntervalTime = intervalTime
				resJson.ContentType = contentType
				proxyLog.CallInfo = []*models.TargetApiInfo{apiContext.RealTargetApi}
			}
		}
		if apiContext.ApiInfo.ApiType == _const.ApiType_Group{
			var syncWait sync.WaitGroup
			var targetResults []*models.TargetApiResult
			for _, v:=range apiContext.ApiInfo.TargetApis{
				syncWait.Add(1)
				go func() {
					defer syncWait.Done()
					result := new(models.TargetApiResult)
					result.ApiKey = v.TargetKey
					body, _, intervalTime, err := request.DoRequestTarget(apiContext)
					if err != nil {
						result.RetCode = RetCode_Error
						result.RetMsg = err.Error()
					} else {
						errJson := json.Unmarshal([]byte(body), result)
						if errJson != nil {
							result.RetCode = RetCode_JsonUnmarshalError
							result.RetMsg = errJson.Error()
						}
					}
					result.IntervalTime = intervalTime
					targetResults = append(targetResults, result)
				}()
			}
			syncWait.Wait()
			resJson.RetCode = RetCode_OK
			resJson.RetMsg = "ok"
			resJson.Message = targetResults
			resJson.IntervalTime = 0
			resJson.ContentType = ""
			proxyLog.CallInfo = apiContext.ApiInfo.TargetApis
		}
	}

	//not support apiContext.ApiInfo.RawResponseFlag

	//proxy log set
	proxyLog.RawResponseFlag = apiContext.ApiInfo.RawResponseFlag
	convertProxyLog(proxyLog, resJson)

	jsonLogB, _ := json.Marshal(proxyLog)
	rpcLogger.Info(string(jsonLogB))
	//do metrics
	metrics.AddApiCount(apiContext.GateAppID, apiContext.ApiInfo.ApiID, apiContext.ApiModule, apiContext.ApiName, apiContext.ApiVersion, 1, strconv.Itoa(resJson.RetCode))

	result = resJson

	return nil
}


func (self *RequestJson) String() string {
	str, _ := json.Marshal(self)
	if str == nil{
		return ""
	}
	return string(str)
}

// resolveApiPath resolve api path, from args to api context
// update: ApiModule、ApiKey、ApiVersion、ApiUrlKey
func resolveApiPath(args *RequestJson, apiContext *ApiContext) {
	apiContext.ApiModule = args.Module
	apiContext.ApiName = args.ApiKey
	apiContext.ApiVersion = args.ApiVersion
	apiContext.ApiUrlKey = args.Module + "/" + args.ApiKey + "/" + args.ApiVersion
}

// convertProxyLog convert response to proxy log
func convertProxyLog(log *ProxyLog, res *ProxyResponse){
	log.RetCode = res.RetCode
	log.RetMsg = res.RetMsg
	log.IntervalTime = res.IntervalTime
	log.LoadConfigTime = res.LoadConfigTime
	log.ContentType = res.ContentType
	log.Message = res.Message
}
