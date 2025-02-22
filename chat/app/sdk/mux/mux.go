package mux

import (
	"github.com/gin-gonic/gin"
	"github.com/zhangpetergo/chat/chat/app/domain/chatapp"
	"github.com/zhangpetergo/chat/chat/app/sdk/mid"
	"net/http"
)

func WebAPI() http.Handler {
	app := gin.New()

	// add mid
	app.Use(mid.Logger(), mid.Errors(), mid.Panics())

	// add route
	chatapp.Routes(app)

	return app
}
