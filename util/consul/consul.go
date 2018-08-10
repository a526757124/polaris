package consul

import (
	consulapi "github.com/hashicorp/consul/api"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type ServiceConfig struct {
	Name string
	Tags []string
	Address string
	Port int
	ChechUrl string
}


// FindService find services from consul server with serviceName and tag
func FindService(addr string, serviceName string, tag string)([]*ServiceConfig, error){
	config := consulapi.DefaultConfig()
	config.Address = addr
	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.Catalog().Services(&consulapi.QueryOptions{
		Datacenter:"dc1",
	})
	services, _, err:=client.Catalog().Service(serviceName, tag, nil)
	if err != nil{
		return nil, err
	}

	var appServices []*ServiceConfig
	for _, v:=range services{
		appServices = append(appServices, &ServiceConfig{
			Name:v.ServiceName,
			Tags:v.ServiceTags,
			Address:v.ServiceAddress,
			Port:v.ServicePort,
		})
	}
	return appServices, nil
}

// RegisteService register service to consul server with service config info
func RegisteService(addr string, service *ServiceConfig) error{
	config := consulapi.DefaultConfig()
	config.Address = addr

	client, err := consulapi.NewClient(config)
	if err != nil {
		return err
	}
	//创建一个新服务。
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = hashService(service)
	registration.Name = service.Name
	registration.Port = service.Port
	registration.Tags = service.Tags
	registration.Address = service.Address

	registration.Check = &consulapi.AgentServiceCheck{
		HTTP:                           service.ChechUrl,
		Timeout:                        "3s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "30s", //check失败后30秒删除本服务
	}
	err = client.Agent().ServiceRegister(registration)
	return err
}

// hashService hash service to string
func hashService(service *ServiceConfig) string {
	data, err := json.Marshal(service)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}