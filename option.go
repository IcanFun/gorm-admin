package gorm_admin

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

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

type JoinCon struct {
	JoinTable  string
	TableAlias string //连表别名
	ON         string //连表条件
}

type CurdCon struct {
	Open    bool              //能否增删改查
	Select  string            //mysql select 的列
	Func    gin.HandlerFunc   //自定义方法
	Mw      []gin.HandlerFunc //增删改查中间件
	MwParam []string          //中间件保存的参数，用于增删改查的参数补充
	Join    []JoinCon
}

type Option struct {
	Table               table  //表格struct
	Key                 string //表主键
	Url                 string //注册路由
	tablePrtType        reflect.Type
	filter              map[string]FilterType //查询所用的筛选条件
	add, del, edit, sel CurdCon
	globalMwParam       []string //中间件保存的参数，用于增删改查的参数补充,全局通用的
}

func (c *CurdCon) SetFunc(fun gin.HandlerFunc) *CurdCon {
	c.Func = fun
	return c
}

func (c *CurdCon) SetMw(mws ...gin.HandlerFunc) *CurdCon {
	c.Mw = mws
	return c
}

func (c *CurdCon) SetMwParam(params ...string) *CurdCon {
	c.MwParam = params
	return c
}

func (c *CurdCon) SetJoin(joins ...JoinCon) *CurdCon {
	c.Join = joins
	return c
}

func (o *Option) SetGlobalMwParam(keys []string) *Option {
	o.globalMwParam = keys
	return o
}

func (o *Option) SetSelect(con CurdCon) *Option {
	o.sel = con
	return o
}

func (o *Option) SetAdd(con CurdCon) *Option {
	o.add = con
	return o
}

func (o *Option) SetEdit(con CurdCon) *Option {
	o.edit = con
	return o
}

func (o *Option) SetDel(con CurdCon) *Option {
	o.del = con
	return o
}

func (o *Option) SetFilter(filter map[string]FilterType) *Option {
	o.filter = filter
	return o
}

func (o *Option) GetSelectFunc(db *gorm.DB) gin.HandlerFunc {
	if o.sel.Func == nil {
		o.sel.MwParam = append(o.sel.MwParam, o.globalMwParam...)
		if o.sel.Select == "" {
			o.sel.Select = "*"
		}
		o.sel.Func = func(context *gin.Context) {
			for _, value := range o.sel.Mw {
				value(context)
				if context.IsAborted() {
					return
				}
			}

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
					session = session.Where(fmt.Sprintf("%s like ?", strings.ReplaceAll(key, "%", "")), data)
				case FilterOperatorIn:
					s := strings.Split(data, ",")
					d := make([]interface{}, len(s))
					for key, value := range s {
						d[key] = String2Type(value, operator.DatabaseType)
					}
					session = session.Where(fmt.Sprintf("%s in (?)", key), d)
				case FilterOperatorBetween:
					s := strings.Split(data, "-")
					if len(s) == 2 {
						session = session.Where(fmt.Sprintf("%s between ? AND ?", key), String2Type(s[0], operator.DatabaseType), String2Type(s[1], operator.DatabaseType))
					}
				default:
					session = session.Where(fmt.Sprintf("%s %s ?", key, operator.FilterOperator), String2Type(data, operator.DatabaseType))
				}
			}
			for _, value := range o.sel.MwParam {
				if data, ok := context.Get(value); ok {
					session = session.Where(fmt.Sprintf("%s = ?", value), data)
				}
			}

			for _, value := range o.sel.Join {
				session = session.Joins(fmt.Sprintf("LEFT JOIN `%s` %s ON %s", value.JoinTable, value.TableAlias, value.ON))
			}

			Debug("%+v %+v %+v", list, list.Elem(), list.Interface())

			err := session.Select(o.sel.Select).Find(list.Interface()).Count(&total).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				Error("SelectFunc=>Find error:%s", err.Error())
				renderError(context, err)
				return
			}

			renderOk(context, map[string]interface{}{"list": list.Interface(), "total": total})
		}
	}
	return o.sel.Func
}

