package gorm_admin

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

type User struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
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
	g := gin.Default()

	admin := InitAdmin(db, g, "/admin")
	admin.Table(&User{}, "/users").CanAdd()
	admin.Start()
	g.Run(":8080")
}
