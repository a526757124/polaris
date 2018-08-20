package middlewares

import (
	"github.com/devfeel/dotweb"
	"net/http"
	"github.com/devfeel/polaris/gateway/auth"
	"github.com/devfeel/polaris/const"
	. "github.com/devfeel/polaris/gateway/const"
	. "github.com/devfeel/polaris/gateway/models"
	"github.com/devfeel/polaris/models"
	"github.com/devfeel/polaris/gateway/request"
)

// ValidateMiddleware the middleware to do base validate
// 1. ResolveApiPath
// 2. Check exists gate_appid in QueryString or request Header
// 3. Load AppInfo from config
// 4. Load ApiInfo from config
// 5. Check Api Permissions
// 6. Validate request IP
// 7. Validate RateLimit
// 8. Validate EnoughApi when in Group Mode
// 9. Validate MD5 Sign
type ValidateMiddleware struct{
	dotweb.BaseMiddlware
}

// Handle current middleware's handler
func (m *ValidateMiddleware) Handle(ctx dotweb.Context) error {

	apiContext := &ApiContext{
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
		apiContext.RetCode = RetCode_Validate_ResolveApiPathError
	}

	apiContext.ApiModule = apiModule
	apiContext.ApiName = apiKey
	apiContext.ApiVersion = apiVersion
	apiContext.ApiUrlKey = apiUrlKey
	apiContext.RemoteIP = ctx.RemoteIP()
	apiContext.ContentType = ctx.Request().QueryHeader("Content-Type")

	//get appid, encrypt
	gateAppID, gateEncrypt := getGateParam(ctx)

	//query data cleaning
	apiContext.Query = cleanQueryData(ctx)

	if apiContext.RetCode == RetCode_OK {
		if gateAppID == "" {
			apiContext.RetMsg = "unable to resolve query:lost gate_appid"
			apiContext.RetCode = RetCode_Validate_NotExistsAppID
		}
	}

	if apiContext.RetCode == RetCode_OK {
		request.DoValidate(apiContext)
	}

	//validate md5 sign
	if apiContext.RetCode == RetCode_OK {
		if apiContext.ApiInfo.ValidateType == _const.ValidateType_MD5 {
			retCode, retMsg := validateMD5Sign(ctx, apiContext.AppInfo.AppKey, gateEncrypt)
			apiContext.RetCode = retCode
			apiContext.RetMsg = retMsg
		}
	}

	ctx.Items().Set(HttpContext_ApiContext, apiContext)
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
		retCode = RetCode_OK
	}else{
		retMsg = "CheckEncrypt failed! -> " + appVal + " == " + gateVal
		retCode = RetCode_Validate_MD5SignError
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
