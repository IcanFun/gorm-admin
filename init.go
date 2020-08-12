package gorm_admin

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Admin struct {
	db      *gorm.DB
	e       *gin.Engine
	prefix  string
	options []*Option
}

func InitAdmin(db *gorm.DB, e *gin.Engine, prefix string) *Admin {
	ConfigZapLog("")
	return &Admin{db: db, e: e, prefix: prefix, options: []*Option{}}
}

func (a *Admin) SetLogLevel(level string) {
	ConfigZapLog(level)
}

func (a *Admin) Table(table string, url string) *Option {
	a.db.AutoMigrate(table)
	option := &Option{table: table, url: url}
	a.options = append(a.options, option)
	return option
}

func (a *Admin) Start() {
	for _, option := range a.options {
		a.e.GET(a.prefix+option.url, option.GetSelectFunc(a.db))
		if option.canAdd {
			a.e.POST(a.prefix+option.url, option.GetAddFunc(a.db))
		}
		if option.canEdit {
			a.e.PUT(a.prefix+option.url, option.GetEditFunc(a.db))
		}
		if option.canDel {
			a.e.DELETE(a.prefix+option.url, option.GetDelFunc(a.db))
		}
	}
}
