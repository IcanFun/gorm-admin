package gorm_admin

import (
	"reflect"

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

func (a *Admin) Table(table interface{}, tableKey, url string) *Option {
	if table == nil || tableKey == "" {
		panic("参数必填")
	}
	itemPtrType := reflect.TypeOf(table)
	if itemPtrType.Kind() != reflect.Ptr {
		itemPtrType = reflect.PtrTo(itemPtrType)
	}
	item := reflect.New(itemPtrType.Elem())
	if !item.Elem().FieldByName(tableKey).IsValid() {
		panic("tableKey填写错误")
	}
	option := &Option{table: table, tablePrtType: itemPtrType, key: tableKey, url: url}
	a.options = append(a.options, option)

	err := a.db.AutoMigrate(table).Error
	if err != nil {
		Error(err.Error())
	}
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
