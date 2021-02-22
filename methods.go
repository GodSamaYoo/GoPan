package main

import (
	"time"
)

func GetPathData(email,path string) (data_ []PathData) {
	tmp_ := UserQuery(&User{
		Email: email,
	})
	data := DatasQuery(&Data{UserID: tmp_.UserID,Path: path})
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

func CreateDir(email,path,name string) bool {
	tmp_ := UserQuery(&User{
		Email: email,
	})
	tmp := Data{
		FileID:  md5_(time.Now().String()+name),
		UserID:  tmp_.UserID,
		Name:    name,
		Type:    "dir",
		Path:    path,
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

func AddressConvert(email,path,OneDrivePath string) string {
	if path == "/" {
		return OneDrivePath+"/"+email
	}
	return OneDrivePath+"/"+email+path
}