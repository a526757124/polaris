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

const(
	CallMethod_HttpGet = "HttpGet"
	CallMethod_HttpPost = "HttpPost"
	CallMethod_JsonRPC = "JsonRPC"
)