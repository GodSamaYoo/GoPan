package main

import (
	"bytes"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

//首次获取刷新密匙
func FirstGetRefreshToken(tmp *StoreModel) bool {
	url_ := "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	method := "POST"
	payload := strings.NewReader("client_id=" + tmp.ClientID + "&code=" + tmp.Code +
		"&redirect_uri=" + tmp.RedirectUrl + "&grant_type=authorization_code&client_secret=" +
		tmp.ClientSecret)
	client := &http.Client{}
	req, err := http.NewRequest(method, url_, payload)

	if err != nil {
		return false
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false
	}
	RefreshToken_ := gjson.Get(string(body), "refresh_token").Str
	tmp_ := Store{
		Name:         tmp.Name,
		Type:         tmp.Type,
		Path:         tmp.Path,
		Volume:       tmp.Volume,
		RefreshToken: RefreshToken_,
		ClientSecret: tmp.ClientSecret,
		ClientID:     tmp.ClientID,
	}
	if RefreshToken_ != "" {
		if StoreAdd(&tmp_) {
			RefreshAllToken()
			return true
		}
	}
	return false
}

//刷新token
func RefreshToken(id int) error {
	tmp := StoreQuery(&Store{
		ID: id,
	})
	url := "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	method := "POST"

	payload := strings.NewReader("client_id=" + tmp.ClientID + "&grant_type=refresh_token&client_secret=" + tmp.ClientSecret +
		"&refresh_token=" + tmp.RefreshToken)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	Token := gjson.Get(string(body), "access_token").Str
	tmp_ := OneDriveInfo{
		Token: Token,
		Path:  tmp.Path,
	}
	OneDriveTokens[id] = tmp_

	return nil
}

//全部刷新
func RefreshAllToken() {
	var tmp []Store
	db.Where(&Store{
		Type: "onedrive",
	}).Find(&tmp)
	for _, v := range tmp {
		i := 0
		for {
			err := RefreshToken(v.ID)
			if err == nil || i > 4 {
				break
			}
			i++
		}
	}
}

//获得上传地址
func GetOneDriveAdd(email, path, filename string, Need int64) (int, string) {
	Need = Need/1024 + 1
	a := UserQuery(&User{
		Email: email,
	})
	if Need > (a.Volume - a.Used - 1) {
		fmt.Println("用户容量不足")
		return -1, ""
	}
	c := UserGroupQuery(&UserGroup{
		GroupID: a.GroupID,
	})
	d := strings.Split(c.StoreID, ",")
	f := new(Store)
	var e int
	for i, v := range d {
		e, _ = strconv.Atoi(v)
		f = StoreQuery(&Store{
			ID: e,
		})
		if Need > (f.Volume - f.Used - 1) {
			if i == len(d) {
				fmt.Println("储存策略容量不足")
				return -1, ""
			}
			continue
		}
	}
	token := OneDriveTokens[e].Token
	x := 0
	var g string
	for {
		g = UploadAddress(token, f.Path, path, email, filename)
		if g != "" {
			break
		}
		if x > 5 {
			break
		}
		x++
	}
	return e, g
}
func UploadAddress(token, path1, path2, email, filename string) string {
	var url string
	if path2 != "/" {
		url = "https://graph.microsoft.com/v1.0/me/drive/root:" + path1 + "/" + email + path2 + "/" + filename + ":/createUploadSession"
	} else {
		url = "https://graph.microsoft.com/v1.0/me/drive/root:" + path1 + "/" + email + "/" + filename + ":/createUploadSession"
	}
	method := "POST"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return ""
	}
	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ""
	}
	url_ := gjson.Get(string(body), "uploadUrl").Str
	return url_
}

//获取文件下载连链接
func GetOneDriveDownload(itemid string, storeid int) string {
	url := "https://graph.microsoft.com/v1.0/me/drive/items/" + itemid
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("Authorization", "Bearer "+OneDriveTokens[storeid].Token)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	url_ := gjson.Get(string(body), "@microsoft\\.graph\\.downloadUrl").Str
	return url_
}

//重命名文件
func RenameOneDriveFile(itemid string, storeid int, name string) bool {
	url := "https://graph.microsoft.com/v1.0/me/drive/items/" + itemid
	method := "PATCH"
	payload := strings.NewReader(`{"name": "` + name + `"}`)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return false
	}
	req.Header.Add("Authorization", "Bearer "+OneDriveTokens[storeid].Token)
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	id := gjson.Get(string(body), "id").Str
	if id != "" {
		return true
	}
	return false
}

//删除文件
func DeleteOneDriveFile(itemid string, onedriveid int) bool {
	url := "https://graph.microsoft.com/v1.0/me/drive/items/" + itemid
	method := "DELETE"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return false
	}
	req.Header.Add("Authorization", "Bearer "+OneDriveTokens[onedriveid].Token)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer res.Body.Close()

	return true
}

