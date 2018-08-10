package cache

import "sync"

var (
	runtime_cache  Cache
	runtimeCacheLock *sync.RWMutex
)

func init() {
	runtimeCacheLock = new(sync.RWMutex)
}

type Cache interface {
	// Exist return true if value cached by given key
	Exists(key string) (bool, error)
	// Get returns value by given key
	Get(key string) (interface{}, error)
	// GetString returns value string format by given key
	GetString(key string) (string, error)
	// GetInt returns value int format by given key
	GetInt(key string) (int, error)
	// GetInt64 returns value int64 format by given key
	GetInt64(key string) (int64, error)
	// Set cache value by given key
	Set(key string, v interface{}, ttl int64) error
	// Incr increases int64-type value by given key as a counter
	// if key not exist, before increase set value with zero
	Incr(key string) (int64, error)
	// Decr decreases int64-type value by given key as a counter
	// if key not exist, before increase set value with zero
	Decr(key string) (int64, error)
	// Delete delete cache item by given key
	Delete(key string) error
	// ClearAll clear all cache items
	ClearAll() error
	// Expire Set a timeout on key. After the timeout has expired, the key will automatically be deleted.
	// timeout time duration is second
	Expire(key string, timeOutSeconds int) (int, error)
}



// GetCache get default cache
func GetCache() Cache {
	if runtime_cache == nil {
		runtimeCacheLock.Lock()
		if runtime_cache == nil{
			runtime_cache = NewRuntimeCache()
		}
		runtimeCacheLock.Unlock()
	}
	return runtime_cache
}
