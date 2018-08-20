package middlewares

import (
	"github.com/devfeel/dotweb"
	"net/http"
	"github.com/devfeel/polaris/gateway/auth"
	"github.com/devfeel/polaris/const"
	"github.com/devfeel/polaris/gateway/httpserver/servermodel"
	"github.com/devfeel/polaris/models"
	"github.com/devfeel/polaris/config"
	"strconv"
	"github.com/devfeel/polaris/control/ratelimit"
)

// ValidateMiddleware the middleware to do base validate
// 1. ResolveApiPath
// 2. Check exists gate_appid in QueryString or request Header
// 3. Load AppInfo from config
// 4. Load ApiInfo from config
// 5. Check Api Permissions
// 6. Validate request IP
// 7. Validate RateLimit
// 8. Validate MD5 Sign
// 9. Validate EnoughApi when in Group Mode
type ValidateMiddleware struct{
	dotweb.BaseMiddlware
}

// Handle current middleware's handler
func (m *ValidateMiddleware) Handle(ctx dotweb.Context) error {
	flag := true

	apiContext := &servermodel.ApiContext{
		RetCode:      0,
		RetMsg:       "",
		RealTargetApi: nil,
		ApiInfo:&models.GatewayApiInfo{},
		AppInfo:&models.AppInfo{},
	}

	//解析api请求目录
	apiModule, apiKey, apiVersion, apiUrlKey := resolveApiPath(ctx)
	if apiModule == "" || apiKey == "" || apiVersion == "" || apiUrlKey == "" {
		apiContext.RetMsg = "Not supported Api(QueryPath resolve error) => " + string(ctx.Request().RequestURI)
		apiContext.RetCode = _const.RetCode_Validate_ResolveApiPathError
	}

	apiContext.ApiModule = apiModule
	apiContext.ApiName = apiKey
	apiContext.ApiVersion = apiVersion

	//get appid, encrypt
	gateAppID, gateEncrypt := getGateParam(ctx)

	//query data cleaning
	apiContext.Query = cleanQueryData(ctx)

	if apiContext.RetCode == _const.RetCode_OK {
		if gateAppID == "" {
			apiContext.RetMsg = "unable to resolve query:lost gate_appid"
			apiContext.RetCode = _const.RetCode_Validate_NotExistsAppID
		}
	}

	if apiContext.RetCode == _const.RetCode_OK {
		apiContext.GateAppID = gateAppID
		apiContext.AppInfo, flag = config.GetAppInfo(gateAppID)
		if !flag {
			apiContext.RetMsg = "not support app"
			apiContext.RetCode = _const.RetCode_Validate_NotSupportApp
		}
	}

	//判断App状态是否合法
	if apiContext.RetCode == _const.RetCode_OK {
		if apiContext.AppInfo.Status != _const.AppStatus_Normal {
			apiContext.RetMsg = "app not activate status"
			apiContext.RetCode = _const.RetCode_Validate_AppNotActive
		}
	}

	//获取对应GatewayApiInfo
	if apiContext.RetCode == _const.RetCode_OK {
		apiContext.ApiInfo, flag = config.GetApiInfo(apiUrlKey)
		if !flag {
			apiContext.RetMsg = "not support api"
			apiContext.RetCode = _const.RetCode_Validate_NotSupportAPI
		}
	}

	//判断Api状态是否合法
	if apiContext.RetCode == _const.RetCode_OK {
		if apiContext.ApiInfo.Status != _const.ApiStatus_Normal {
			apiContext.RetMsg = "api not activate status"
			apiContext.RetCode = _const.RetCode_Validate_ApiNotActive
		}
	}

	if apiContext.RetCode == _const.RetCode_OK {
		if !config.CheckAppApiRelation(apiContext.AppInfo.AppID, apiContext.ApiInfo.ApiID) {
			apiContext.RetMsg = "no have this api's permissions"
			apiContext.RetCode = _const.RetCode_Validate_NoHaveApiPermissions
		}
	}

	//IP validate
	if apiContext.RetCode == _const.RetCode_OK {
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
				apiContext.RetCode = _const.RetCode_Validate_NotAllowedIP
			}
		}
	}

	//rate limit
	if apiContext.RetCode == _const.RetCode_OK {
		isInLimit := ratelimit.RedisLimiter.RequestCheck(strconv.Itoa(apiContext.ApiInfo.ApiID)+"_"+ctx.RemoteIP(), 1)
		if !isInLimit {
			apiContext.RetMsg = "The number of requests exceeds the upper limit of each minute"
			apiContext.RetCode = _const.RetCode_Validate_RateLimit
		}
	}

	//validate md5 sign
	if apiContext.RetCode == _const.RetCode_OK {
		if apiContext.ApiInfo.ValidateType == _const.ValidateType_MD5 {
			retCode, retMsg := validateMD5Sign(ctx, apiContext.AppInfo.AppKey, gateEncrypt)
			apiContext.RetCode = retCode
			apiContext.RetMsg = retMsg
		}
	}

	//validate enough target api when is group type
	if apiContext.RetCode == _const.RetCode_OK {
		if apiContext.ApiInfo.ApiType == _const.ApiType_Group {
			if apiContext.ApiInfo.TargetApi == nil || len(apiContext.ApiInfo.TargetApi) <= 0 {
				apiContext.RetMsg = "get targetapi failed, load targetapi nil!"
				apiContext.RetCode = _const.RetCode_Validate_NoEnoughApiInGroup
			}
		}
	}

	apiContext.ContentType = ctx.Request().QueryHeader("Content-Type")
	ctx.Items().Set(_const.HttpContext_ApiContext, apiContext)
	m.Next(ctx)
	return nil

}

