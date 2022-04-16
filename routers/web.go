package routers

import (
	"magic/controllers"

	"github.com/labstack/echo/v4"
)

// Routers function
func WebRouters(app *echo.Echo) {
	app.GET("/", controllers.Home)
	app.GET("/about", controllers.About)
	app.GET("/terms", controllers.Terms)
	app.GET("/search", controllers.Search) //search
	app.GET("/icon/:ai", controllers.Icon)
}
