package routers

import (
	"github.com/labstack/echo/v4"
)

// Routers function
func Routers(app *echo.Echo) {

	//Pre route

	//some settings
	app.Static("/css", "./public/css")
	app.Static("/vendors", "./public/vendors")
	app.Static("/images", "./public/img")
	app.Static("/js", "./public/js")
	app.Static("/icons", "./.storage/images/icons")
}
