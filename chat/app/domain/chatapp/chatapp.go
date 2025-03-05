package chatapp

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zhangpetergo/chat/chat/app/sdk/chat"
	"github.com/zhangpetergo/chat/chat/app/sdk/errs"

	"github.com/zhangpetergo/chat/chat/foundation/web"
)

type app struct {
	WS   websocket.Upgrader
	Chat *chat.Chat
}

func NewApp() *app {
	return &app{
		Chat: chat.NewChat(),
	}
}

func (a *app) connect(c *gin.Context) {
	ctx := c.Request.Context()

	traceID := web.GetTraceID(c.Request.Context()).String()

	// client connect websocket
	// 升级http协议为websocket协议
	conn, err := a.WS.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Error(errs.Newf(errs.FailedPrecondition, "websocket upgrade failed: %v", err))
		return
	}
	defer conn.Close()

	err = a.Chat.HandleShake(conn, traceID)
	if err != nil {
		c.Error(errs.Newf(errs.FailedPrecondition, "handshake failed: %v", err))
		return
	}

	a.Chat.Listen(ctx, conn, traceID)
	
}

func (a *app) test(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello World",
	})
}

func (a *app) testError(c *gin.Context) {
	c.Error(errs.NewError(errors.New("text error")))
	return
}

func (a *app) testPanic(c *gin.Context) {
	panic("Hello World")
}
