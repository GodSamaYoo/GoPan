package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/robfig/cron/v3"
	"strconv"
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
	TmpVolume,_ = strconv.ParseInt(ReadIni("TmpFile", "volume"),10,64)
	OneDriveTokens = make(map[int]OneDriveInfo)
	go RefreshAllToken()
	c := cron.New()
	_, _ = c.AddFunc("*/50 * * * *", RefreshAllToken)
	c.Start()
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "html",
		Browse: true,
		HTML5: true,
	}))
	e.Use(middleware.CORS())
	e.HideBanner = true
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 3}))
	e.Pre(middleware.HTTPSRedirect())
	go func() {
		e.Logger.Fatal(e.Start(":"+ServicePort))
	}()
	e.Logger.Fatal(e.StartTLS(":443", "crt/server.crt", "crt/server.key"))
}
