package main

import (
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"time"
)

const IniPath = "./GodCloud.ini"

func CheckIni() {
	_, err := os.Stat(IniPath)
	if err != nil || os.IsNotExist(err) {
		CreateIni(IniPath)
		cfg := OpenIni(IniPath)
		if cfg != nil {
			_, _ = cfg.Section("TmpFile").NewKey("path", "TmpFile")
			_, _ = cfg.Section("TmpFile").NewKey("volume", string(1024*1024))
			_, _ = cfg.Section("Aria2").NewKey("enable", "no")
			_, _ = cfg.Section("Service").NewKey("port", "2020")
			_, _ = cfg.Section("Des").NewKey("key", md5_(time.Now().String())[8:16])
			_ = os.Mkdir("TmpFile", 0777)
			err = cfg.SaveTo(IniPath)
		}
	}
}

func OpenIni(path string) *ini.File {
	cfg, err := ini.Load(path)
	if err != nil {
		fmt.Println("配置读取失败")
	} else {
		return cfg
	}
	return nil
}

func CreateIni(path string) {
	_, err := os.Create(path)
	if err != nil {
		fmt.Println("创建失败")
	}
}

func ReadIni(Section string, Key string) string {
	cfg := OpenIni("./GodCloud.ini")
	text := cfg.Section(Section).Key(Key).String()
	return text
}

func ModifyIni(Section, Key ,value string) {
	cfg := OpenIni(IniPath)
	cfg.Section(Section).Key(Key).SetValue(value)
	_ = cfg.SaveTo(IniPath)
}