func NewValidateMiddleware() dotweb.Middleware{
	return &ValidateMiddleware{}
}

// resolveApiPath resolve api's request path
// returns: ApiModule、ApiKey、ApiVersion、ApiUrlKey
func resolveApiPath(ctx dotweb.Context) (apiModule, apiKey, apiVersion, apiUrlKey string) {
	apiModule = ctx.GetRouterName("module")
	apiKey = ctx.GetRouterName("apikey")
	apiVersion = ctx.GetRouterName("version")
	apiUrlKey = apiModule + "/" + apiKey + "/" + apiVersion
	return apiModule, apiKey, apiVersion, apiUrlKey
}

// validateMD5Sign validate md5 sign
func validateMD5Sign(ctx dotweb.Context, md5Key string, appEncrypt string) (retCode int, retMsg string) {
	queryArgs := ctx.Request().QueryStrings()
	queryArgs.Del(_const.HttpParam_GateEncrypt)
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
		retCode = _const.RetCode_Validate_MD5SignError
	}
	return
}

// getGateParam get Gate param(GateAppID, GateEncrypt)
// first get from head, if not exists, get from query string
func getGateParam(ctx dotweb.Context) (appid, encrypt string){
	gateEncrypt := ctx.Request().QueryHeader(_const.HttpParam_GateEncrypt)
	gateAppID := ctx.Request().QueryHeader(_const.HttpParam_GateAppID)
	if gateAppID == ""{
		gateAppID = ctx.QueryString(_const.HttpParam_GateAppID)
	}
	if gateEncrypt == ""{
		gateEncrypt = ctx.QueryString(_const.HttpParam_GateEncrypt)
	}
	return gateAppID, gateEncrypt
}

// cleanQueryData cleaning query data
func cleanQueryData(ctx dotweb.Context) string{
	queryArgs := ctx.Request().QueryStrings()
	queryArgs.Del(_const.HttpParam_GateEncrypt)
	queryArgs.Del(_const.HttpParam_GateAppID)
	query := queryArgs.Encode()
	return query
}
