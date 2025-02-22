package mid

import (
	"github.com/gin-gonic/gin"
	"github.com/zhangpetergo/chat/chat/foundation/logger"
	"github.com/zhangpetergo/chat/chat/foundation/web"
	"go.uber.org/zap"
	"time"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 打印时需要带着uuid

		now := time.Now()

		r := c.Request
		ctx := c.Request.Context()

		traceID := web.GetTraceID(ctx).String()

		log := logger.Log.With(zap.String("uuid", traceID))

		// 程序运行之前打印
		log.Infow("request started", "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr)

		c.Next()

		statusCode := c.Writer.Status()

		log.Infow("request completed", "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr,
			"statuscode", statusCode, "since", time.Since(now).String())

	}
}
