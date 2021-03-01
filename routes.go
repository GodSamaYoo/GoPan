package main

import (
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
)

func RegisterRoutes(e *echo.Echo) {

	//获取目录中文件
	e.GET("/api/file/:path", func(ctx echo.Context) error {
		path, _ := url.QueryUnescape(ctx.Param("path"))
		EnKey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(EnKey.Value,DesKey)
		return ctx.JSON(200, GetPathData(email, path))
	})

	//文件夹创建
	e.POST("/api/dir", func(ctx echo.Context) error {
		path,_ := url.QueryUnescape(ctx.FormValue("path"))
		name := ctx.FormValue("name")
		EnKey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(EnKey.Value,DesKey)
		if CreateDir(email,path,name) {
			return ctx.JSON(200,"succeed")
		}
		return ctx.JSON(200,"failed")
	})

	//删除文件
	e.DELETE("/api/files/:id", func(ctx echo.Context) error {
		id := ctx.Param("id")
		EnKey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(EnKey.Value,DesKey)
		a := DataQuery(&Data{
			FileID: id,
		})
		b := UserQuery(&User{
			UserID: a.UserID,
		})
		if email != b.Email {
			return ctx.JSON(200,"failed")
		}
		if DeleteOneDriveFile(a.ItemID,a.StoreID) {
			UserUpdate(&User{UserID: b.UserID,Used: b.Used-a.Size})
			x:=StoreQuery(&Store{
				ID: a.StoreID,
			})
			StoreUpdate(&Store{
				ID:   x.ID,
				Used: x.Used - a.Size,
			})
			DataDelete(&Data{
				FileID: id,
			})
			return ctx.JSON(200,"succeed")
		}
		return ctx.JSON(200,"failed")
	})

	//删除文件夹
	e.DELETE("/api/dir/:id", func(ctx echo.Context) error {
		id := ctx.Param("id")
		EnKey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(EnKey.Value,DesKey)
		a := DataQuery(&Data{
			FileID: id,
		})
		b := UserQuery(&User{
			UserID: a.UserID,
		})
		if email != b.Email {
			return ctx.JSON(200,"failed")
		}
		go DirDelete(id)
		return ctx.JSON(200,"succeed")
	})

	//文件获取下载
	e.GET("/api/filedown/:id", func(ctx echo.Context) error {
		fileid := ctx.Param("id")
		if fileid == "" {
			return nil
		}
		EnKey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		_,err = DesDecrypt(EnKey.Value,DesKey)
		if err != nil {
			return err
		}
		tmp := DataQuery(&Data{
			FileID: fileid,
		})
		q := GetOneDriveDownload(tmp.ItemID,tmp.StoreID)
		if q != "" {
			return ctx.Redirect(302,q)
		}else {
			return ctx.JSON(200,"failed")
		}
	})

	//文件缩略图
	e.GET("/api/thumbnail/:fileid", func(ctx echo.Context) error {
		fileid := ctx.Param("fileid")
		return ctx.Attachment("Thumbnail/"+fileid+".jpg",fileid+".jpg")
	})

	//onedrive上传地址获取
	e.POST("/api/uploadaddress", func(ctx echo.Context) error {
		EnKey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(EnKey.Value,DesKey)
		path := ctx.FormValue("path")
		filename := ctx.FormValue("filename")
		filesize,_ := strconv.ParseInt(ctx.FormValue("filesize"),10,64)
		tmp := OnedriveAddReturn{}
		tmp.StoreID,tmp.Address = GetOneDriveAdd(email,path,filename,filesize)
		if tmp.StoreID != -1 {
			return ctx.JSON(200,tmp)
		}
		return ctx.JSON(200,"failed")
	})

	//onedrive上传成功回调
	e.POST("/api/onedrivestoredata", func(ctx echo.Context) error {
		EnKey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(EnKey.Value,DesKey)
		tmp := new(OnedriveDatas)
		_ = ctx.Bind(tmp)
		picture := []string{"jpg", "jpeg", "bmp", "gif", "png", "tif"}
		tp := path.Ext(tmp.Name)
		url_ := ""
		if tp != "" {
			for _,v_ := range picture{
				if tp[1:] == v_{
					url_ = GetThumbnail(tmp.ItemID,tmp.StoreID)
					break
				}
			}
		}
		fileid := md5_(tmp.Name + time.Now().String())
		if url_ != "" {
			_ = os.Mkdir("Thumbnail", 0777)
			file, _ := http.Get(url_)
			defer file.Body.Close()
			files,_ := ioutil.ReadAll(file.Body)
			_ = ioutil.WriteFile("Thumbnail/"+fileid+".jpg", files, 0644)
		}
		user_ := UserQuery(&User{
			Email: email,
		})
		DataAdd(&Data{
			FileID:    fileid,
			Name:      tmp.Name,
			Type:      tmp.Type,
			Path:      tmp.Path,
			UserID:    user_.UserID,
			Size:      tmp.Size,
			StoreID:   tmp.StoreID,
			ItemID:    tmp.ItemID,
		})
		UserUpdate(&User{UserID: user_.UserID,Used: user_.Used+tmp.Size})
		x:=StoreQuery(&Store{
			ID: tmp.StoreID,
		})
		StoreUpdate(&Store{
			ID:   tmp.StoreID,
			Used: x.Used + tmp.Size,
		})
		return ctx.JSON(200,"succeed")
	})

	//重命名文件(夹)

	e.PUT("/api/file", func(ctx echo.Context) error {
		tmp := new(UpdateType)
		_ = ctx.Bind(tmp)
		if RenameData(tmp) {
			return ctx.JSON(200, "succeed")
		}
		return ctx.JSON(200, "failed")
	})

	//用户登录验证

	e.POST("/api/login", func(ctx echo.Context) error {
		email := ctx.FormValue("email")
		pw := ctx.FormValue("pw")
		pw = md5_(pw)
		rememberme := ctx.FormValue("checked")
		if !GetLogin(email, pw) {
			return ctx.JSON(200, "failed")
		}
		cookie := new (http.Cookie)
		cookie.Name = "GODKEY"
		cookie.Value,_ = DesEncrypt(email,DesKey)
		cookie.Path = "/"
		if rememberme == "true" {
			cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
		}
		ctx.SetCookie(cookie)
		return ctx.JSON(200, "succeed")
	})

	//用户容量查询
	e.POST("/api/cookies", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		userinfo_ := GetUserInfo(email)
		if userinfo_ != nil {
			return ctx.JSON(200,userinfo_)
		}
		return err
	})

	//新增储存策略
	e.POST("/api/store", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		if !IsAdmin(email) {
			return err
		}
		tmp := new(StoreModel)
		_ = ctx.Bind(tmp)
		if tmp.Type == "onedrive" {
			if !FirstGetRefreshToken(tmp) {
				i :=0
				for {
					if FirstGetRefreshToken(tmp) {
						break
					}
					if i>4 {
						return ctx.JSON(200,"failed")
					}
					i++
				}
			}
		}
		return ctx.JSON(200,"succeed")
	})

	//查询储存策略
	e.GET("/api/store", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		if !IsAdmin(email) {
			return ctx.JSON(400,"failed")
		}
		tmp := StoresQuery(nil)
		return ctx.JSON(200,tmp)
	})

	//查询用户
	e.GET("/api/user", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		if !IsAdmin(email) {
			return ctx.JSON(400,"failed")
		}
		tmp := UsersQuery(nil)
		return ctx.JSON(200,tmp)
	})

	//查询用户组
	e.GET("/api/usergroup", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		if !IsAdmin(email) {
			return ctx.JSON(400,"failed")
		}
		tmp := UserGroupsQuery(nil)
		return ctx.JSON(200,tmp)
	})

	//增加用户
	e.POST("/api/user", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		if !IsAdmin(email) {
			return ctx.JSON(400,"failed")
		}
		tmp := new(useradd)
		_ = ctx.Bind(tmp)
		if UserAdd(&User{
			Email:    tmp.Email,
			Password: md5_(tmp.Password),
			GroupID:  tmp.GroupID,
			Volume:   tmp.Volume,
		}) {
			return ctx.JSON(200,"succeed")
		}
		return  ctx.JSON(200,"failed")
	})

	//编辑用户
	e.PUT("/api/user", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		if !IsAdmin(email) {
			return ctx.JSON(200,"failed")
		}
		var tmp User
		tmp.Email = ctx.FormValue("email")
		groupid := ctx.FormValue("groupid")
		volume := ctx.FormValue("volume")
		userid := ctx.FormValue("userid")
		pw := ctx.FormValue("pw")
		if tmp.Email == "" {
			return ctx.JSON(200,"failed")
		}
		if  userid == "" {
			return ctx.JSON(200,"failed")
		}
		userid_,_ := strconv.Atoi(userid)
		tmp.UserID = userid_
		if groupid != "" {
			id,_ := strconv.Atoi(groupid)
			tmp.GroupID = id
		}
		if volume != "" {
			vol,_ := strconv.ParseInt(volume,10,64)
			tmp.Volume = vol
		}
		if pw != "" {
			tmp.Password = md5_(pw)
		}
		if UserUpdate(&tmp) {
			return ctx.JSON(200,"succeed")
		}
		return ctx.JSON(200,"failed")
	})

	//删除用户
	e.DELETE("/api/user", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		if !IsAdmin(email) {
			return ctx.JSON(400,"failed")
		}
		email_ := ctx.FormValue("email")
		if UserDelete(&User{
			Email: email_,
		}) {
			return ctx.JSON(200,"succeed")
		}
		return  ctx.JSON(400,"failed")
	})

	//增加用户组
	e.POST("/api/usergroup", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		if !IsAdmin(email) {
			return ctx.JSON(400,"failed")
		}
		tmp := new(usergroupadd)
		_ = ctx.Bind(tmp)
		if  UserGroupAdd(&UserGroup{
			Name:    tmp.Name,
			Volume:  tmp.Volume,
			StoreID: tmp.StoreID,
		}) {
			return ctx.JSON(200,"succeed")
		}
		return  ctx.JSON(200,"failed")
	})

	//编辑用户组
	e.PUT("/api/usergroup", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		if !IsAdmin(email) {
			return ctx.JSON(200,"failed")
		}
		var tmp UserGroup
		groupid := ctx.FormValue("groupid")
		storeid := ctx.FormValue("storeid")
		volume := ctx.FormValue("volume")
		tmp.Name = ctx.FormValue("name")
		if groupid == "" {
			return ctx.JSON(200,"failed")
		}
		tmp.GroupID,_ = strconv.Atoi(groupid)
		if volume != "" {
			vol,_ := strconv.ParseInt(volume,10,64)
			tmp.Volume = vol
		}
		if storeid != "" {
			tmp.StoreID = storeid
		}
		if UserGroupUpdate(&tmp){
			return ctx.JSON(200,"succeed")
		}
		return ctx.JSON(200,"failed")
	})

	//删除用户组
	e.DELETE("/api/usergroup", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		if !IsAdmin(email) {
			return ctx.JSON(400,"failed")
		}
		groupid, _ := strconv.Atoi(ctx.FormValue("groupid"))
		if UserGroupDelete(&UserGroup{
			GroupID: groupid,
		}) {
			return ctx.JSON(200,"succeed")
		}
		return  ctx.JSON(200,"failed")
	})

	//aria2配置设置
	e.POST("/api/aria2config", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		if !IsAdmin(email) {
			return ctx.JSON(400,"failed")
		}
		tmp := new(aria2config)
		_ = ctx.Bind(tmp)
		ModifyIni("aria2","enable",tmp.Status)
		ModifyIni("aria2","port",tmp.Port)
		ModifyIni("aria2","token",tmp.Secret)
		ModifyIni("TmpFile","path",tmp.TmpDownPath)
		ModifyIni("aria2","time",tmp.Time)
		if aria2client != nil {
			_ = aria2client.Close()
		}
		aria2client = aria2begin()
		return ctx.JSON(200,"succeed")
	})

	//读取aria2配置
	e.GET("/api/aria2config", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		if !IsAdmin(email) {
			return ctx.JSON(400,"failed")
		}
		tmp := aria2config{
			Status:      ReadIni("aria2","enable"),
			Port:        ReadIni("aria2","port"),
			Secret:      ReadIni("aria2","token"),
			TmpDownPath: ReadIni("TmpFile","path"),
			Time:        ReadIni("aria2","time"),
		}
		return ctx.JSON(200,tmp)
	})

	//aria2状态
	e.GET("/api/aria2test", func(ctx echo.Context) error {
		if  aria2client == nil {
			return ctx.JSON(200,"failed")
		}
		version,err := aria2client.GetVersion()
		if err != nil {
			return ctx.JSON(200,"failed")
		}
		type hhhh struct {
			Version string
			Time string
		}
		return ctx.JSON(200,hhhh{
			Version: version.Version,
			Time:    ReadIni("aria2","time"),
		})
	})

	//添加aria2下载
	e.POST("/api/aria2", func(ctx echo.Context) error {
		tmp := new(aria2accepturl)
		_ = ctx.Bind(tmp)
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		p := UserQuery(&User{
			Email: email,
		})
		a := aria2download(tmp.Url,tmp.Path,p.UserID)
		return ctx.JSON(200, len(a))
	})

	//获取aria2下载状态
	e.GET("/api/aria2", func(ctx echo.Context) error {
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		p := UserQuery(&User{
			Email: email,
		})
		infos := aria2status(p.UserID)

		return ctx.JSON(200,infos)
	})

	//修改aria2下载状态
	e.PUT("/api/aria2", func(ctx echo.Context) error {
		tmp := new(aria2change)
		_ = ctx.Bind(tmp)
		err := aria2taskchange(tmp)
		if err != nil {
			return ctx.JSON(200,"failed")
		}
		return ctx.JSON(200,"succeed")
	})

	//文件解压缩
	e.PUT("/api/UnArchive", func(ctx echo.Context) error {
		tmp := new(UnArchiveFile)
		_ = ctx.Bind(tmp)
		enkey,err := ctx.Cookie("GODKEY")
		if err != nil {
			return err
		}
		email,_ := DesDecrypt(enkey.Value,DesKey)
		go UnArchiver(tmp,email)
		return ctx.JSON(200, "succeed")
	})

}


