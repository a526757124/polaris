package _const

const(
	ProjectName = "Polaris"
	ProjectVersion = "1.0.0"
)


const (
	ValidateType_Normal = 0
	ValidateType_MD5    = 1

	AppStatus_Normal = 0
	AppStatus_Stop   = -100

	ApiStatus_Init   = 0
	ApiStatus_Normal = 100
	ApiStatus_Stop   = -100

	ServiceHostType_Input = 0	//手工录入
	ServiceHostType_Consul = 1	//Consul匹配

	ApiType_Balance = 0 //负载方式
	ApiType_Group = 1 //组合方式

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


	RetCode_OK = 0
)

const(
	Redis_Key_AppMap         = "Polaris:GatewayAppHash"
	Redis_Key_ApiMap         = "Polaris:GatewayApiHash"
	Redis_Key_AppApiRelation = "Polaris:AppGatewayApiRelationHash"
	Redis_Key_CommonPre      = "Polaris.ApiGateway"
)

const(
	Default_Redis_MaxIdle = 20
	Default_Redis_MaxActive = 100

	DefaultDateLayout     = "2006-01-02"
	DefaultFullTimeLayout = "2006-01-02 15:04:05.999999"
	DefaultTimeLayout     = "2006-01-02 15:04:05"
)

const (
	HttpParam_GateAppID   = "gate_appid"
	HttpParam_GateEncrypt = "gate_encrypt"

	HttpContext_ApiContext = "polaris_ApiContext"
)
