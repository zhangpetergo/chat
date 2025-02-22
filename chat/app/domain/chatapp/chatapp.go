package chatapp

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/zhangpetergo/chat/chat/app/sdk/errs"
)

type app struct {
}

func NewApp() *app {
	return &app{}
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
