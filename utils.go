package gorm_admin

import (
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
)

func renderError(c *gin.Context, err error) {
	c.JSON(200, map[string]interface{}{"code": 400, "msg": err.Error()})
}

func renderOk(c *gin.Context, data interface{}) {
	if data == nil {
		data = "ok"
	}
	c.JSON(200, map[string]interface{}{"code": 200, "msg": "", "data": data})
}

func String2Type(src string, ty DatabaseType) interface{} {
	switch ty {
	case Int:
		return com.StrTo(src).MustInt64()
	case Float:
		return com.StrTo(src).MustFloat64()
	default:
		return src
	}
}
