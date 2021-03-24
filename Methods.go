package main

import (
	"os"
	"path/filepath"
	"time"
)

func GetPathData(email, path string) (data_ []PathData) {
	tmp_ := UserQuery(&User{
		Email: email,
	})
	data := DatasQuery(&Data{UserID: tmp_.UserID, Path: path})
	for _, v := range data {
		tmp := PathData{
			DataFileId: v.FileID,
			DataName:   v.Name,
			DataType:   v.Type,
			DataTime:   v.Time,
			DataPath:   v.Path,
			DataSize:   v.Size,
		}
		data_ = append(data_, tmp)
	}
	return data_
}

func CreateDir(email, path, name string) bool {
	tmp_ := UserQuery(&User{
		Email: email,
	})
	tmp := Data{
		FileID: md5_(time.Now().String() + name),
		UserID: tmp_.UserID,
		Name:   name,
		Type:   "dir",
		Path:   path,
	}
	return DataAdd(&tmp)
}

func RenameData(tmp *UpdateType) bool {
	if !UpdateFile(tmp) {
		return false
	}
	return true
}

func GetLogin(email string, pw string) bool {
	row := UserQuery(&User{
		Email:    email,
		Password: pw,
	})
	if row != nil {
		return true
	}
	return false
}

func GetUserInfo(email string) *UserInfo {
	tmp := UserQuery(&User{
		Email: email,
	})
	var a UserInfo
	if tmp != nil {
		a.Volume = tmp.Volume
		a.GroupID = tmp.GroupID
		a.RegisterTime = tmp.Time
		a.Used = tmp.Used
		tmp_ := UserGroupQuery(&UserGroup{
			GroupID: tmp.GroupID,
		})
		a.GroupName = tmp_.Name
		return &a
	}
	return nil
}

//判断本地容量是否足够
func IsLocalVolume(Need int64) bool {
	used, _ := DirSize(TmpPath)
	free := TmpVolume - used
	if free < Need {
		return false
	}
	return true
}

//判断用户容量是否足够
func IsUserVolume(email string, Need int64) bool {
	a := UserQuery(&User{
		Email: email,
	})
	free := a.Volume - a.Used
	if free < Need {
		return false
	}
	return true
}

//判断储存策略容量是否足够

func IsStoreVolume(id int, Need int64) bool {
	a := StoreQuery(&Store{
		ID: id,
	})
	free := a.Volume - a.Used
	if free < Need {
		return false
	}
	return true
}

//获取目录已用大小  单位：KB
func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size/1024 + 1, err
}
