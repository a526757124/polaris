// httpHandler
package handlers

import (
	"strconv"
	"strings"
	"time"
	"os"
	"fmt"
	"sync"
	"encoding/json"
	"errors"

	"github.com/devfeel/dotweb"
	"github.com/devfeel/polaris/util/httpx"
	"github.com/devfeel/polaris/config"
	"github.com/devfeel/polaris/const"
	"github.com/devfeel/polaris/models"
	"github.com/devfeel/polaris/util/logx"
	"github.com/devfeel/polaris/gateway/balance"
	"github.com/devfeel/polaris/control/metric"
	"github.com/devfeel/polaris/core/exception"
	"github.com/devfeel/polaris/gateway/httpserver/servermodel"
	"github.com/devfeel/polaris/util/rpcx"
)

type ResponseJson struct {
	RetCode          int
	RetMsg           string
	LoadConfigTime 		 time.Time
	IntervalTime     int64
	ContentType      string
	Message          interface{}
}

type ProxyLog struct {
	RetCode           int
	RetMsg            string
	RequestUrl        string
	CallInfo 	  	  []*models.TargetApiInfo
	RawResponseFlag	  bool
	LoadConfigTime  time.Time
	IntervalTime      int64
	ContentType       string
	Message           interface{}
}

var(
	gatewayLogger = logger.GatewayLogger
)

// convertProxyLog convert response to proxy log
func convertProxyLog(log *ProxyLog, res *ResponseJson){
	log.RetCode = res.RetCode
	log.RetMsg = res.RetMsg
	log.IntervalTime = res.IntervalTime
	log.LoadConfigTime = res.LoadConfigTime
	log.ContentType = res.ContentType
	log.Message = res.Message
}

// combineApiUrl根据api地址与查询参数，组合实际访问地址
func combineApiUrl(targetApiUrl, queryString string) string{
	//处理参数拼接，考虑是否匹配?和&符号的情况
	if strings.Contains(targetApiUrl, "?") {
		if !strings.HasSuffix(targetApiUrl, "&") && !strings.HasSuffix(targetApiUrl, "?") {
			targetApiUrl = targetApiUrl + "&" + queryString
		} else {
			targetApiUrl = targetApiUrl + queryString
		}

	} else {
		targetApiUrl = targetApiUrl + "?" + queryString
	}
	return targetApiUrl
}

// getApiContextFromItem get ApiContext from item which is created in validate middleware
func getApiContextFromItem(ctx dotweb.Context)(*servermodel.ApiContext){
	var apiContext *servermodel.ApiContext
	//get result from validate middleware
	itemContext, isExists := ctx.Items().Get(_const.HttpContext_ApiContext)
	if !isExists{
		apiContext = &servermodel.ApiContext{
			RetCode:      _const.RetCode_GetApiContextError,
			RetMsg:       "get ApiContex error, no exists item key",
			ApiInfo:&models.GatewayApiInfo{},
			AppInfo:&models.AppInfo{},
		}
	}else{
		var isOk bool
		apiContext, isOk = itemContext.(*servermodel.ApiContext)
		if !isOk || apiContext == nil {
			apiContext = &servermodel.ApiContext{
				RetCode:      _const.RetCode_GetApiContextError,
				RetMsg:       "get ApiContex error, not match type",
				ApiInfo:      &models.GatewayApiInfo{},
				AppInfo:      &models.AppInfo{},
			}
		}
	}
	return apiContext
}

// doBalanceTargetApi do balance real target apiurl
// if not exists alive target, return error info
func doBalanceTargetApi(apiContext *servermodel.ApiContext) (retCode int, retMsg string, realTargetApi *models.TargetApiInfo){
	retCode = _const.RetCode_OK

	if apiContext.ApiInfo.ApiType != _const.ApiType_Balance {
		retCode = -100010
		retMsg = "get targetapi failed, not balance mode!"
		return
	}
	//获取本次请求处理的目标Api，加入负载机制
	if apiContext.ApiInfo.TargetApi != nil && len(apiContext.ApiInfo.TargetApi) >0 {
		targetApi := balance.GetAliveApi(apiContext.ApiInfo)
		if targetApi == nil {
			retCode = -100010
			retMsg = "get targetapi failed, load targetapi nil!"
			return
		} else{
			//组合API地址与参数
			realTargetApi = targetApi
		}
	}else{
		retCode = -100010
		retMsg = "get targetapi failed, load targetapi nil!"
		return
	}
	return
}

