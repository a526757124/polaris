package config


var defaultConfig = &ProxyConfig{
	Server: Server{HttpPort:80, JsonRpcPort:1789},
	ConsulSet: ConsulSet{IsUse:false},
	Redis: Redis{},
	GlobalSet: GlobalSet{ConfigCacheMins:1},
}
