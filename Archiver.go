package main

import (
	"fmt"
	"github.com/mholt/archiver/v3"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

//解压
func UnArchiver(tmp *UnArchiveFile,email string)  {
	t := UserQuery(&User{
		Email: email,
	})
	a := DataQuery(&Data{FileID: tmp.FileID})
	n :=Task{
		UserID:  t.UserID,
		Path:    tmp.NewPath,
		TmpPath: "",
		Type:    "解压",
		Status:  "正在解压",
	}
	if !IsUserVolume(email,a.Size * 3 / 2 + 1) {
		n.Status = "用户容量不足"
		TaskAdd(&n)
		return
	}
	if !IsLocalVolume (a.Size * 3 / 2 + 1) {
		n.Status = "服务器本地容量不足"
		TaskAdd(&n)
		return
	}
	path_ := md5_(time.Now().String()+tmp.FileID)
	u := path_
	n.TmpPath = path_
	TaskAdd(&n)
	path_ = TmpPath +"/"+path_
	_ = os.Mkdir(path_,0777)
	url := GetOneDriveDownload(a.ItemID,a.StoreID)
	file, _ := http.Get(url)
	defer file.Body.Close()
	f, _ := os.Create(path_+"/"+a.Name)
	buf := make([]byte, 10485760)
	for {
		n_, err := file.Body.Read(buf)
		if err != nil {
			break
		}
		f.Write(buf[:n_])
	}
	if tmp.PassWord == "" {
		err := archiver.Unarchive(path_+"/"+a.Name, path_+"/unarchive")
		if err != nil {
			fmt.Println("解压失败:")
			fmt.Println(err)
			TaskUpdate(&Task{
				TmpPath: u,
				Status:  "解压失败",
			})
			return
		}
	} else if tmp.PassWord != "" || strings.ToLower(path.Ext(a.Name))  == "rar"{
		b := archiver.NewRar()
		b.Password = tmp.PassWord
		err := b.Unarchive(path_+"/"+a.Name, path_+"/unarchive")
		if err != nil {
			fmt.Println("解压失败:")
			fmt.Println(err)
			TaskUpdate(&Task{
				TmpPath: u,
				Status:  "解压失败",
			})
			return
		}
	} else {
		TaskUpdate(&Task{
			TmpPath: u,
			Status:  "解压失败",
		})
		return
	}
	TaskUpdate(&Task{
		TmpPath: u,
		Status:  "解压失败",
	})
	TaskUpdate(&Task{
		TmpPath: u,
		Status:  "正在上传",
	})
	PathFileUpload(path_+"/unarchive",email,tmp.NewPath)
	TaskUpdate(&Task{
		TmpPath: u,
		Status:  "解压成功",
	})
	_ = os.RemoveAll(path_)
	return
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
				FileUpOneDrive(info.Size(),email,ppp,saveph,pppp)
			}
		}
		i++
		return nil
	})
}