// doRequestTarget do request match HttpGet\HttpPost\JsonRpc
func doRequestTarget(apiCtx *servermodel.ApiContext)(body, contentType string, intervalTime int64, err error){
	if apiCtx.RealTargetApi.CallName == _const.CallMethod_HttpGet{
		realApiUrl := combineApiUrl(apiCtx.RealTargetApi.TargetUrl, apiCtx.Query)
		body, contentType, intervalTime, err = httpx.HttpGet(realApiUrl)
	}else if apiCtx.RealTargetApi.CallName == _const.CallMethod_HttpPost {
		realApiUrl := combineApiUrl(apiCtx.RealTargetApi.TargetUrl, apiCtx.Query)
		body, contentType, intervalTime, err = httpx.HttpPost(realApiUrl, string(apiCtx.PostBody), apiCtx.ContentType)
	}else if apiCtx.RealTargetApi.CallName == _const.CallMethod_JsonRPC {
		var rpcBody []byte
		rpcBody, intervalTime, err = rpcx.CallJsonRPC(apiCtx.RealTargetApi.TargetUrl, apiCtx.RealTargetApi.CallName, apiCtx.PostBody)
		if err != nil{
			body = string(rpcBody)
		}
		contentType = apiCtx.ContentType
	}else{
		return "", "", 0, errors.New("no match Call Method")
	}
	return
}

// OneProxy route all Get/Post/JsonRpc requests to real target server
// returns: ResponseJson: RetCode,RetMsg,LastConfigTime,IntervalTime,ContentType,Message
func OneProxy(ctx dotweb.Context) error{
	defer func() {
		if err := recover(); err != nil {
			ex := exception.CatchError(_const.ProjectName+":ProxyHttp", err)
			gatewayLogger.Error(ex.GetDefaultLogString())
			os.Stdout.Write([]byte(ex.GetDefaultLogString()))
		}
	}()

	proxyLog := &ProxyLog{
		RequestUrl:	string(ctx.Request().Url()),
	}
	resJson := &ResponseJson{
		RetCode : 0,
		RetMsg : "ok",
	}
	var apiContext *servermodel.ApiContext

	//get api context from item
	apiContext = getApiContextFromItem(ctx)
	resJson.RetCode = apiContext.RetCode
	resJson.RetMsg = apiContext.RetMsg
	resJson.LoadConfigTime = config.LoadConfigTime

	if resJson.RetCode == 0 {
		apiContext.PostBody = ctx.Request().PostBody()
		if apiContext.ApiInfo.ApiType == _const.ApiType_Balance {
			resJson.RetCode, resJson.RetMsg, apiContext.RealTargetApi = doBalanceTargetApi(apiContext)
			if resJson.RetCode ==0 {
				body, contentType, intervalTime, err := doRequestTarget(apiContext)
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
			for _, v:=range apiContext.ApiInfo.TargetApi{
				syncWait.Add(1)
				go func() {
					defer syncWait.Done()
					result := new(models.TargetApiResult)
					result.ApiKey = v.TargetKey
					body, _, intervalTime, err := doRequestTarget(apiContext)
					if err != nil {
						result.RetCode = _const.RetCode_Error
						result.RetMsg = err.Error()
					} else {
						errJson := json.Unmarshal([]byte(body), result)
						if errJson != nil {
							result.RetCode = _const.RetCode_JsonUnmarshalError
							result.RetMsg = errJson.Error()
						}
					}
					result.IntervalTime = intervalTime
					targetResults = append(targetResults, result)
				}()
			}
			syncWait.Wait()
			resJson.RetCode = _const.RetCode_OK
			resJson.RetMsg = "ok"
			resJson.Message = targetResults
			resJson.IntervalTime = 0
			resJson.ContentType = ""
			proxyLog.CallInfo = apiContext.ApiInfo.TargetApi
		}
	}

	responseContent := ""
	if resJson.RetCode == 0 {
		if apiContext.ApiInfo.RawResponseFlag {
			if resJson.RetCode == _const.RetCode_OK {
				responseContent = fmt.Sprint(resJson.Message)
			}
		}
	}

	if responseContent == "" {
		jsonb, err := json.Marshal(resJson)
		//解析异常处理
		if err != nil {
			responseContent = `{"RetCode":-109999,"RetMsg":"json marshal error - ` + err.Error() + `","Message":"` + resJson.Message.(string) + `}`
		} else {
			responseContent = string(jsonb)
		}
	}

	//proxy log set
	convertProxyLog(proxyLog, resJson)

	if apiContext.ApiInfo!= nil{
		proxyLog.RawResponseFlag = apiContext.ApiInfo.RawResponseFlag
	}else{
		proxyLog.RawResponseFlag = false
	}

	jsonLogB, _ := json.Marshal(proxyLog)
	gatewayLogger.Info(string(jsonLogB))
	//do metrics
	apiId := 0
	if apiContext.ApiInfo!= nil{
		apiId = apiContext.ApiInfo.ApiID
	}
	metric.AddApiCount(apiContext.GateAppID, apiId, apiContext.ApiModule, apiContext.ApiName, apiContext.ApiVersion, 1, strconv.Itoa(resJson.RetCode))
	ctx.WriteString(responseContent)
	return nil
}
