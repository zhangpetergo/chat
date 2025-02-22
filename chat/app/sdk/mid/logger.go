package mid

import (
	"github.com/gin-gonic/gin"
	"github.com/zhangpetergo/chat/chat/foundation/logger"
	"time"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {

		now := time.Now()

		r := c.Request

		// 程序运行之前打印
		logger.Log.Infow("request started", "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr)

		c.Next()

		statusCode := c.Writer.Status()

		logger.Log.Infow("request completed", "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr,
			"statuscode", statusCode, "since", time.Since(now).String())

	}
}
