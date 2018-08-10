package monitor

import (
	"runtime"
	"sync/atomic"
	"time"
)

var Current *ServerMonitor

func init() {
	Current = &ServerMonitor{
		ServerStartTime:   time.Now(),
		TotalRequestCount: 0,
	}
}

//服务器监控信息
type ServerMonitor struct {
	ServerStartTime   time.Time
	TotalRequestCount uint64
	LastRequestUrls   []string
}

//获取当前Goroutine数量
func (monitor *ServerMonitor) GetGoroutineCount() int {
	return runtime.NumGoroutine()
}

//增加请求数，内部为原子操作
func (monitor *ServerMonitor) AddRequestCount(num uint64) uint64 {
	atomic.AddUint64(&monitor.TotalRequestCount, num)
	return monitor.TotalRequestCount
}
