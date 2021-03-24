package main

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	path_ "path"
)

func CheckSqlite() {
	_, err := os.Stat("data.db")
	if err != nil {
		dbs, err_ := gorm.Open(sqlite.Open("data.db"), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		if err_ != nil {
			fmt.Println("创建Sqlite数据库失败")
		}
		db = *dbs

		_ = db.AutoMigrate(&User{})
		_ = db.AutoMigrate(&Data{})
		_ = db.AutoMigrate(&UserGroup{})
		_ = db.AutoMigrate(&Task{})
		_ = db.AutoMigrate(&Store{})
		db.Create(&User{
			UserID:   1,
			Email:    "admin@godcloud.com",
			Password: "49ba59abbe56e057",
			GroupID:  1,
			Volume:   1048576,
		})
		db.Create(&UserGroup{
			GroupID: 1,
			Name:    "管理员",
			Volume:  1048576,
			StoreID: "1",
		})
		fmt.Println("用户名：admin@godcloud.com\n密码：123456")
	} else {
		ConnectSqlite()
	}
}

func ConnectSqlite() {
	dbs, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic("连接Sqlite数据库失败")
	}
	db = *dbs
}

func UserQuery(tmp *User) *User {
	tmp_ := new(User)
	num := db.Where(tmp).Find(&tmp_).RowsAffected
	if num == 0 {
		return nil
	}
	return tmp_
}
func UsersQuery(tmp *User) []User {
	var tmp_ []User
	db.Where(tmp).Find(&tmp_)
	return tmp_
}
func UserAdd(tmp *User) bool {
	a := UserQuery(&User{
		Email: tmp.Email,
	})
	if a != nil {
		return false
	}
	if tmp.Volume == 0 {
		b := UserGroupQuery(&UserGroup{GroupID: tmp.GroupID})
		tmp.Volume = b.Volume
	}
	num := db.Create(&tmp).RowsAffected
	if num == 1 {
		return true
	}
	return false
}
func UserUpdate(tmp *User) bool {
	num := db.Model(&User{}).Where("user_id = ?", tmp.UserID).Updates(tmp).RowsAffected
	if num == 1 {
		return true
	}
	return false
}
func UserDelete(tmp *User) bool {
	num := db.Delete(tmp, "email = ?", tmp.Email).RowsAffected
	if num == 1 {
		db.Delete(&Data{}, "user_id = ?", tmp.UserID)
		return true
	}
	return false
}

func UserGroupQuery(tmp *UserGroup) *UserGroup {
	tmp_ := new(UserGroup)
	num := db.Where(tmp).Find(&tmp_).RowsAffected
	if num == 0 {
		return nil
	}
	return tmp_
}
func UserGroupsQuery(tmp *UserGroup) []UserGroup {
	var tmp_ []UserGroup
	db.Where(tmp).Find(&tmp_)
	return tmp_
}
func UserGroupAdd(tmp *UserGroup) bool {
	a := UserGroupQuery(&UserGroup{
		Name: tmp.Name,
	})
	if a != nil {
		return false
	}
	num := db.Create(&tmp).RowsAffected
	if num == 1 {
		return true
	}
	return false
}
func UserGroupUpdate(tmp *UserGroup) bool {
	num := db.Model(&UserGroup{}).Where("group_id = ?", tmp.GroupID).Updates(tmp).RowsAffected
	if num == 1 {
		return true
	}
	return false
}
func UserGroupDelete(tmp *UserGroup) bool {
	num := db.Delete(tmp, "group_id = ?", tmp.GroupID).RowsAffected
	if num == 1 {
		a := UsersQuery(&User{
			GroupID: tmp.GroupID,
		})
		for _, v := range a {
			UserDelete(&v)
		}
		return true
	}
	return false
}

func DataQuery(tmp *Data) *Data {
	tmp_ := new(Data)
	num := db.Where(tmp).Find(&tmp_).RowsAffected
	if num == 0 {
		return nil
	}
	return tmp_
}
func DatasQuery(tmp *Data) []Data {
	var tmp_ []Data
	db.Where(tmp).Find(&tmp_)
	return tmp_
}
func DataAdd(tmp *Data) bool {
	a := DataQuery(&Data{
		Path:   tmp.Path,
		Name:   tmp.Name,
		UserID: tmp.UserID,
	})
	if a != nil {
		return false
	}
	num := db.Create(&tmp).RowsAffected
	if num == 1 {
		return true
	}
	return false
}
func DataDelete(tmp *Data) bool {
	num := db.Delete(tmp, "file_id = ?", tmp.FileID).RowsAffected
	if num == 1 {
		return true
	}
	return false
}

