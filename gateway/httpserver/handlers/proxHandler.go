// httpHandler
package handlers

import (
	"github.com/devfeel/polaris/config"
	"github.com/devfeel/polaris/const"
	"github.com/devfeel/polaris/models"
	"github.com/devfeel/polaris/control/ratelimit"
	"github.com/devfeel/polaris/util/logx"
	"github.com/devfeel/polaris/gateway/balance"
	"github.com/devfeel/polaris/gateway/auth"

	"strconv"
	"strings"
	"time"
	"github.com/devfeel/dotweb"
	"net/http"
	"os"
	"fmt"
	"runtime"
	"github.com/devfeel/polaris/util/httpx"
	"sync"
	"encoding/json"
	"github.com/devfeel/polaris/control/metric"
	"github.com/devfeel/polaris/core/exception"
)

type ResponseJson struct {
	RetCode          int
	RetMsg           string
	LastLoadApisTime time.Time
	IntervalTime     int64
	ContentType      string
	Message          interface{}
}

//Api上下文
type ApiContext struct {
	GateAppID    string
	Query string
	AppInfo      *models.AppInfo
	ApiInfo      *models.GatewayApiInfo
	ApiModule    string
	ApiName      string
	ApiVersion   string
	RetCode      int
	RetMsg       string
	TargetApiUrl string
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

const (
	HttpParam_GateAppID   = "gate_appid"
	HttpParam_GateEncrypt = "gate_encrypt"
)

var(
	gatewayLogger = logger.GatewayLogger
)

// resolveApiPath解析api请求目录
// 解析成功则返回ApiModule、ApiKey、ApiVersion、ApiUrlKey
func resolveApiPath(ctx dotweb.Context) (apiModule, apiKey, apiVersion, apiUrlKey string) {
	apiModule = ctx.GetRouterName("module")
	apiKey = ctx.GetRouterName("apikey")
	apiVersion = ctx.GetRouterName("version")
	apiUrlKey = apiModule + "/" + apiKey + "/" + apiVersion
	return apiModule, apiKey, apiVersion, apiUrlKey
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

// validateMD5Sign validate md5 sign
func validateMD5Sign(ctx dotweb.Context, md5Key string, appEncrypt string) (retCode int, retMsg string) {
	queryArgs := ctx.Request().QueryStrings()
	queryArgs.Del(HttpParam_GateEncrypt)
	postBody := ""
	//if post, add post string
	if ctx.Request().Method == http.MethodPost {
		if ctx.Request().PostBody() != nil && len(ctx.Request().PostBody()) > 0 {
			postBody = string(ctx.Request().PostBody())
		}
	}
	appVal, gateVal, isOk := auth.ValidateMD5Sign(queryArgs, postBody, md5Key, appEncrypt)
	if isOk{
		retMsg = ""
		retCode = _const.RetCode_OK
	}else{
		retMsg = "CheckEncrypt failed! -> " + appVal + " == " + gateVal
		retCode = -100009
	}
	return
}

// getGateParam get Gate param(GateAppID, GateEncrypt)
// first get from head, if not exists, get from query string
func getGateParam(ctx dotweb.Context) (appid, encrypt string){
	gateEncrypt := ctx.Request().QueryHeader(HttpParam_GateEncrypt)
	gateAppID := ctx.Request().QueryHeader(HttpParam_GateAppID)
	if gateAppID == ""{
		gateAppID = ctx.QueryString(HttpParam_GateAppID)
	}
	if gateEncrypt == ""{
		gateEncrypt = ctx.QueryString(HttpParam_GateEncrypt)
	}
	return gateAppID, gateEncrypt
}

// cleanQueryData cleaning query data
func cleanQueryData(ctx dotweb.Context) string{
	queryArgs := ctx.Request().QueryStrings()
	queryArgs.Del(HttpParam_GateEncrypt)
	queryArgs.Del(HttpParam_GateAppID)
	query := queryArgs.Encode()
	return query
}

// doBalanceTargetApi do balance real target apiurl
// if not exists alive target, return error info
func doBalanceTargetApi(apiContext *ApiContext) (retCode int, retMsg string, realApiUrl string){
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

/* 基础检查
* 检查AppID是否合法
* 检查ApiKey是否合法
* 检查AppID与ApiKey是否有权限
* 在指定MD5鉴权方式下，检查鉴权串是否合法
* 负载决定最终TargetApiUrl，参数匹配等 --add by pxm 20160830
* 判断是否为负载类型
* 如果合法，返回TargetApiUrl\AppInfo\ApiInfo及RetCode、RetMsg
 */
func baseCheck(ctx dotweb.Context) (apiContext *ApiContext) {
	flag := true

	apiContext = &ApiContext{
		RetCode:      0,
		RetMsg:       "",
		TargetApiUrl: "",
		ApiInfo:&models.GatewayApiInfo{},
		AppInfo:&models.AppInfo{},
	}

	//解析api请求目录
	apiModule, apiKey, apiVersion, apiUrlKey := resolveApiPath(ctx)
	if apiModule == "" || apiKey == "" || apiVersion == "" || apiUrlKey == "" {
		apiContext.RetMsg = "Not supported Api(QueryPath resolve error) => " + string(ctx.Request().RequestURI)
		apiContext.RetCode = -100001
		return
	}

	apiContext.ApiModule = apiModule
	apiContext.ApiName = apiKey
	apiContext.ApiVersion = apiVersion

	//get appid, encrypt
	gateAppID, gateEncrypt := getGateParam(ctx)

	//query data cleaning
	apiContext.Query = cleanQueryData(ctx)

	if gateAppID == "" {
		apiContext.RetMsg = "unable to resolve query:lost gate_appid"
		apiContext.RetCode = -100002
		return
	}

	apiContext.GateAppID = gateAppID

	//获取对应AppInfo
	apiContext.AppInfo, flag = config.GetAppInfo(gateAppID)
	if !flag {
		apiContext.RetMsg = "not support app"
		apiContext.RetCode = -100003
		return
	}

	//判断App状态是否合法
	if apiContext.AppInfo.Status != _const.AppStatus_Normal {
		apiContext.RetMsg = "app not activate status"
		apiContext.RetCode = -100004
		return
	}

	//获取对应GatewayApiInfo
	apiContext.ApiInfo, flag = config.GetApiInfo(apiUrlKey)
	if !flag {
		apiContext.RetMsg = "not support api"
		apiContext.RetCode = -100005
		return
	}

	//判断Api状态是否合法
	if apiContext.ApiInfo.Status != _const.ApiStatus_Normal {
		apiContext.RetMsg = "api not activate status"
		apiContext.RetCode = -100006
		return
	}

	if !config.CheckAppApiRelation(apiContext.AppInfo.AppID, apiContext.ApiInfo.ApiID) {
		apiContext.RetMsg = "no have this api's permissions"
		apiContext.RetCode = -100007
		return
	}

	//IP检查
	if len(apiContext.ApiInfo.ValidIPs) > 0 {
		isValid := false
		for _, v := range apiContext.ApiInfo.ValidIPs {
			if v == ctx.RemoteIP() {
				isValid = true
				break
			}
		}
		if !isValid {
			apiContext.RetMsg = "not allowed ip"
			apiContext.RetCode = -100011
			return
		}
	}

	//风控处理，每分钟同IP调用次数限制
	isInLimit := ratelimit.RedisLimiter.RequestCheck(strconv.Itoa(apiContext.ApiInfo.ApiID)+ "_" + ctx.RemoteIP(), 1)
	if !isInLimit {
		apiContext.RetMsg = "The number of requests exceeds the upper limit of each minute"
		apiContext.RetCode = -100008
		return
	}

	//鉴权处理
	if apiContext.ApiInfo.ValidateType == _const.ValidateType_MD5 {
		if retCode, retMsg := validateMD5Sign(ctx, apiContext.AppInfo.AppKey, gateEncrypt); retCode != _const.RetCode_OK {
			apiContext.RetMsg = retMsg
			apiContext.RetCode = retCode
			return
		}
	}

	//组合方式基础检查
	if apiContext.ApiInfo.ApiType == _const.ApiType_Group {
		if apiContext.ApiInfo.TargetApi == nil || len(apiContext.ApiInfo.TargetApi) <=0 {
			apiContext.RetMsg = "get targetapi failed, load targetapi nil!"
			apiContext.RetCode = -100010
			return
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
	resJson.RetCode = 0
	resJson.RetMsg = "ok"

	//基础检查，如果合法，返回ApiContent
	apiContext := baseCheck(ctx)

	resJson.RetCode = apiContext.RetCode
	resJson.RetMsg = apiContext.RetMsg
	resJson.LastLoadApisTime = config.LastLoadApisTime

	if resJson.RetCode == 0 {
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
	resJson.RetCode = 0
	resJson.RetMsg = "ok"
	sourceContentType := ctx.Request().QueryHeader("Content-Type")

	//基础检查，如果合法，返回TargetApiUrl
	apiContext := baseCheck(ctx)
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

func ProxyLocal(ctx dotweb.Context) error{

	defer func() {
		if err := recover(); err != nil {
			os.Stdout.Write([]byte("httpHandler::ProxyLocal error! -> " + fmt.Sprint(err)))
			gatewayLogger.Error(fmt.Sprint(err))
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, true)
			gatewayLogger.Error(string(buf[:n]))
			os.Stdout.Write(buf[:n])
		}
	}()

	metric.AddRequestCount(1)

	var resJson ResponseJson
	resJson.RetCode = 0
	resJson.RetMsg = "ok"

	//基础检查，如果合法，返回TargetApiUrl
	apiContext := baseCheck(ctx)
	resJson.RetCode = apiContext.RetCode
	resJson.RetMsg = apiContext.RetMsg
	resJson.LastLoadApisTime = config.LastLoadApisTime
	contentType := ctx.Request().QueryHeader("Content-Type")
	if resJson.RetCode == 0 {
		//解析url参数

		echo := ctx.QueryString("echo")
		resJson.RetCode = 0
		resJson.RetMsg = "ok"
		resJson.Message = echo

		resJson.IntervalTime = 0
		resJson.ContentType = contentType
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
