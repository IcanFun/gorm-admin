package gorm_admin

import (
	"reflect"

	"github.com/gin-gonic/gin"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Admin struct {
	db      *gorm.DB
	e       *gin.RouterGroup
	options []*Option
}

func InitAdmin(db *gorm.DB, e *gin.RouterGroup) *Admin {
	ConfigZapLog("")
	return &Admin{db: db, e: e, options: []*Option{}}
}

func (a *Admin) SetDb(db *gorm.DB) {
	a.db = db
}

func (a *Admin) SetLogLevel(level string) {
	ConfigZapLog(level)
}

func (a *Admin) Table(table table, tableKey, url string) *Option {
	option := &Option{Table: table, Key: tableKey, Url: url}
	a.AddOption(option)

	return option
}

func (a *Admin) AddOption(option *Option) {
	if option == nil {
		return
	}
	if option.Table == nil || option.Key == "" {
		panic("参数必填")
	}
	itemPtrType := reflect.TypeOf(option.Table)
	if itemPtrType.Kind() != reflect.Ptr {
		itemPtrType = reflect.PtrTo(itemPtrType)
	}
	item := reflect.New(itemPtrType.Elem())
	if !item.Elem().FieldByName(option.Key).IsValid() {
		panic("tableKey无法解析")
	}
	option.tablePrtType = itemPtrType
	a.options = append(a.options, option)
}

func (a *Admin) Start() {
	for _, option := range a.options {
		var handlers []gin.HandlerFunc
		if len(option.globalMw) > 0 {
			handlers = append(handlers, option.globalMw...)
		}
		if len(option.sel.Mw) > 0 {
			handlers = append(handlers, option.sel.Mw...)
		}
		handlers = append(handlers, option.GetSelectFunc(a.db))
		a.e.GET(option.Url, handlers...)
		if option.add.Open {
			var handlers []gin.HandlerFunc
			if len(option.globalMw) > 0 {
				handlers = append(handlers, option.globalMw...)
			}
			if len(option.add.Mw) > 0 {
				handlers = append(handlers, option.add.Mw...)
			}
			handlers = append(handlers, option.GetAddFunc(a.db))
			a.e.POST(option.Url, handlers...)
		}
		if option.edit.Open {
			var handlers []gin.HandlerFunc
			if len(option.globalMw) > 0 {
				handlers = append(handlers, option.globalMw...)
			}
			if len(option.edit.Mw) > 0 {
				handlers = append(handlers, option.edit.Mw...)
			}
			handlers = append(handlers, option.GetEditFunc(a.db))
			a.e.PUT(option.Url, handlers...)
		}
		if option.del.Open {
			var handlers []gin.HandlerFunc
			if len(option.globalMw) > 0 {
				handlers = append(handlers, option.globalMw...)
			}
			if len(option.del.Mw) > 0 {
				handlers = append(handlers, option.del.Mw...)
			}
			handlers = append(handlers, option.GetDelFunc(a.db))
			a.e.DELETE(option.Url, handlers...)
		}
	}
}
