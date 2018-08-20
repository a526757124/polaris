// httpHandler
package handlers

import (
	"github.com/devfeel/polaris/config"
	"github.com/devfeel/polaris/const"
	"github.com/devfeel/polaris/models"
	"github.com/devfeel/polaris/util/logx"
	"github.com/devfeel/polaris/gateway/balance"
	"strconv"
	"strings"
	"time"
	"github.com/devfeel/dotweb"
	"os"
	"fmt"
	"github.com/devfeel/polaris/util/httpx"
	"sync"
	"encoding/json"
	"github.com/devfeel/polaris/control/metric"
	"github.com/devfeel/polaris/core/exception"
	"github.com/devfeel/polaris/gateway/httpserver/servermodel"
)

type ResponseJson struct {
	RetCode          int
	RetMsg           string
	LastLoadApisTime time.Time
	IntervalTime     int64
	ContentType      string
	Message          interface{}
}

type LogJson struct {
	RetCode           int
	RetMsg            string
	RequestUrl        string
	HttpMethod        string
	RawResponseFlag	  bool
	TargetApiUrl      string
	LastLoadApisTime  time.Time
	IntervalTime      int64
	SourceContentType string
	ContentType       string
	Message           interface{}
}



var(
	gatewayLogger = logger.GatewayLogger
)

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
			TargetApiUrl: "",
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
				TargetApiUrl: "",
				ApiInfo:      &models.GatewayApiInfo{},
				AppInfo:      &models.AppInfo{},
			}
		}
	}
	return apiContext
}

// doBalanceTargetApi do balance real target apiurl
// if not exists alive target, return error info
func doBalanceTargetApi(apiContext *servermodel.ApiContext) (retCode int, retMsg string, realApiUrl string){
	retCode = _const.RetCode_OK

	if apiContext.ApiInfo.ApiType != _const.ApiType_Balance {
		retCode = -100010
		retMsg = "get targetapi failed, not balance mode!"
		return
	}
	//获取本次请求处理的目标Api，加入负载机制
	if apiContext.ApiInfo.TargetApi != nil && len(apiContext.ApiInfo.TargetApi) >0 {
		targetApi := balance.GetAliveApi(apiContext.ApiInfo)
		if targetApi == "" {
			retCode = -100010
			retMsg = "get targetapi failed, load targetapi nil!"
			return
		} else{
			//组合API地址与参数
			realApiUrl = targetApi
		}
	}else{
		if apiContext.ApiInfo.ApiUrl != "" {
			realApiUrl = apiContext.ApiInfo.ApiUrl
		}else{
			retCode = -100010
			retMsg = "get targetapi failed, no config apiurl!"
		}
	}
	return
}

