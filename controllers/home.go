package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Home(c echo.Context) error {
	data := echo.Map{
		"title":           c.Get("Massaverse home"),
		"hideSmallSearch": true,
	}
	return c.Render(http.StatusOK, "index", data)
}

func About(c echo.Context) error {

	return c.Render(http.StatusOK, "about", echo.Map{
		"title": "About",
	})
}
func Terms(c echo.Context) error {

	return c.Render(http.StatusOK, "terms", echo.Map{
		"title": "Terms, Conditions and Privancy Policy",
	})
}
