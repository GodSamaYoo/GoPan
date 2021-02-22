package main

import (
	"github.com/zyxar/argo/rpc"
	"gorm.io/gorm"
)

var db gorm.DB
var DesKey string
var TmpPath string
var OneDriveTokens map[int]OneDriveInfo
var aria2client rpc.Client
var aria2path string