package models


type GlobalConfig struct {
	Server 	*Server
	Consul  *Consul
	Redis   *Redis
	Global	*Global
}

type Global struct {
	CountLogApi			string
	ConfigCacheMins 	int
	UseDefaultTestApp	bool
}

//Consul config
type Consul struct{
	IsUse 		bool
	ServerUrl 	string
}

//Server server config
type Server struct {
	HttpPort 	int
	JsonRpcPort int
}

//Redis redis config
type Redis struct {
	ServerUrl 		string
	BackupServerUrl string
	MaxIdle 		int
	MaxActive 		int
}