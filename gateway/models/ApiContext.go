package models

import "github.com/devfeel/polaris/models"

//Api上下文
type ApiContext struct {
	GateAppID    string
	Query 		string
	PostBody 	[]byte
	ContentType string
	RemoteIP	string
	AppInfo      *models.AppInfo
	ApiInfo      *models.GatewayApiInfo
	ApiUrlKey	 string
	ApiModule    string
	ApiName      string
	ApiVersion   string
	RetCode      int
	RetMsg       string
	RealTargetApi	 *models.TargetApiInfo
}
