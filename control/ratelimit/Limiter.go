package ratelimit

import (
	"github.com/devfeel/polaris/control/ratelimit/redisbucket"
	"github.com/devfeel/polaris/control/ratelimit/tokenbucket"
	"github.com/devfeel/polaris/control/ratelimit/leakybucket"
)

var(
	RedisLimiter Limiter
	TokenLimiter Limiter
	LeakyLimiter Limiter
)

type Limiter interface{
	// RequestCheck call request to check limit rules
	RequestCheck(key string, calls int) bool
}

func init(){
	RedisLimiter = redisbucket.NewRedisLimiter()
	TokenLimiter = tokenbucket.NewTokenLimiter()
	LeakyLimiter = leakybucket.NewLeakyLimiter()
}