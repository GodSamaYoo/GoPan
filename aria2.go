package main

import (
	"context"
	"fmt"
	"github.com/zyxar/argo/rpc"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

//下载服务启动
func aria2begin() rpc.Client {
	aria2enable := ReadIni("Aria2", "enable")
	if aria2enable == "no" {
		return nil
	}
	var err error
	aria2token := ReadIni("Aria2", "token")
	port := ReadIni("Aria2", "port")
	aria2url := "http://127.0.0.1:" + port + "/jsonrpc"
	ctx := context.Background()
	var notifier rpc.Notifier = DummyNotifier{}
	t, _ := time.ParseDuration("9999h")
	client, err := rpc.New(ctx, aria2url, aria2token, t, notifier)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return client
}

func (DummyNotifier) OnDownloadStart(events []rpc.Event)      {}
func (DummyNotifier) OnDownloadPause(events []rpc.Event)      {}
func (DummyNotifier) OnDownloadStop(events []rpc.Event)       {}
func (DummyNotifier) OnDownloadError(events []rpc.Event)      {}
func (DummyNotifier) OnBtDownloadComplete(events []rpc.Event) {}
func (DummyNotifier) OnDownloadComplete(events []rpc.Event) {
	for _, v := range events {
		infos, _ := aria2client.TellStatus(v.Gid)
		if len(infos.FollowedBy) == 0 {
			a := TaskQuery(&Task{
				TmpPath: infos.Dir,
			})
			b := UserQuery(&User{
				UserID: a.UserID,
			})
			dir_ := strings.ReplaceAll(infos.Dir, `\`, `/`)
			for _, vv := range infos.Files {
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
					Size:    int64(math.Floor(float64(length)/1024 + 0.5)),
					StoreID: c,
					ItemID:  itemid,
				})
				UserUpdate(&User{
					UserID: a.UserID,
					Used:   b.Used + int64(math.Floor(float64(length)/1024+0.5)),
				})
				x := StoreQuery(&Store{
					ID: c,
				})
				StoreUpdate(&Store{
					ID:   x.ID,
					Used: x.Used + int64(math.Floor(float64(length)/1024+0.5)),
				})
				if path_ != "/" {
					CreateDir(b.Email, path.Dir(path_), path.Base(path_))
				}
			}
			err := os.RemoveAll(infos.Dir)
			if err != nil {
				fmt.Println("移除临时文件失败")
				fmt.Println(err)
			}
		}
	}
}

//开始下载

func aria2download(url []string, path string, userid int) []string {
	var gids []string
	for _, v := range url {
		var url_ string
		var gid string
		var err error
		tmp := TmpPath + "/" + md5_(time.Now().String())
		if len(v) == 40 && !strings.Contains(v, ".") {
			url_ = "magnet:?xt=urn:btih:" + v
			gid, err = aria2client.AddURI([]string{url_}, rpc.Option{"dir": tmp})
		} else if strings.Contains(strings.ToLower(v), "magnet:?xt=urn:btih:") {
			url_ = v
			gid, err = aria2client.AddURI([]string{url_}, rpc.Option{"dir": tmp})
		} else {
			gid, err = aria2client.AddURI([]string{v}, rpc.Option{"dir": tmp})
		}
		if err != nil {
			fmt.Println(err)
			break
		}
		TaskAdd(&Task{
			Gid:     gid,
			UserID:  userid,
			Path:    path,
			TmpPath: tmp,
		})
		gids = append(gids, gid)
	}
	return gids
}

//下载进度查询

func aria2status(userid int) []aria2downloadinfo {
	var infos []aria2downloadinfo
	task := TasksQuery(&Task{
		UserID: userid,
	})
	for _, v := range task {
		totalinfo, _ := aria2client.TellStatus(v.Gid)
		if len(totalinfo.FollowedBy) > 0 && totalinfo.Status == "complete" {
			totalinfo, _ = aria2client.TellStatus(totalinfo.FollowedBy[0])
		}
		var info aria2downloadinfo
		info.Gid = totalinfo.Gid
		info.TotalLength = totalinfo.TotalLength
		info.FileNums = len(totalinfo.Files)
		info.DownloadSpeed = totalinfo.DownloadSpeed
		info.Status = totalinfo.Status
		info.CompletedLength = totalinfo.CompletedLength
		info.BeginTime = v.Time
		info.Path = v.Path
		info.Files = totalinfo.Files
		a := strings.Split(totalinfo.Dir, "/")
		b := strings.Split(totalinfo.Files[0].Path, "/")
		if info.FileNums > 1 {
			info.Name = b[len(a)]
		} else {
			info.Name = b[len(b)-1]
		}
		for i, _ := range info.Files {
			tmp := strings.Split(info.Files[i].Path, "/")
			info.Files[i].Path = tmp[len(tmp)-1]
		}
		infos = append(infos, info)
	}
	return infos
}

//下载任务 暂停 取消 继续

func aria2taskchange(tmp *aria2change) error {
	var err error
	if tmp.Type == 1 {
		_, err = aria2client.ForcePause(tmp.Gid)
		if err != nil {
			return err
		}
	} else if tmp.Type == 2 {
		_, err = aria2client.ForceRemove(tmp.Gid)
	} else if tmp.Type == 3 {
		_, err = aria2client.Unpause(tmp.Gid)
	} else if tmp.Type == 4 {
		_, err = aria2client.Remove(tmp.Gid)
		totalinfo, _ := aria2client.TellStatus(tmp.Gid)
		err = os.RemoveAll(totalinfo.Dir)
		TaskDelete(&Task{
			TmpPath: totalinfo.Dir,
		})
	} else if tmp.Type == 5 {
		totalinfo, _ := aria2client.TellStatus(tmp.Gid)
		err = os.RemoveAll(totalinfo.Dir)
		TaskDelete(&Task{
			TmpPath: totalinfo.Dir,
		})
	}
	if err != nil {
		return err
	}
	return nil
}