func (o *Option) GetAddFunc(db *gorm.DB) gin.HandlerFunc {
	if o.add.Func == nil {
		o.add.MwParam = append(o.add.MwParam, o.globalMwParam...)

		o.add.Func = func(context *gin.Context) {
			for _, value := range o.add.Mw {
				value(context)
				if context.IsAborted() {
					return
				}
			}

			req := reflect.New(o.tablePrtType.Elem())
			if err := context.ShouldBind(req.Interface()); err != nil {
				Error("AddFunc=>param error:%s", err.Error())
				renderError(context, err)
				return
			}
			for _, value := range o.add.MwParam {
				if data, ok := context.Get(value); ok {
					Info("%+v %+v", req.Elem(), req.Elem().FieldByName(value))
					if req.Elem().FieldByName(value).IsValid() && req.Elem().FieldByName(value).IsZero() {
						req.Elem().FieldByName(value).Set(reflect.ValueOf(data))
					}
				}
			}
			Debug("%+v %+v %+v", req, req.Elem(), req.Interface())
			err := db.Create(req.Interface()).Error
			if err != nil {
				Error("AddFunc=>Find error:%s", err.Error())
				renderError(context, err)
				return
			}
			renderOk(context, req.Interface())
		}
	}
	return o.add.Func
}

func (o *Option) GetEditFunc(db *gorm.DB) gin.HandlerFunc {
	if o.edit.Func == nil {
		o.edit.MwParam = append(o.edit.MwParam, o.globalMwParam...)

		o.edit.Func = func(context *gin.Context) {
			for _, value := range o.edit.Mw {
				value(context)
				if context.IsAborted() {
					return
				}
			}

			req := map[string]interface{}{}
			if err := context.ShouldBind(&req); err != nil {
				Error("EditFunc=>param error:%s", err.Error())
				renderError(context, err)
				return
			}

			key := SnakeString(o.Key)
			if _, ok := req[key]; !ok {
				renderError(context, errors.New(o.Key+"必填"))
				return
			}
			for _, value := range o.edit.MwParam {
				if data, ok := context.Get(value); ok {
					value = SnakeString(value)
					if _, ok := req[value]; !ok {
						req[value] = data
					}
				}
			}

			if _, ok := req["update_at"]; !ok {
				if _, ok := o.tablePrtType.Elem().FieldByName("UpdateAt"); ok {
					req["update_at"] = time.Now().Unix()
				}
			}

			Debug("%+v", req)
			err := db.Table(o.Table.TableName()).Where(fmt.Sprintf("`%s`.`%s` = ?", o.Table.TableName(), key), req[key]).Updates(req).Error
			if err != nil {
				Error("EditFunc=>Find error:%s", err.Error())
				renderError(context, err)
				return
			}
			renderOk(context, req)
		}
	}
	return o.edit.Func
}

func (o *Option) GetDelFunc(db *gorm.DB) gin.HandlerFunc {
	if o.del.Func == nil {
		o.del.MwParam = append(o.del.MwParam, o.globalMwParam...)

		o.del.Func = func(context *gin.Context) {
			for _, value := range o.del.Mw {
				value(context)
				if context.IsAborted() {
					return
				}
			}

			req := reflect.New(o.tablePrtType.Elem())
			if err := context.ShouldBind(req.Interface()); err != nil {
				Error("DelFunc=>param error:%s", err.Error())
				renderError(context, err)
				return
			}

			if req.Elem().FieldByName(o.Key).IsZero() {
				renderError(context, errors.New(o.Key+"必填"))
				return
			}
			for _, value := range o.del.MwParam {
				if data, ok := context.Get(value); ok {
					if req.Elem().FieldByName(value).IsValid() && req.Elem().FieldByName(value).IsZero() {
						req.Elem().FieldByName(value).Set(reflect.ValueOf(data))
					}
				}
			}
			Debug("%+v %+v %+v", req, req.Elem(), req.Interface())
			err := db.Delete(req.Interface()).Error
			if err != nil {
				Error("DelFunc=>Find error:%s", err.Error())
				renderError(context, err)
				return
			}
			renderOk(context, nil)
		}
	}
	return o.del.Func
}
