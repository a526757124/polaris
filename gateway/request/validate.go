package request

import (
	"github.com/devfeel/polaris/control/ratelimit"
	. "github.com/devfeel/polaris/gateway/models"
	. "github.com/devfeel/polaris/gateway/const"
	"github.com/devfeel/polaris/config"
	"github.com/devfeel/polaris/const"
	"github.com/devfeel/polaris/models"
)

// DoValidate validate the ApiContext
// 1. Load AppInfo from config
// 2. Load ApiInfo from config
// 3. Check Api Permissions
// 4. Validate request IP
// 5. Validate RateLimit
// 6. Validate EnoughApi when in Group Mode
func DoValidate(apiContext *ApiContext) {
	flag := true
	apiContext.AppInfo, flag = config.GetAppInfo(apiContext.GateAppID)
	if !flag {
		apiContext.RetMsg = "not support app"
		apiContext.RetCode = RetCode_Validate_NotSupportApp
		apiContext.AppInfo = &models.AppInfo{}
	}

	//check app's status
	if apiContext.RetCode == RetCode_OK {
		if apiContext.AppInfo.Status != _const.AppStatus_Normal {
			apiContext.RetMsg = "app not activate status"
			apiContext.RetCode = RetCode_Validate_AppNotActive
		}
	}

	if apiContext.RetCode == RetCode_OK {
		apiContext.ApiInfo, flag = config.GetApiInfo(apiContext.ApiUrlKey)
		if !flag {
			apiContext.RetMsg = "not support api"
			apiContext.RetCode = RetCode_Validate_NotSupportAPI
			apiContext.ApiInfo = &models.GatewayApiInfo{}
		}
	}

	//check api's status
	if apiContext.RetCode == RetCode_OK {
		if apiContext.ApiInfo.Status != _const.ApiStatus_Normal {
			apiContext.RetMsg = "api not activate status"
			apiContext.RetCode = RetCode_Validate_ApiNotActive
		}
	}

	if apiContext.RetCode == RetCode_OK {
		if !config.CheckRelation(apiContext.AppInfo.AppID, apiContext.ApiInfo.ApiID) {
			apiContext.RetMsg = "no have this api's permissions"
			apiContext.RetCode = RetCode_Validate_NoHaveApiPermissions
		}
	}

	//IP validate
	if apiContext.RetCode == RetCode_OK {
		if len(apiContext.ApiInfo.ValidIPs) > 0 {
			isValid := false
			for _, v := range apiContext.ApiInfo.ValidIPs {
				if v == apiContext.RemoteIP {
					isValid = true
					break
				}
			}
			if !isValid {
				apiContext.RetMsg = "not allowed ip"
				apiContext.RetCode = RetCode_Validate_NotAllowedIP
			}
		}
	}

	//rate limit
	if apiContext.RetCode == RetCode_OK {
		isInLimit := ratelimit.RedisLimiter.RequestCheck(apiContext.ApiInfo.ApiID+"_"+apiContext.RemoteIP, 1)
		if !isInLimit {
			apiContext.RetMsg = "The number of requests exceeds the upper limit of each minute"
			apiContext.RetCode = RetCode_Validate_RateLimit
		}
	}

	//validate enough target api when is group type
	if apiContext.RetCode == RetCode_OK {
		if apiContext.ApiInfo.ApiType == _const.ApiType_Group {
			if apiContext.ApiInfo.TargetApis == nil || len(apiContext.ApiInfo.TargetApis) <= 0 {
				apiContext.RetMsg = "get targetapi failed, load targetapi nil!"
				apiContext.RetCode = RetCode_Validate_NoEnoughApiInGroup
			}
		}
	}
}
