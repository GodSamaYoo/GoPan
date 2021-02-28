package main

import (
	rice "github.com/GeertJohan/go.rice"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/robfig/cron/v3"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	"strconv"
)


func main() {
	e := echo.New()
	e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("hello.godsama.cc")
	e.AutoTLSManager.Cache = autocert.DirCache("hello")
	//e.Pre(middleware.HTTPSRedirect())
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
	assetHandler := http.FileServer(rice.MustFindBox("html").HTTPBox())
	e.GET("/*",echo.WrapHandler(assetHandler))
	e.Use(middleware.CORS())
	e.HideBanner = true
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 3}))
	go e.Logger.Fatal(e.Start(":" + ServicePort))
	e.StartAutoTLS(":443")
}
