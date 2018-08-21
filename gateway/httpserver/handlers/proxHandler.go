// httpHandler
package handlers

import (
	"strconv"
	"os"
	"fmt"
	"sync"
	"encoding/json"
	"github.com/devfeel/dotweb"
	"github.com/devfeel/polaris/config"
	"github.com/devfeel/polaris/const"
	. "github.com/devfeel/polaris/gateway/const"
	"github.com/devfeel/polaris/models"
	"github.com/devfeel/polaris/util/logx"
	"github.com/devfeel/polaris/gateway/balance"
	"github.com/devfeel/polaris/control/metrics"
	"github.com/devfeel/polaris/core/exception"
	. "github.com/devfeel/polaris/gateway/models"
	"github.com/devfeel/polaris/gateway/request"
)




var(
	gatewayLogger = logger.GatewayLogger
)

// convertProxyLog convert response to proxy log
func convertProxyLog(log *ProxyLog, res *ProxyResponse){
	log.RetCode = res.RetCode
	log.RetMsg = res.RetMsg
	log.IntervalTime = res.IntervalTime
	log.LoadConfigTime = res.LoadConfigTime
	log.ContentType = res.ContentType
	log.Message = res.Message
}

// getApiContextFromItem get ApiContext from item which is created in validate middleware
func getApiContextFromItem(ctx dotweb.Context)(*ApiContext){
	var apiContext *ApiContext
	//get result from validate middleware
	itemContext, isExists := ctx.Items().Get(HttpContext_ApiContext)
	if !isExists{
		apiContext = &ApiContext{
			RetCode:      RetCode_GetApiContextError,
			RetMsg:       "get ApiContex error, no exists item key",
			ApiInfo:&models.GatewayApiInfo{},
			AppInfo:&models.AppInfo{},
		}
	}else{
		var isOk bool
		apiContext, isOk = itemContext.(*ApiContext)
		if !isOk || apiContext == nil {
			apiContext = &ApiContext{
				RetCode:      RetCode_GetApiContextError,
				RetMsg:       "get ApiContex error, not match type",
				ApiInfo:      &models.GatewayApiInfo{},
				AppInfo:      &models.AppInfo{},
			}
		}
	}
	return apiContext
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
	resJson := &ProxyResponse{
		RetCode : 0,
		RetMsg : "ok",
	}
	var apiContext *ApiContext

	//get api context from item
	apiContext = getApiContextFromItem(ctx)
	resJson.RetCode = apiContext.RetCode
	resJson.RetMsg = apiContext.RetMsg
	resJson.LoadConfigTime = config.LoadConfigTime

	if resJson.RetCode == 0 {
		apiContext.PostBody = ctx.Request().PostBody()
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

	responseContent := ""
	if resJson.RetCode == 0 {
		if apiContext.ApiInfo.RawResponseFlag {
			if resJson.RetCode == RetCode_OK {
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
	proxyLog.RawResponseFlag = apiContext.ApiInfo.RawResponseFlag
	convertProxyLog(proxyLog, resJson)

	jsonLogB, _ := json.Marshal(proxyLog)
	gatewayLogger.Info(string(jsonLogB))
	//do metrics
	metrics.AddApiCount(apiContext.GateAppID, apiContext.ApiInfo.ApiID, apiContext.ApiModule, apiContext.ApiName, apiContext.ApiVersion, 1, strconv.Itoa(resJson.RetCode))
	ctx.WriteString(responseContent)
	return nil
}
