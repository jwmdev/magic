package controllers

import (
	"net/http"

	"github.com/k0kubun/pp/v3"
	"github.com/labstack/echo/v4"
)

func Search(c echo.Context) error {
	q := c.QueryParam("q")
	qLen := len(q)
	pp.Printf("length = %v\n", qLen)

	if q == "massaverse" || q == "meta watch" {
		return c.Redirect(http.StatusMovedPermanently, "/")
	}

	if qLen > 40 && qLen < 60 {
		return c.Redirect(http.StatusMovedPermanently, "/icon/"+q)
	}

	return echo.NotFoundHandler(c)
}
