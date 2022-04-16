package controllers

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/k0kubun/pp/v3"
	"github.com/labstack/echo/v4"
)

func Magic(c echo.Context) error {
	ai := c.Param("ai")

	path := fmt.Sprintf("%s%s.png", baseIconPath, ai)
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

	return c.File(path)
}
