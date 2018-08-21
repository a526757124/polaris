package cache

import "sync"

const(
	RedisConnPool_MaxIdle = 20
	RedisConnPool_MaxActive = 100
)
)

var (
	runtime_cache  Cache
	redisCacheMap  map[string]RedisCache
	redisCacheLock *sync.RWMutex
)

func init() {
	redisCacheMap = make(map[string]RedisCache)
	redisCacheLock = new(sync.RWMutex)
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
	// Expire Set a timeout on key. After the timeout has expired, the key will automatically be deleted.
	// timeout time duration is second
	Expire(key string, timeOutSeconds int) (int, error)
}

type RedisCache interface {
	Cache

	// GetJsonObj get obj with SetJsonObj key
	GetJsonObj(key string, result interface{}) error
	// SetJsonObj set obj use json encode string
	SetJsonObj(key string, val interface{}) (interface{}, error)


	// SetBackupServer set backup redis server
	SetBackupServer(serverUrl string, maxIdle int, maxActive int)
	/*---------- Hash -----------*/
	// HGet Returns the value associated with field in the hash stored at key.
	HGet(hashID string, field string) (string, error)
	// HMGet Returns the values associated with the specified fields in the hash stored at key.
	HMGet(hashID string, field ...interface{}) ([]string, error)
	// HSet Sets field in the hash stored at key to value. If key does not exist, a new key holding a hash is created.
	// If field already exists in the hash, it is overwritten.
	HSet(hashID string, field string, val string) error
	// HSetNX Sets field in the hash stored at key to value, only if field does not yet exist
	HSetNX(hashID string, field string, val string) (string, error)
	// HDel Removes the specified fields from the hash stored at key.
	HDel(hashID string, fields ...interface{}) (int, error)
	// HExists Returns if field is an existing field in the hash stored at key
	HExists(hashID string, field string) (int, error)
	// HIncrBy Increments the number stored at field in the hash stored at key by increment.
	HIncrBy(hashID string, field string, increment int) (int, error)
	// HKeys Returns all field names in the hash stored at key.
	HKeys(hashID string) ([]string, error)
	// HLen Returns the number of fields contained in the hash stored at key
	HLen(hashID string) (int, error)

	/*---------- List -----------*/
	// BRPOP is a blocking list pop primitive
	// It is the blocking version of RPOP because it blocks the connection when there are no elements to pop from any of the given lists
	BRPop(key ...interface{}) (map[string]string, error)
	// LLen return length of list
	LLen(key string) (int, error)
	// LPush Insert all the specified values at the head of the list stored at key
	LPush(key string, value ...interface{}) (int, error)
}


//get runtime cache
func GetRuntimeCache() Cache {
	if runtime_cache == nil {
		runtime_cache = NewRuntimeCache()
	}
	return runtime_cache
}

//get redis cache
//must set serverUrl like "redis://:password@10.0.1.11:6379/0"
func GetRedisCache(serverUrl string, backServerUrl string, maxIdle int, maxActive int) RedisCache {
	if maxIdle <= 0{
		maxIdle = RedisConnPool_MaxIdle
	}
	if maxActive <= 0{
		maxActive = RedisConnPool_MaxActive
	}
	c, ok := redisCacheMap[serverUrl]
	if !ok {
		c = NewRedisCache(serverUrl, maxIdle, maxActive)
		c.SetBackupServer(backServerUrl, maxIdle, maxActive)
		redisCacheLock.Lock()
		redisCacheMap[serverUrl] = c
		redisCacheLock.Unlock()
		return c

	} else {
		return c
	}
}