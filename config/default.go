package config


var defaultConfig = &ProxyConfig{
	Server: Server{HttpPort:80, JsonRpcPort:1789},
	Consul: Consul{IsUse:false},
	Redis: Redis{},
	Global: Global{ConfigCacheMins:1},
}