//获取缩略图
func GetThumbnail(itemid string, onedriveid int) string {
	url := "https://graph.microsoft.com/v1.0/me/drive/items/" + itemid + "/thumbnails?select=large"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("Authorization", "Bearer "+OneDriveTokens[onedriveid].Token)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	url_ := gjson.Get(string(body), "value.#(large).large.url").Str
	return url_
}

//本地文件上传  长度,邮箱,绝对路径（带文件名）,网盘路径,文件总体相对路径
func FileUpOneDrive(length int64, email, path1, path2, path3 string) {
	path_ := path.Dir(path1[len(path3):])
	if path2 == "/" {
		if path_ == "." {
			path_ = "/"
		}
	} else {
		if path_ == "." || path_ == "/" {
			path_ = path2
		} else {
			path_ = path2 + path_
		}
	}
	c, d := GetOneDriveAdd(email, path_, path.Base(path1), length)
	if d == "" {
		fmt.Println(path2)
		fmt.Println(path.Base(path1))
		return
	}
	itemid := Aria2OneDriveUp(path1, length, d, c)
	picture := []string{"jpg", "jpeg", "bmp", "gif", "png", "tif"}
	tp := path.Ext(path.Base(path1))
	url_ := ""
	if tp != "" {
		for _, v_ := range picture {
			if tp[1:] == v_ {
				url_ = GetThumbnail(itemid, c)
				break
			}
		}
	}
	fileid := md5_(path.Base(path1) + time.Now().String())
	if url_ != "" {
		os.Mkdir("Thumbnail", 0777)
		file, _ := http.Get(url_)
		defer file.Body.Close()
		files, _ := ioutil.ReadAll(file.Body)
		_ = ioutil.WriteFile("Thumbnail/"+fileid+".jpg", files, 0644)
	}
	a := UserQuery(&User{Email: email})
	DataAdd(&Data{
		FileID:  fileid,
		UserID:  a.UserID,
		Name:    path.Base(path1),
		Type:    "file",
		Path:    path_,
		Size:    int64(math.Floor(float64(length)/1024 + 0.5)),
		StoreID: c,
		ItemID:  itemid,
	})
	UserUpdate(&User{UserID: a.UserID, Used: a.Used + int64(math.Floor(float64(length)/1024+0.5))})
	x := StoreQuery(&Store{
		ID: c,
	})
	StoreUpdate(&Store{
		ID:   x.ID,
		Used: x.Used + int64(math.Floor(float64(length)/1024+0.5)),
	})
}
func Aria2OneDriveUp(filepath string, size int64, url string, storeid int) string {
	f, _ := os.Open(filepath)
	defer f.Close()
	buf := make([]byte, 1024*320*80)
	n := int64(0)
	var status string
	for {
		num, err := f.Read(buf)
		n++
		if err == io.EOF {
			break
		}
		if num < 1024*320*80 {
			buf_ := make([]byte, num)
			buf_ = buf[:num]
			reader := bytes.NewReader(buf_)
			status = OneDriveUp(url, reader, (n-1)*1024*320*80, (n-1)*1024*320*80+int64(num)-1, size, storeid)
			if status == "-1" {
				i := 0
				for {
					status = OneDriveUp(url, reader, (n-1)*1024*320*80, (n-1)*1024*320*80+int64(num)-1, size, storeid)
					if status != "-1" || i == 5 {
						break
					}
					i++
				}
			}
		} else {
			reader := bytes.NewReader(buf)
			status = OneDriveUp(url, reader, (n-1)*1024*320*80, n*1024*320*80-1, size, storeid)
			if status == "-1" {
				i := 0
				for {
					status = OneDriveUp(url, reader, (n-1)*1024*320*80, n*1024*320*80-1, size, storeid)
					if status != "-1" || i == 5 {
						break
					}
					i++
				}
			}
		}
	}
	return gjson.Get(status, "id").Str
}
func OneDriveUp(url string, chunk io.Reader, begin, end, filesize int64, storeid int) string {
	method := "PUT"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, chunk)

	if err != nil {
		fmt.Println(err)
		return "-1"
	}
	begin_ := strconv.FormatInt(begin, 10)
	end_ := strconv.FormatInt(end, 10)
	filesize_ := strconv.FormatInt(filesize, 10)
	req.Header.Add("Content-Range", "bytes "+begin_+"-"+end_+"/"+filesize_)
	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Add("Authorization", "Bearer "+OneDriveTokens[storeid].Token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "-1"
	}
	defer res.Body.Close()

	if res.StatusCode == 202 {
		return "1"
	} else if res.StatusCode == 201 {
		body, _ := ioutil.ReadAll(res.Body)
		return string(body)
	} else {
		return "-1"
	}

}
