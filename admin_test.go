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

func TestAdmin(t *testing.T) {
	db, err := gorm.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/fg?charset=utf8&parseTime=True&loc=Local")
	if !assert.Nil(t, err) {
		return
	}
	db.LogMode(true)
	g := gin.Default()

	admin := InitAdmin(db, g, "/admin")
	admin.Table("users", "/users")
	admin.Start()
	g.Run(":8080")
}
