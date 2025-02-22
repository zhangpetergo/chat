package mid

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zhangpetergo/chat/chat/foundation/web"
)

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成 uuid 进行日志打印跟踪

		ctx := c.Request.Context()

		traceID := uuid.New()

		ctx = web.SetTraceID(ctx, traceID)

		c.Request = c.Request.WithContext(ctx)

		c.Next()

	}
}