func TaskQuery(tmp *Task) *Task {
	tmp_ := new(Task)
	num := db.Where(tmp).Find(&tmp_).RowsAffected
	if num == 0 {
		return nil
	}
	return tmp_
}
func TasksQuery(tmp *Task) []Task {
	var tmp_ []Task
	db.Where(tmp).Find(&tmp_)
	return tmp_
}
func TaskAdd(tmp *Task) bool {
	a := TaskQuery(&Task{
		Gid: tmp.Gid,
	})
	if a != nil {
		return false
	}
	num := db.Create(&tmp).RowsAffected
	if num == 1 {
		return true
	}
	return false
}
func TaskUpdate(tmp *Task) bool {
	num := db.Model(&Task{}).Where("tmp_path = ?", tmp.TmpPath).Updates(tmp).RowsAffected
	if num == 1 {
		return true
	}
	return false
}
func TaskDelete(tmp *Task) bool {
	num := db.Delete(tmp, "tmp_path = ?", tmp.TmpPath).RowsAffected
	if num == 1 {
		return true
	}
	return false
}

func StoreQuery(tmp *Store) *Store {
	tmp_ := new(Store)
	num := db.Where(tmp).Find(&tmp_).RowsAffected
	if num == 0 {
		return nil
	}
	return tmp_
}
func StoresQuery(tmp *Store) []Store {
	var tmp_ []Store
	db.Where(tmp).Find(&tmp_)
	return tmp_
}
func StoreAdd(tmp *Store) bool {
	a := StoreQuery(&Store{
		ClientID: tmp.ClientID,
	})
	if a != nil {
		return false
	}
	num := db.Create(&tmp).RowsAffected
	if num == 1 {
		return true
	}
	return false
}
func StoreUpdate(tmp *Store) bool {
	num := db.Model(&Store{}).Where("id = ?", tmp.ID).Updates(tmp).RowsAffected
	if num == 1 {
		return true
	}
	return false
}

//重命名文件 文件夹
func UpdateFile(tmp *UpdateType) bool {
	tmp_ := UserQuery(&User{
		Email: tmp.Email,
	})
	num := db.Where(&Data{UserID: tmp_.UserID, Name: tmp.NewName, Type: tmp.Type, Path: tmp.Path}).Find(&Data{}).RowsAffected
	if num != 0 {
		return false
	}
	if tmp.Type == "dir" {
		var oldpath, newpath string
		if tmp.Path == "/" {
			oldpath = tmp.Path + tmp.OldName + "%"
			newpath = tmp.Path + tmp.NewName
		} else {
			oldpath = tmp.Path + "/" + tmp.OldName + "%"
			newpath = tmp.Path + "/" + tmp.NewName
		}
		db.Model(&Data{}).Where("path LIKE ? AND user_id = ?", oldpath, tmp_.UserID).Update("path", newpath)
	} else {
		tmps := DataQuery(&Data{
			FileID: tmp.FileID,
		})
		if !RenameOneDriveFile(tmps.ItemID, tmps.StoreID, tmp.NewName) {
			return false
		}
	}
	db.Model(&Data{}).Where("file_id = ?", tmp.FileID).Updates(Data{
		Name: tmp.NewName,
	})
	return true
}

//验证是否为管理员
func IsAdmin(email string) bool {
	var user User
	row := db.Where(&User{Email: email, GroupID: 1}).Find(&user).RowsAffected
	if row == 1 {
		return true
	}
	return false
}

//删除文件夹
func DirDelete(fileid string) {
	a := DataQuery(&Data{
		FileID: fileid,
	})
	var tmp_ []Data
	var path string
	if a.Path == "/" {
		path = a.Path + a.Name + "%"
	} else {
		path = a.Path + "/" + a.Name + "%"
	}
	db.Where("path LIKE ? AND user_id = ?", path, a.UserID).Find(&tmp_)
	totalSize := int64(0)
	for _, v := range tmp_ {
		if DeleteOneDriveFile(v.ItemID, v.StoreID) {
			totalSize += v.Size
			picture := []string{"jpg", "jpeg", "bmp", "gif", "png", "tif"}
			tp := path_.Ext(v.Name)
			if tp != "" {
				for _, v_ := range picture {
					if tp[1:] == v_ {
						_ = os.Remove("Thumbnail/" + v.FileID + ".jpg")
						break
					}
				}
			}
			DataDelete(&Data{
				FileID: v.FileID,
			})
			x := StoreQuery(&Store{
				ID: v.StoreID,
			})
			StoreUpdate(&Store{
				ID:   x.ID,
				Used: x.Used - v.Size,
			})
		}
	}
	b := UserQuery(&User{UserID: a.UserID})
	UserUpdate(&User{UserID: a.UserID, Used: b.Used - totalSize})
	DataDelete(&Data{
		FileID: fileid,
	})
}
