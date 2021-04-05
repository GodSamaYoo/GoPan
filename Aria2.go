package main

import (
	"context"
	"fmt"
	"github.com/zyxar/argo/rpc"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

//下载服务启动
func aria2begin() rpc.Client {
	aria2enable := ReadIni("Aria2", "enable")
	if aria2enable == "no" {
		return nil
	}
	if UploadPools == nil {
		UploadPools = NewUploadPool(8)
		UploadPools.Run()
	}
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
		info, _ := aria2client.TellStatus(v.Gid)
		if len(info.FollowedBy) == 0 {
			UploadPools.JobsChannel <- &UploadTasks{infos: info}
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
			gid, err = aria2client.AddURI([]string{url_}, rpc.Option{"dir": tmp, "bt-tracker": Tracker})
		} else if strings.Contains(strings.ToLower(v), "magnet:?xt=urn:btih:") {
			url_ = v
			gid, err = aria2client.AddURI([]string{url_}, rpc.Option{"dir": tmp, "bt-tracker": Tracker})
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
			Type:    "下载",
			Status:  "下载中",
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
		Type:   "下载",
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

//获取最新的Tracker
func GetLatestTracker() {
	url := "https://trackerslist.com/all_aria2.txt"
	res, _ := http.Get(url)
	body, _ := ioutil.ReadAll(res.Body)
	Tracker = string(body)
}
