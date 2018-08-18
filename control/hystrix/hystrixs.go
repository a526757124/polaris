package hystrix

import (
	"sync"
)

var hystrixs *sync.Map

func init(){
	hystrixs = new(sync.Map)
}

// GetHystrix get Hystrix with catalog name
func GetHystrix(catalog string) (hystrix Hystrix, isInit bool){
	loadHystrix, exists:= hystrixs.Load(catalog)
	if !exists{
		hystrix = NewHystrix(nil, nil)
		isInit = true
	}else{
		hystrix = loadHystrix.(Hystrix)
	}
	return
}

// SetHystrix set Hystrix with catalog name
func SetHystrix(catalog string, hystrix Hystrix) {
	hystrixs.Store(catalog, hystrixs)
}

// ExistsHystrix check is exists Hystrix with catalog name
func ExistsHystrix(catalog string) bool{
	_, exists := hystrixs.Load(catalog)
	return exists
}