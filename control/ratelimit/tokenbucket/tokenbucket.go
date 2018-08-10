package tokenbucket


import (
"time"
"sync"
"math"
)


type TokenBucket struct {
	stableIntervalMillisecond int64	// create tokens num per Millisecond
	tokens float64            		// current token size in bucket
	maxBucketCapacity float64		// max bucket capacity
	refreshTime time.Time          	// time for last token refresh
	refreshMutex *sync.Mutex		// refresh mutex
}


func newTokenBucket() *TokenBucket{
	return &TokenBucket{
		stableIntervalMillisecond : 10, //每毫秒生成数
		tokens : 10,                   //初始令牌数
		maxBucketCapacity : 10000,
		refreshTime : time.Now(),
		refreshMutex : new(sync.Mutex),
	}
}

func (tb *TokenBucket) refreshTokens() {
	now := time.Now()
	var added float64
	if now.Sub(tb.refreshTime).Nanoseconds() >= 1000000 {
		tb.refreshMutex.Lock()
		added = float64(now.Sub(tb.refreshTime).Nanoseconds() / 1000000 * tb.stableIntervalMillisecond)
		tb.refreshMutex.Unlock()
	}
	tb.tokens =math.Min(tb.maxBucketCapacity, tb.tokens+added)
	tb.refreshTime = now

}

func (tb *TokenBucket) acquire(tokenNum int) bool{
	tb.refreshTokens()
	if tb.tokens > float64(tokenNum) { // 令牌还有剩余
		tb.tokens -= float64(tokenNum)
		return true
	} else {
		return false
	}
}

