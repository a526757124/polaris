package tokenbucket

import "sync"

type TokenLimiter struct {
	bucketMap map[string]*TokenBucket
	bucketMutex *sync.RWMutex
}

func NewTokenLimiter() *TokenLimiter{
	return &TokenLimiter{
		bucketMap:make(map[string]*TokenBucket),
		bucketMutex:new(sync.RWMutex),
	}
}

func (l *TokenLimiter) getTokenBucket(key string) *TokenBucket{
	l.bucketMutex.RLock()
	tb, exists :=l.bucketMap[key]
	l.bucketMutex.RUnlock()
	if !exists{
		l.bucketMutex.Lock()
		defer l.bucketMutex.RUnlock()
		tb = newTokenBucket()
		l.bucketMap[key] = tb
	}
	return tb
}

func (l *TokenLimiter) RequestCheck(key string, calls int) bool{
	tb := l.getTokenBucket(key)
	if tb == nil{
		return true
	}
	return tb.acquire(calls)
}