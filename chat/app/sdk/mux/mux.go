package mux

import (
	"github.com/gin-gonic/gin"
	"github.com/zhangpetergo/chat/chat/app/domain/chatapp"
	"github.com/zhangpetergo/chat/chat/app/sdk/mid"
	"net/http"
)

// WebAPI 返回一个 http.Handler，用于设置带有中间件和路由的 Gin 引擎。
func WebAPI() http.Handler {
	app := gin.New()

	// add mid
	app.Use(mid.TraceID(), mid.Logger(), mid.Errors(), mid.Panics())

	// add route
	chatapp.Routes(app)

	return app
}
