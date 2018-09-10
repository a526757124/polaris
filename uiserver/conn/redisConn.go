package conn

import (
	"github.com/a526757124/polaris/config"
	"github.com/devfeel/polaris/util/redisx"
)

var redisClient *redisx.RedisClient

type RedisConn struct{}

func init() {
	address := config.CurrentConfig.Redis.ServerUrl
	maxIdle := config.CurrentConfig.Redis.MaxIdle
	maxActive := config.CurrentConfig.Redis.MaxActive
	redisClient = redisx.GetRedisClient(address, maxIdle, maxActive)
}

//get redisclient conn
func GetRedisClient() *redisx.RedisClient {
	if redisClient == nil {
		panic("redis connection failed!")
	}
	return redisClient
}
