package gorm_admin

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
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
	if table == nil || tableKey == "" {
		panic("参数必填")
	}
	itemPtrType := reflect.TypeOf(table)
	if itemPtrType.Kind() != reflect.Ptr {
		itemPtrType = reflect.PtrTo(itemPtrType)
	}
	item := reflect.New(itemPtrType.Elem())
	if !item.Elem().FieldByName(tableKey).IsValid() {
		panic("tableKey无法解析")
	}
	option := &Option{table: table, tablePrtType: itemPtrType, key: tableKey, url: url}
	a.options = append(a.options, option)

	return option
}

func (a *Admin) Start() {
	for _, option := range a.options {
		a.e.GET(option.url, option.GetSelectFunc(a.db))
		if option.add.Open {
			a.e.POST(option.url, option.GetAddFunc(a.db))
		}
		if option.edit.Open {
			a.e.PUT(option.url, option.GetEditFunc(a.db))
		}
		if option.del.Open {
			a.e.DELETE(option.url, option.GetDelFunc(a.db))
		}
	}
}
