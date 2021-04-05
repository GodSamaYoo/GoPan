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
	DesKey = ReadIni("Des", "key")
	TmpPath = ReadIni("TmpFile", "path")
	TmpVolume, _ = strconv.ParseInt(ReadIni("TmpFile", "volume"), 10, 64)
	OneDriveTokens = make(map[int]OneDriveInfo)

	//刷新OneDrive Token
	go RefreshAllToken()
	c := cron.New()
	_, _ = c.AddFunc("*/50 * * * *", RefreshAllToken)
	//每日凌晨定时更新Tracker
	go GetLatestTracker()
	_, _ = c.AddFunc("0 0 * * *", GetLatestTracker)
	c.Start()
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "html",
		Browse: true,
		HTML5:  true,
	}))
	e.Use(middleware.CORS())
	e.HideBanner = true
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 3}))
	//e.Pre(middleware.HTTPSRedirect())
	/*e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
	e.Logger.Fatal(e.StartAutoTLS(":" + ServicePort))*/
	e.Logger.Fatal(e.Start(":" + ServicePort))
}
