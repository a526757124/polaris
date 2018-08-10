// httpHandler
package handlers

import (
	"github.com/devfeel/polaris/config"
	"github.com/devfeel/polaris/const"
	"github.com/devfeel/polaris/models"
	"github.com/devfeel/polaris/control/metric"
	"github.com/devfeel/polaris/control/ratelimit"
	"github.com/devfeel/polaris/util/httpx"
	"github.com/devfeel/polaris/util/logx"
	"github.com/devfeel/polaris/gateway/httpserver/monitor"
	"github.com/devfeel/polaris/gateway/balance"
	"github.com/devfeel/polaris/gateway/auth"

	"encoding/json"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
	"github.com/devfeel/dotweb"
	"net/http"
	"fmt"

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

/*解析api请求目录
* Author: Panxinming
* LastUpdateTime: 2016-05-23 18:00
* 解析成功则返回ApiModule、ApiKey、ApiVersion、ApiUrlKey
 */
func resolveApiPath(ctx dotweb.Context) (apiModule, apiKey, apiVersion, apiUrlKey string) {
	apiModule = ctx.GetRouterName("module")
	apiKey = ctx.GetRouterName("apikey")
	apiVersion = ctx.GetRouterName("version")
	apiUrlKey = apiModule + "/" + apiKey + "/" + apiVersion
	return apiModule, apiKey, apiVersion, apiUrlKey
}

//根据api地址与查询参数，组合实际访问地址
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

//签名判断，检查传入加密串是否与服务器端一致
func checkSign(ctx dotweb.Context, md5Key string, appEncrypt string) (appVal, gateVal string, isOk bool) {
	queryArgs := ctx.Request().QueryStrings()
	queryArgs.Del(HttpParam_GateEncrypt)
	postBody := ""
	//if post, add post string
	if ctx.Request().Method == http.MethodPost {
		if ctx.Request().PostBody() != nil && len(ctx.Request().PostBody()) > 0 {
			postBody = string(ctx.Request().PostBody())
		}
	}
	return auth.CheckSign(queryArgs, postBody, md5Key, appEncrypt)
}

/* 基础检查
* Author: Panxinming
* LastUpdateTime: 2016-08-30 18:00
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

	//解析url参数
	queryArgs := ctx.Request().QueryStrings()
	gateEncrypt := ctx.QueryString(HttpParam_GateEncrypt)
	gateAppID := ctx.QueryString(HttpParam_GateAppID)
	queryArgs.Del(HttpParam_GateEncrypt)
	queryArgs.Del(HttpParam_GateAppID)
	query := queryArgs.Encode()

	//query string
	apiContext.Query = query

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
		if appVal, gateVal, isOK := checkSign(ctx, apiContext.AppInfo.AppKey, gateEncrypt); !isOK {
			apiContext.RetMsg = "CheckEncrypt failed! -> " + appVal + " == " + gateVal
			apiContext.RetCode = -100009
			return
		}
	}

	//负载方式检查TargetApi
	if apiContext.ApiInfo.ApiType == _const.ApiType_Balance {
		//获取本次请求处理的目标Api，加入负载机制
		if apiContext.ApiInfo.TargetApi != nil && len(apiContext.ApiInfo.TargetApi) >0 {
			targetApi := balance.GetTargetApi(apiContext.ApiInfo)
			if targetApi == nil {
				apiContext.RetMsg = "get targetapi failed, load targetapi nil!"
				apiContext.RetCode = -100010
				return
			} else{
				//组合API地址与参数
				apiContext.TargetApiUrl = combineApiUrl(targetApi.TargetUrl, query)
			}
		}else{
			if apiContext.ApiInfo.ApiUrl != "" {
				apiContext.TargetApiUrl = combineApiUrl(apiContext.ApiInfo.ApiUrl, query)
			}else{
				apiContext.RetMsg = "get targetapi failed, no config apiurl!"
				apiContext.RetCode = -100010
				return
			}
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

/* Get模式代理
* Author: Panxinming
* LastUpdateTime: 2016-08-30 18:00
* Get方式请求目标Api地址，参数透传
* 如果合法，输出状态码及目标Api返回结果
* RetCode,RetMsg,TargetApiUrl,HttpMethod,IntervalTime,ContentType,Message
* 增加负载支持 - add by pxm 20160830
 */
func ProxyGet(ctx dotweb.Context) error{

	defer func() {
		if err := recover(); err != nil {
			os.Stdout.Write([]byte("httpHandler::ProxyGet error! -> " + fmt.Sprint(err)))
			gatewayLogger.Error(fmt.Sprint(err))
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, true)
			gatewayLogger.Error(string(buf[:n]))
			os.Stdout.Write(buf[:n])
		}
	}()

	monitor.Current.AddRequestCount(1)
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
			body, contentType, intervalTime, err := httpx.HttpGet(apiContext.TargetApiUrl)
			if err != nil {
				resJson.RetCode = _const.RetCode_Error
				resJson.RetMsg = body
				resJson.Message = err.Error()

			} else {
				resJson.RetCode = _const.RetCode_OK
				resJson.RetMsg = "ok"
				resJson.Message = body
			}

			resJson.IntervalTime = intervalTime
			resJson.ContentType = contentType
		}
		if apiContext.ApiInfo.ApiType == _const.ApiType_Group{
			var targetResults []*models.TargetApiResult
			for _, v:=range apiContext.ApiInfo.TargetApi{
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
			}
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

//Post代理
func ProxyPost(ctx dotweb.Context) error{

	defer func() {
		if err := recover(); err != nil {
			os.Stdout.Write([]byte("httpHandler::ProxyPost error! -> " + fmt.Sprint(err)))
			gatewayLogger.Error(fmt.Sprint(err))
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, true)
			gatewayLogger.Error(string(buf[:n]))
			os.Stdout.Write(buf[:n])
		}
	}()

	monitor.Current.AddRequestCount(1)
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

			body, contentType, intervalTime, err := httpx.HttpPost(apiContext.TargetApiUrl, string(postcontent), sourceContentType)

			if err != nil {
				resJson.RetCode = -209999
				resJson.RetMsg = body
				resJson.Message = err.Error()

			} else {
				resJson.RetCode = 0
				resJson.RetMsg = "ok"
				resJson.Message = body
			}

			resJson.IntervalTime = intervalTime
			resJson.ContentType = contentType
		}
		if apiContext.ApiInfo.ApiType == _const.ApiType_Group{
			var targetResults []*models.TargetApiResult
			for _, v:=range apiContext.ApiInfo.TargetApi{
				result := new(models.TargetApiResult)
				result.ApiKey = v.TargetKey

				body, _, intervalTime, err := httpx.HttpPost(combineApiUrl(v.TargetUrl, apiContext.Query), string(postcontent), sourceContentType)
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
			}
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

	monitor.Current.AddRequestCount(1)
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
