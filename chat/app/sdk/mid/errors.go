package mid

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/zhangpetergo/chat/chat/app/sdk/errs"
	"github.com/zhangpetergo/chat/chat/foundation/logger"
	"github.com/zhangpetergo/chat/chat/foundation/web"
	"go.uber.org/zap"
	"path"
)

func Errors() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()

		ctx := c.Request.Context()

		// 程序处理完毕
		// 判断是否存在错误
		if len(c.Errors) > 0 {
			// 处理第一个错误
			// 在 gin 中，错误是一个数组，这里只处理第一个错误，一般来说我们在程序中遇到错误时，只会返回一个错误
			// 如果出现了例外情况，那么我们需要修改这里的代码
			err := c.Errors[0].Err

			var appErr *errs.Error
			// 如果不是我们自定义错误，代表服务器发生了一些我们不知道的错误，所以这里返回 500
			if !errors.As(err, &appErr) {
				appErr = errs.Newf(errs.Internal, "Internal Server Error")
			}

			traceID := web.GetTraceID(ctx).String()

			log := logger.Log.With(zap.String("uuid", traceID))

			log.Errorw("handled error during request",
				"err", err,
				"source_err_file", path.Base(appErr.FileName),
				"source_err_func", path.Base(appErr.FuncName))

			if appErr.Code == errs.InternalOnlyLog {
				appErr = errs.Newf(errs.Internal, "Internal Server Error")
			}

			c.JSON(appErr.HTTPStatus(), appErr)

			c.Abort()
			return
		}

	}
}
