package gorm_admin

import (
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

type ChangeModel struct {
	ID       int64  `json:"id"`
	CreateAt int64  `json:"create_at"`
	UpdateAt int64  `json:"update_at"`
	DeleteAt *int64 `json:"delete_at,omitempty"`
}

type User struct {
	ChangeModel
	Username string `json:"username"`
	Update   bool   `json:"update"`
	Email    Email
}

func (User) TableName() string {
	return "users"
}

func TestAdmin(t *testing.T) {
	db, err := gorm.Open("mysql", "root:123456@tcp(192.168.31.201:3306)/fg?charset=utf8&parseTime=True&loc=Local")
	if !assert.Nil(t, err) {
		return
	}
	db.LogMode(true)
	db.Callback().Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	db.Callback().Update().Replace("gorm:update_time_stamp", updateTimeStampForUpdateCallback)
	db.Callback().Delete().Replace("gorm:delete", deleteCallback)
	db.AutoMigrate(Email{})

	g := gin.Default()

	admin := InitAdmin(db, g.Group("/admin"))
	option := admin.Table(&User{}, "ID", "/users").CanAdd().CanEdit().CanDel()
	option.SetFilter(map[string]FilterType{
		"id": {
			FilterOperator: FilterOperatorEqual,
			DatabaseType:   Int,
		},
		"username": {
			FilterOperator: FilterOperatorLike,
			DatabaseType:   String,
		},
		"update": {
			FilterOperator: FilterOperatorIn,
			DatabaseType:   Float,
		},
		"update_at": {
			FilterOperator: FilterOperatorBetween,
			DatabaseType:   Int,
		},
	})
	admin.Start()
	g.Run(":8080")
}

// // 注册新建钩子在持久化之前
func updateTimeStampForCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		nowTime := time.Now().Unix()
		if createTimeField, ok := scope.FieldByName("create_at"); ok {
			if createTimeField.IsBlank {
				createTimeField.Set(nowTime)
			}
		}

		if modifyTimeField, ok := scope.FieldByName("update_at"); ok {
			if modifyTimeField.IsBlank {
				modifyTimeField.Set(nowTime)
			}
		}
	}
}

// 注册更新钩子在持久化之前
func updateTimeStampForUpdateCallback(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:update_column"); !ok {
		scope.SetColumn("update_at", time.Now().Unix())
	}
}

// 注册删除钩子在删除之前
func deleteCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		var extraOption string
		if str, ok := scope.Get("gorm:delete_option"); ok {
			extraOption = fmt.Sprint(str)
		}

		deletedOnField, hasDeletedOnField := scope.FieldByName("delete_at")

		if !scope.Search.Unscoped && hasDeletedOnField {
			scope.Raw(fmt.Sprintf(
				"UPDATE %v SET %v=%v%v%v",
				scope.QuotedTableName(),
				scope.Quote(deletedOnField.DBName),
				scope.AddToVars(time.Now().Unix()),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		} else {
			scope.Raw(fmt.Sprintf(
				"DELETE FROM %v%v%v",
				scope.QuotedTableName(),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		}
	}
}

func addExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}