// ProxyGet route all Get requests to real target server
// returns: ResponseJson: RetCode,RetMsg,LastLoadApisTime,IntervalTime,ContentType,Message
func ProxyGet(ctx dotweb.Context) error{

	defer func() {
		if err := recover(); err != nil {
			ex := exception.CatchError(_const.ProjectName+":ProxyGet", err)
			gatewayLogger.Error(ex.GetDefaultLogString())
			os.Stdout.Write([]byte(ex.GetDefaultLogString()))
		}
	}()

	var resJson ResponseJson
	var apiContext *servermodel.ApiContext
	resJson.RetCode = 0
	resJson.RetMsg = "ok"

	//get ApiContext from item
	apiContext = getApiContextFromItem(ctx)
	resJson.RetCode = apiContext.RetCode
	resJson.RetMsg = apiContext.RetMsg
	resJson.LastLoadApisTime = config.LastLoadApisTime

	if resJson.RetCode == _const.RetCode_OK {
		if apiContext.ApiInfo.ApiType == _const.ApiType_Balance {
			resJson.RetCode, resJson.RetMsg, apiContext.TargetApiUrl = doBalanceTargetApi(apiContext)
			if resJson.RetCode ==0{
				realApiUrl := combineApiUrl(apiContext.TargetApiUrl, apiContext.Query)
				body, contentType, intervalTime, err := httpx.HttpGet(realApiUrl)
				if err != nil {
					resJson.RetCode = _const.RetCode_Error
					resJson.RetMsg = body
					resJson.Message = err.Error()
					balance.SetError(apiContext.ApiInfo, apiContext.TargetApiUrl)
				} else {
					resJson.RetCode = _const.RetCode_OK
					resJson.RetMsg = "ok"
					resJson.Message = body
				}
				resJson.IntervalTime = intervalTime
				resJson.ContentType = contentType
			}
		}
		if apiContext.ApiInfo.ApiType == _const.ApiType_Group{
			var syncWait sync.WaitGroup
			var targetResults []*models.TargetApiResult
			for _, v:=range apiContext.ApiInfo.TargetApi{
				syncWait.Add(1)
				go func(){
					defer syncWait.Done()
					result := new(models.TargetApiResult)
					result.ApiKey = v.TargetKey
					body, _, intervalTime, err := httpx.HttpGet(combineApiUrl(v.TargetUrl, apiContext.Query))
					if err != nil{
						result.RetCode =_const.RetCode_Error
						result.RetMsg = err.Error()
					}else{
						errJson := json.Unmarshal([]byte(body), result)
						if errJson != nil{
							result.RetCode =_const.RetCode_JsonUnmarshalError
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

	//日志处理
	logJson := LogJson{
		RetCode:          resJson.RetCode,
		RetMsg:           resJson.RetMsg,
		RequestUrl:       string(ctx.Request().Url()),
		TargetApiUrl:     apiContext.TargetApiUrl,
		HttpMethod:       "Get",
		IntervalTime:     resJson.IntervalTime,
		LastLoadApisTime: resJson.LastLoadApisTime,
		ContentType:      resJson.ContentType,
		Message:          resJson.Message,
	}
	if apiContext.ApiInfo!= nil{
		logJson.RawResponseFlag = apiContext.ApiInfo.RawResponseFlag
	}else{
		logJson.RawResponseFlag = false
	}

	jsonLogB, _ := json.Marshal(logJson)
	gatewayLogger.Info(string(jsonLogB))

	//计数处理
	apiId := 0
	if apiContext.ApiInfo!= nil{
		apiId = apiContext.ApiInfo.ApiID
	}
	metric.AddApiCount(apiContext.GateAppID, apiId, apiContext.ApiModule, apiContext.ApiName, apiContext.ApiVersion, 1, strconv.Itoa(resJson.RetCode))
	ctx.WriteString(responseContent)
	return nil
}

// ProxyGet route all Post requests to real target server
// returns: ResponseJson: RetCode,RetMsg,LastLoadApisTime,IntervalTime,ContentType,Message
func ProxyPost(ctx dotweb.Context) error{
	defer func() {
		if err := recover(); err != nil {
			ex := exception.CatchError(_const.ProjectName+":ProxyGet", err)
			gatewayLogger.Error(ex.GetDefaultLogString())
			os.Stdout.Write([]byte(ex.GetDefaultLogString()))
		}
	}()

	var resJson ResponseJson
	var apiContext *servermodel.ApiContext
	resJson.RetCode = 0
	resJson.RetMsg = "ok"
	sourceContentType := ctx.Request().QueryHeader("Content-Type")

	//get api context from item
	apiContext = getApiContextFromItem(ctx)
	resJson.RetCode = apiContext.RetCode
	resJson.RetMsg = apiContext.RetMsg
	resJson.LastLoadApisTime = config.LastLoadApisTime

	if resJson.RetCode == 0 {
		postcontent := ctx.Request().PostBody()
		if apiContext.ApiInfo.ApiType == _const.ApiType_Balance {
			resJson.RetCode, resJson.RetMsg, apiContext.TargetApiUrl = doBalanceTargetApi(apiContext)
			if resJson.RetCode ==0 {
				realApiUrl := combineApiUrl(apiContext.TargetApiUrl, apiContext.Query)
				body, contentType, intervalTime, err := httpx.HttpPost(realApiUrl, string(postcontent), sourceContentType)
				if err != nil {
					resJson.RetCode = -209999
					resJson.RetMsg = body
					resJson.Message = err.Error()
					balance.SetError(apiContext.ApiInfo, apiContext.TargetApiUrl)
				} else {
					resJson.RetCode = 0
					resJson.RetMsg = "ok"
					resJson.Message = body
				}
				resJson.IntervalTime = intervalTime
				resJson.ContentType = contentType
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
					realApiUrl := combineApiUrl(v.TargetUrl, apiContext.Query)
					body, _, intervalTime, err := httpx.HttpPost(realApiUrl, string(postcontent), sourceContentType)
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

	//日志处理
	logJson := LogJson{
		RetCode:           resJson.RetCode,
		RetMsg:            resJson.RetMsg,
		RequestUrl:        string(ctx.Request().Url()),
		TargetApiUrl:      apiContext.TargetApiUrl,
		HttpMethod:        "Post",
		IntervalTime:      resJson.IntervalTime,
		LastLoadApisTime:  resJson.LastLoadApisTime,
		SourceContentType: sourceContentType,
		ContentType:       resJson.ContentType,
		Message:           resJson.Message,
	}

	if apiContext.ApiInfo!= nil{
		logJson.RawResponseFlag = apiContext.ApiInfo.RawResponseFlag
	}else{
		logJson.RawResponseFlag = false
	}

	jsonLogB, _ := json.Marshal(logJson)
	gatewayLogger.Info(string(jsonLogB))
	//计数处理
	apiId := 0
	if apiContext.ApiInfo!= nil{
		apiId = apiContext.ApiInfo.ApiID
	}
	metric.AddApiCount(apiContext.GateAppID, apiId, apiContext.ApiModule, apiContext.ApiName, apiContext.ApiVersion, 1, strconv.Itoa(resJson.RetCode))
	ctx.WriteString(responseContent)
	return nil
}

// ProxyGet route all Post requests to real target server
func ProxyLocal(ctx dotweb.Context) error{

	defer func() {
		if err := recover(); err != nil {
			ex := exception.CatchError(_const.ProjectName+":ProxyLocal", err)
			gatewayLogger.Error(ex.GetDefaultLogString())
			os.Stdout.Write([]byte(ex.GetDefaultLogString()))
		}
	}()

	var resJson ResponseJson
	var apiContext *servermodel.ApiContext
	resJson.RetCode = 0
	resJson.RetMsg = "ok"
	sourceContentType := ctx.Request().QueryHeader("Content-Type")

	//get api context from item
	apiContext = getApiContextFromItem(ctx)
	resJson.RetCode = apiContext.RetCode
	resJson.RetMsg = apiContext.RetMsg
	resJson.LastLoadApisTime = config.LastLoadApisTime

	if resJson.RetCode == 0 {
		//解析url参数

		echo := ctx.QueryString("echo")
		resJson.RetCode = 0
		resJson.RetMsg = "ok"
		resJson.Message = echo

		resJson.IntervalTime = 0
		resJson.ContentType = sourceContentType
	}

	responseContent := ""
	jsonb, err := json.Marshal(resJson)
	//解析异常处理
	if err != nil {
		responseContent = `{"RetCode":-109999,"RetMsg":"json marshal error - ` + err.Error() + `","Message":"` + resJson.Message.(string) + `}`
	} else {
		responseContent = string(jsonb)
	}
	//日志处理
	logJson := LogJson{
		RetCode:          resJson.RetCode,
		RetMsg:           resJson.RetMsg,
		RequestUrl:       string(ctx.Request().Url()),
		TargetApiUrl:     apiContext.TargetApiUrl,
		HttpMethod:       "Get",
		IntervalTime:     resJson.IntervalTime,
		LastLoadApisTime: resJson.LastLoadApisTime,
		ContentType:      resJson.ContentType,
		Message:          resJson.Message,
	}
	jsonLogB, _ := json.Marshal(logJson)
	gatewayLogger.Info(string(jsonLogB))
	//计数处理
	apiId := 0
	if apiContext.ApiInfo!= nil{
		apiId = apiContext.ApiInfo.ApiID
	}
	metric.AddApiCount(apiContext.GateAppID, apiId, apiContext.ApiModule, apiContext.ApiName, apiContext.ApiVersion, 1, strconv.Itoa(resJson.RetCode))
	ctx.WriteString(responseContent)

	return nil
}
