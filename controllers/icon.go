package controllers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/k0kubun/pp/v3"
	"github.com/labstack/echo/v4"
)

const baseIconPath = ".storage/images/icons/"

func Icon(c echo.Context) error {
	ai := c.Param("ai")

	path := fmt.Sprintf("%s%s.svg", baseIconPath, ai)
	pp.Printf("path: %v\n", path)

	if _, err := os.Stat(path); err != nil {
		pp.Errorf(err.Error())
		pp.Println("generating image which does not exist")
		cmd := exec.Command("./magic", ai)
		err := cmd.Run()
		if err != nil {
			pp.Printf("error generating image: %v\n", err)
			return echo.NotFoundHandler(c)
		}
		pp.Println("image generated")

	}

	data := echo.Map{
		"title":   c.Get("Massaverse address icon"),
		"icon":    fmt.Sprintf("%s.svg", ai),
		"address": ai,
	}
	return c.Render(http.StatusOK, "icon", data)

	//return echo.NotFoundHandler(c)
}
