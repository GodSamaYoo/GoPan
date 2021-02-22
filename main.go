package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/robfig/cron/v3"
)


func main() {
	e := echo.New()
	CheckIni()
	CheckSqlite()
	RegisterRoutes(e)
	aria2client = aria2begin()
	ServicePort := ReadIni("Service", "port")
	DesKey = ReadIni("Des","key")
	TmpPath = ReadIni("TmpFile", "path")
	OneDriveTokens = make(map[int]OneDriveInfo)
	go RefreshAllToken()
	c := cron.New()
	_, _ = c.AddFunc("*/50 * * * *", RefreshAllToken)
	c.Start()
	e.Use(middleware.CORS())
	e.HideBanner = true
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 3}))
	e.Logger.Fatal(e.Start(":" + ServicePort))
}
