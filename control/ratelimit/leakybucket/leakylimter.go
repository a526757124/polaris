package leakybucket

import "sync"

type LeakyLimiter struct {
	bucketMap map[string]*LeakyBucket
	bucketMutex *sync.RWMutex
}

func NewLeakyLimiter() *LeakyLimiter{
	return &LeakyLimiter{
		bucketMap:make(map[string]*LeakyBucket),
		bucketMutex:new(sync.RWMutex),
	}
}

func (l *LeakyLimiter) getLeakyBucket(key string) *LeakyBucket{
	l.bucketMutex.RLock()
	tb, exists :=l.bucketMap[key]
	l.bucketMutex.RUnlock()
	if !exists{
		l.bucketMutex.Lock()
		defer l.bucketMutex.RUnlock()
		tb = newLeakyBucket()
		l.bucketMap[key] = tb
	}
	return tb
}

func (l *LeakyLimiter) RequestCheck(key string, calls int) bool{
	tb := l.getLeakyBucket(key)
	if tb == nil{
		return true
	}
	return tb.acquire(calls)
}
