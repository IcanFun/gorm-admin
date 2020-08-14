package gorm_admin

import (
	"strings"

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

func SnakeString(s string) string {
	s = strings.ReplaceAll(s, "ID", "Id")
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}
