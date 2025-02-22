package mid

import (
	"github.com/gin-gonic/gin"
	"github.com/zhangpetergo/chat/chat/app/sdk/errs"
	"runtime/debug"
)

func Panics() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				trace := debug.Stack()

				err := errs.Newf(errs.InternalOnlyLog, "PANIC [%v] TRACE[%s]", rec, string(trace))

				c.Error(err)
				c.Abort()

			}
		}()
		c.Next()
	}
}
