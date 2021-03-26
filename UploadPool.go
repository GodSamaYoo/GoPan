package main

import (
	"fmt"
	"github.com/zyxar/argo/rpc"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type UploadTasks struct {
	infos rpc.StatusInfo
}

func (t *UploadTasks) Execute() {
	a := TaskQuery(&Task{
		TmpPath: t.infos.Dir,
	})
	TaskUpdate(&Task{
		TmpPath: t.infos.Dir,
		Status:  "上传中",
	})
	dir_ := strings.ReplaceAll(t.infos.Dir, `\`, `/`)
	for _, vv := range t.infos.Files {
		b := UserQuery(&User{
			UserID: a.UserID,
		})
		length, _ := strconv.ParseInt(vv.Length, 10, 64)
		path_ := path.Dir(vv.Path[len(dir_):])
		if a.Path == "/" {
			if path_ == "." {
				path_ = "/"
			}
		} else {
			if path_ == "." || path_ == "/" {
				path_ = a.Path
			} else {
				path_ = a.Path + path_
			}
		}
		c, d := GetOneDriveAdd(b.Email, path_, path.Base(vv.Path), length)
		if d == "" {
			return
		}
		itemid := Aria2OneDriveUp(vv.Path, length, d, c)
		picture := []string{"jpg", "jpeg", "bmp", "gif", "png", "tif"}
		tp := path.Ext(path.Base(vv.Path))
		url_ := ""
		if tp != "" {
			for _, v_ := range picture {
				if tp[1:] == v_ {
					url_ = GetThumbnail(itemid, c)
					break
				}
			}
		}
		fileid := md5_(path.Base(vv.Path) + time.Now().String())
		if url_ != "" {
			os.Mkdir("Thumbnail", 0777)
			file, _ := http.Get(url_)
			defer file.Body.Close()
			files, _ := ioutil.ReadAll(file.Body)
			_ = ioutil.WriteFile("Thumbnail/"+fileid+".jpg", files, 0644)
		}
		DataAdd(&Data{
			FileID:  fileid,
			UserID:  a.UserID,
			Name:    path.Base(vv.Path),
			Type:    "file",
			Path:    path_,
			Size:    length,
			StoreID: c,
			ItemID:  itemid,
		})
		UserUpdate(&User{
			UserID: a.UserID,
			Used:   b.Used + length,
		})
		x := StoreQuery(&Store{
			ID: c,
		})
		StoreUpdate(&Store{
			ID:   x.ID,
			Used: x.Used + length,
		})
		if path_ != "/" {
			CreateDir(b.Email, path.Dir(path_), path.Base(path_))
		}
		_ = os.Remove(vv.Path)
	}
	TaskUpdate(&Task{
		TmpPath: t.infos.Dir,
		Status:  "上传成功",
	})
	err := os.RemoveAll(t.infos.Dir)
	if err != nil {
		fmt.Println("移除临时文件失败")
		fmt.Println(err)
	}
}

type UploadPool struct {
	WorkNum     int
	JobsChannel chan *UploadTasks
}

func (p *UploadPool) Worker() {
	for a := range p.JobsChannel {
		a.Execute()
	}
}

func (p *UploadPool) Run() {
	for i := 0; i < p.WorkNum; i++ {
		go p.Worker()
	}
}
func NewUploadPool(num int) *UploadPool {
	p := UploadPool{
		WorkNum:     num,
		JobsChannel: make(chan *UploadTasks, 100),
	}
	return &p
}
