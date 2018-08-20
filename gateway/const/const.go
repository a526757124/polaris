package _const

const(
	RetCode_Error = -9999
	RetCode_Hystrix = -9998
	RetCode_JsonUnmarshalError = -9001
	RetCode_JsonMarshalError = -9002
	RetCode_GetApiContextError = -9003

	RetCode_Validate_ResolveApiPathError = -100001
	RetCode_Validate_NotExistsAppID = -100002
	RetCode_Validate_NotSupportApp = -100003
	RetCode_Validate_AppNotActive = -100004
	RetCode_Validate_NotSupportAPI = -100005
	RetCode_Validate_ApiNotActive = -100006
	RetCode_Validate_NoHaveApiPermissions = -100007
	RetCode_Validate_NotAllowedIP = -100008
	RetCode_Validate_RateLimit = -100009
	RetCode_Validate_MD5SignError = -100010
	RetCode_Validate_NoEnoughApiInGroup = -100011
	RetCode_Balance_NobalanceMode = -100012
	RetCode_Balance_LoadNil = -100013


	RetCode_OK = 0
)

const (
	HttpParam_GateAppID   = "gate_appid"
	HttpParam_GateEncrypt = "gate_encrypt"
	HttpContext_ApiContext = "polaris_ApiContext"
)
