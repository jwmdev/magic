package main

import (
	"net/http"

	"magic/routers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.DisableHTTP2 = true
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Debug = true //for production, comment this line
	e.Static("/", "public")

	e.Renderer = Renderer()

	echo.NotFoundHandler = func(c echo.Context) error {
		q := c.QueryParam("q")
		return c.Render(http.StatusNotFound, "404", echo.Map{
			"title": "404 - massaverse",
			"q":     q,
		})
	}

	// associate general routes
	routers.Routers(e)
	//associate web routes
	routers.WebRouters(e)

	//associate apis
	routers.API(e)

	e.Logger.Fatal(e.Start(":8000"))
}
