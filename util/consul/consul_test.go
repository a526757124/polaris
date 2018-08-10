package consul

import "testing"

func TestFindService(t *testing.T) {
	consulServer:= "192.168.240.70:8500"
	serviceName:="consul"
	tag := ""
	services, err:= FindService(consulServer, serviceName, tag)
	if err != nil{
		t.Error("FindService error", err)
	}else{
		t.Log("FindService success", services)
	}
}
