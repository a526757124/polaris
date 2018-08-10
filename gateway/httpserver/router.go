// httprouter
package httpserver

import (

	"github.com/devfeel/dotweb"
	"github.com/devfeel/polaris/gateway/httpserver/handlers"
)

func InitRoute(dotweb *dotweb.DotWeb) {
	dotweb.HttpServer.Router().GET("/api/:module/:version/:apikey", handlers.ProxyGet)
	dotweb.HttpServer.Router().POST("/api/:module/:version/:apikey", handlers.ProxyPost)
	dotweb.HttpServer.Router().GET("/local/:module/:version/:apikey", handlers.ProxyLocal)
	dotweb.HttpServer.Router().GET("/", handlers.Index)
	dotweb.HttpServer.Router().GET("/monitor", handlers.Monitor)
	dotweb.HttpServer.Router().GET("/pprof/:module", handlers.Watch)
	dotweb.HttpServer.Router().GET("/info/:infotype", handlers.Info)
	dotweb.HttpServer.Router().GET("/version", handlers.Version)
}
