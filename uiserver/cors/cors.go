package cors

import (
	"fmt"
	"github.com/devfeel/dotweb"
	"strings"
)

type CORSMiddleware struct {
	dotweb.BaseMiddlware
}

func (g *CORSMiddleware) Handle(ctx dotweb.Context) error {
	method := ctx.Request().Method
	origin := ctx.Request().Header.Get("Origin")
	var headerKeys []string
	for k, _ := range ctx.Request().Header {
		headerKeys = append(headerKeys, k)
	}
	headerStr := strings.Join(headerKeys, ",")
	if headerStr != "" {
		headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
	} else {
		headerStr = "access-control-allow-origin, access-control-allow-headers"
	}
	if origin != "" {
		ctx.Response().SetHeader("Access-Control-Allow-Origin", "*")
		ctx.Response().SetHeader("Access-Control-Allow-Headers", headerStr)
		ctx.Response().SetHeader("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		ctx.Response().SetHeader("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		ctx.Response().SetHeader("Access-Control-Allow-Credentials", "true")
		ctx.Response().SetHeader("Content-Type", "application/json")
	}
	if method == "OPTIONS" {
		ctx.WriteJson("Options Request!")
	}
	return g.Next(ctx)
}

func NewSimpleCORS() *CORSMiddleware {
	return &CORSMiddleware{}
}
