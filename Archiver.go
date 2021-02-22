package main

import (
	"fmt"
	"github.com/mholt/archiver/v3"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

//解压缩
func UnArchiver(tmp *UnArchiveFile,email string) bool {
	path_ := md5_(time.Now().String())
	path_ = TmpPath +"/"+path_
	_ = os.Mkdir(path_,0777)
	a := DataQuery(&Data{FileID: tmp.FileID})
	url := GetOneDriveDownload(a.ItemID,a.StoreID)
	file, _ := http.Get(url)
	defer file.Body.Close()
	files,_ := ioutil.ReadAll(file.Body)
	_ = ioutil.WriteFile(path_+"/"+a.Name, files, 0644)
	if tmp.PassWord == "" {
		err := archiver.Unarchive(path_+"/"+a.Name, path_+"/unarchive")
		if err != nil {
			fmt.Println("解压失败:")
			fmt.Println(err)
			return false
		}
	} else if tmp.PassWord != "" || strings.ToLower(path.Ext(a.Name))  == "rar"{
		b := archiver.NewRar()
		b.Password = tmp.PassWord
		err := b.Unarchive(path_+"/"+a.Name, path_+"/unarchive")
		if err != nil {
			fmt.Println("解压失败:")
			fmt.Println(err)
			return false
		}
	} else {
		return false
	}
	PathFileUpload(path_+"/unarchive",email,tmp.NewPath)
	_ = os.RemoveAll(path_)
	return true
}

//目录文件上传

func PathFileUpload(path_ string,email string,saveph string)  {
	i := 0
	_ = filepath.Walk(path_, func(paths string, info os.FileInfo, err error) error {
		if i != 0 {
			lens := len(path_)
			NowPath := string([]rune(paths)[lens+1:])
			PathSlice := strings.Split(NowPath, `\`)
			if info.IsDir() {
				pathh := saveph
				if len(PathSlice) > 1 {
					if saveph == "/" {
						pathh = ""
					}
					for q, v := range PathSlice {
						if q == len(PathSlice)-1 {
							break
						}
						pathh += "/" + v
					}
				}
				CreateDir(email,pathh,info.Name())
			}else {
				ppp,_ := filepath.Abs(paths)
				ppp = strings.ReplaceAll(ppp,`\`,`/`)
				pppp,_ := filepath.Abs(path_)
				pppp = strings.ReplaceAll(pppp,`\`,`/`)
				FileUpOneDrive(int(info.Size()),email,ppp,saveph,pppp)
			}
		}
		i++
		return nil
	})
}
