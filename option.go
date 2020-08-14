package gorm_admin

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/unknwon/com"
)

type FilterType struct {
	FilterOperator
	DatabaseType
}

type table interface {
	TableName() string
}

type Option struct {
	table                                  table
	tablePrtType                           reflect.Type
	key                                    string
	url                                    string
	filter                                 map[string]FilterType
	canAdd, canEdit, canDel                bool
	selectFunc, addFunc, editFunc, delFunc gin.HandlerFunc
}

func (o *Option) CanAdd() *Option {
	o.canAdd = true
	return o
}

func (o *Option) CanEdit() *Option {
	o.canEdit = true
	return o
}

func (o *Option) CanDel() *Option {
	o.canDel = true
	return o
}

func (o *Option) SetFilter(filter map[string]FilterType) *Option {
	o.filter = filter
	return o
}

func (o *Option) SetSelectFunc(fun gin.HandlerFunc) *Option {
	o.selectFunc = fun
	return o
}

func (o *Option) SetAddFunc(fun gin.HandlerFunc) *Option {
	o.addFunc = fun
	return o
}

func (o *Option) SetEditFunc(fun gin.HandlerFunc) *Option {
	o.editFunc = fun
	return o
}

func (o *Option) SetDelFunc(fun gin.HandlerFunc) *Option {
	o.delFunc = fun
	return o
}

type Email struct {
	ID         int
	UserID     int    `gorm:"index"`                          // 外键 (属于), tag `index`是为该列创建索引
	Email      string `gorm:"type:varchar(100);unique_index"` // `type`设置sql类型, `unique_index` 为该列设置唯一索引
	Subscribed bool
}

func (o *Option) GetSelectFunc(db *gorm.DB) gin.HandlerFunc {
	if o.selectFunc == nil {
		o.selectFunc = func(context *gin.Context) {
			limit := com.StrTo(context.Query("limit")).MustInt()
			offset := com.StrTo(context.Query("offset")).MustInt()
			if limit == 0 {
				limit = 10
			}

			var total int64
			list := reflect.New(reflect.SliceOf(o.tablePrtType))
			session := db.Offset(offset).Limit(limit)

			for key, operator := range o.filter {
				data := context.Query(key)
				if data == "" {
					continue
				}
				switch operator.FilterOperator {
				case FilterOperatorLike:
					has := false
					if strings.HasPrefix(key, "%") {
						data = "%" + data
						has = true
					}
					if strings.HasSuffix(key, "%") {
						data = data + "%"
						has = true
					}
					if !has {
						data = "%" + data + "%"
					}
					session = session.Where(fmt.Sprintf("`%s`.`%s` like ?", o.table.TableName(), strings.ReplaceAll(key, "%", "")), data)
				case FilterOperatorIn:
					s := strings.Split(data, ",")
					d := make([]interface{}, len(s))
					for key, value := range s {
						d[key] = String2Type(value, operator.DatabaseType)
					}
					session = session.Where(fmt.Sprintf("`%s`.`%s` in (?)", o.table.TableName(), key), d)
				case FilterOperatorBetween:
					s := strings.Split(data, "-")
					if len(s) == 2 {
						session = session.Where(fmt.Sprintf("`%s`.`%s` between ? AND ?", o.table.TableName(), key), String2Type(s[0], operator.DatabaseType), String2Type(s[1], operator.DatabaseType))
					}
				default:
					session = session.Where(fmt.Sprintf("`%s`.`%s` %s ?", o.table.TableName(), key, operator.FilterOperator), String2Type(data, operator.DatabaseType))
				}
			}
			err := session.Find(list.Interface()).Count(&total).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				Error("SelectFunc=>Find error:%s", err.Error())
				renderError(context, err)
				return
			}

			renderOk(context, map[string]interface{}{"list": list.Interface(), "total": total})
		}
	}
	return o.selectFunc
}

func (o *Option) GetAddFunc(db *gorm.DB) gin.HandlerFunc {
	if o.addFunc == nil {
		o.addFunc = func(context *gin.Context) {
			req := reflect.New(o.tablePrtType.Elem())
			if err := context.ShouldBind(req.Interface()); err != nil {
				Error("AddFunc=>param error:%s", err.Error())
				renderError(context, err)
				return
			}
			err := db.Create(req.Interface()).Error
			if err != nil {
				Error("AddFunc=>Find error:%s", err.Error())
				renderError(context, err)
				return
			}
			renderOk(context, req.Interface())
		}
	}
	return o.addFunc
}

func (o *Option) GetEditFunc(db *gorm.DB) gin.HandlerFunc {
	if o.editFunc == nil {
		o.editFunc = func(context *gin.Context) {
			req := reflect.New(o.tablePrtType.Elem())
			if err := context.ShouldBind(req.Interface()); err != nil {
				Error("EditFunc=>param error:%s", err.Error())
				renderError(context, err)
				return
			}
			Info("%+v %+v", req, req.Elem().FieldByName(o.key).IsValid())
			if req.Elem().FieldByName(o.key).IsZero() {
				renderError(context, errors.New(o.key+"必填"))
				return
			}
			err := db.Save(req.Interface()).Error
			if err != nil {
				Error("EditFunc=>Find error:%s", err.Error())
				renderError(context, err)
				return
			}
			renderOk(context, req.Interface())
		}
	}
	return o.editFunc
}

func (o *Option) GetDelFunc(db *gorm.DB) gin.HandlerFunc {
	if o.delFunc == nil {
		o.delFunc = func(context *gin.Context) {
			req := reflect.New(o.tablePrtType.Elem())
			if err := context.ShouldBind(req.Interface()); err != nil {
				Error("DelFunc=>param error:%s", err.Error())
				renderError(context, err)
				return
			}
			Info("%+v %+v", req, req.Elem().FieldByName(o.key).IsValid())
			if req.Elem().FieldByName(o.key).IsZero() {
				renderError(context, errors.New(o.key+"必填"))
				return
			}
			err := db.Delete(req.Interface()).Error
			if err != nil {
				Error("DelFunc=>Find error:%s", err.Error())
				renderError(context, err)
				return
			}
			renderOk(context, nil)
		}
	}
	return o.delFunc
}
