package chatapp

import "github.com/gin-gonic/gin"

func Routes(app *gin.Engine) {
	api := NewApp()

	app.GET("/test", api.test)
	app.GET("/testerror", api.testError)
	app.GET("/testpanic", api.testPanic)
}
