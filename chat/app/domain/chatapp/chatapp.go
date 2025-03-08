package chatapp

import (
	"github.com/gin-gonic/gin"
	"github.com/zhangpetergo/chat/chat/app/sdk/chat"
	"github.com/zhangpetergo/chat/chat/app/sdk/errs"
)

type app struct {
	Chat *chat.Chat
}

func NewApp() *app {
	return &app{
		Chat: chat.NewChat(),
	}
}

func (a *app) connect(c *gin.Context) {
	ctx := c.Request.Context()

	usr, err := a.Chat.HandleShake(ctx, c.Writer, c.Request)
	if err != nil {
		c.Error(errs.Newf(errs.FailedPrecondition, "handshake failed: %v", err))
		return
	}

	defer usr.Conn.Close()

	a.Chat.Listen(ctx, usr)

}

func (a *app) test(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello World",
	})
}

func (a *app) testError(c *gin.Context) {
	c.Error(errs.Newf(errs.Internal, "text error"))
	return
}

func (a *app) testPanic(c *gin.Context) {
	panic("Hello World")
}
