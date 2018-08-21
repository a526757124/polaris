package redisbucket

import (
	"github.com/devfeel/polaris/config"
	"github.com/devfeel/polaris/util/redisx"
	"github.com/devfeel/polaris/const"
	"time"
)

const (
	defaultDateFormatForMins      = "200601021504"
	DefaultApiCallNumLimitPerMins = 10000
)

type RedisLimiter struct{

}

func NewRedisLimiter() *RedisLimiter{
	return &RedisLimiter{
	}
}


// RequestCheck
func (l *RedisLimiter) RequestCheck(key string, calls int) bool {
	key = getFullKey(key)
	redisClient := redisx.GetRedisClient(config.CurrentConfig.Redis.ServerUrl, config.CurrentConfig.Redis.MaxIdle, config.CurrentConfig.Redis.MaxActive)
	currentNum, err := redisClient.INCR(key)
	if err != nil {
		return true
	}
	return currentNum <= DefaultApiCallNumLimitPerMins
}

func getFullKey(key string) string{
	return _const.Redis_Key_CommonPre + ":" + time.Now().Format(defaultDateFormatForMins) + ":" + key
}
