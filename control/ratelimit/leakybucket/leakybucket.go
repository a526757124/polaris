package leakybucket

import (
	"time"
	"math"
	"sync"
)

type LeakyBucket struct {
	rate float64             // lost water per Millisecond
	water float64            // current water in bucket
	bucketSize float64       // max bucket size
	refreshTime time.Time    // time for last water refresh
	refreshMutex *sync.Mutex // refresh mutex
}

func newLeakyBucket() *LeakyBucket{
	return &LeakyBucket{
		rate : 0.2, //每毫秒流失
		water : 0,
		bucketSize : 10,
		refreshTime : time.Now(),
		refreshMutex : new(sync.Mutex),
	}
}

func (lb *LeakyBucket) refreshWater() {
	now := time.Now()
	lb.refreshMutex.Lock()
	lost := float64(now.Sub(lb.refreshTime).Nanoseconds()/1000/1000) * lb.rate
	lb.refreshMutex.Unlock()
	lb.water =math.Max(0, lb.water-lost)
	lb.refreshTime = now
}

func (lb *LeakyBucket) acquire(tokenNum int) bool{
	lb.refreshWater()
	if lb.water < lb.bucketSize { //存量是否小于桶容量
		lb.water += float64(tokenNum)
		return true
	} else {
		return false
	}
}
