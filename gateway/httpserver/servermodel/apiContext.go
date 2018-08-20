package servermodel

import "github.com/devfeel/polaris/models"

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
