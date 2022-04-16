package routers

import (
	"magic/controllers"

	"github.com/labstack/echo/v4"
)

func API(app *echo.Echo) {
	magicv1 := app.Group("/api/v1")
	magicv1.GET("/icon/:ai", controllers.Magic)
}
