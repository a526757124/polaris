package conn

import (
	"github.com/devfeel/polaris/util/redisx"
)

var redisClient *redisx.RedisClient

type RedisConn struct{}

func init() {
	address := "redis://127.0.0.1:6379/0"
	maxIdle := 20
	maxActive := 100
	redisClient = redisx.GetRedisClient(address, maxIdle, maxActive)
}

//get redisclient conn
func GetRedisClient() *redisx.RedisClient {
	if redisClient == nil {
		panic("redis connection failed!")
	}
	return redisClient
}
