package gorm_admin

import (
	"database/sql"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type table interface {
	TableName() string
}

type Option struct {
	table                                  table
	url                                    string
	canAdd, canEdit, canDel                bool
	selectFunc, addFunc, EditFunc, DelFunc gin.HandlerFunc
}

func (o *Option) CanAdd() *Option {
	o.canAdd = true
	return o
}

func (o *Option) CanEdit() *Option {
	o.canAdd = true
	return o
}

func (o *Option) CanDel() *Option {
	o.canAdd = true
	return o
}

func (o *Option) GetSelectFunc(db *gorm.DB) gin.HandlerFunc {
	if o.selectFunc == nil {
		o.selectFunc = func(context *gin.Context) {
			req := new(BaseForm)
			if err := context.ShouldBind(req); err != nil {
				Error("GetSelectFunc=>param error:%s", err.Error())
				renderError(context, err)
				return
			}
			if req.Limit == 0 {
				req.Limit = 10
			}

			var total int64
			rows, err := db.Table(o.table.TableName()).Offset(req.Offset).Limit(req.Limit).Count(&total).Rows()
			if err != nil {
				Error("GetSelectFunc=>Find error:%s", err.Error())
				renderError(context, err)
				return
			}

			list := rows2maps(rows)
			renderOk(context, map[string]interface{}{"list": list, "total": total})
		}
	}
	return o.selectFunc
}

func (o *Option) GetAddFunc(db *gorm.DB) gin.HandlerFunc {
	if o.addFunc == nil {
		o.addFunc = func(context *gin.Context) {
			req := reflect.New(reflect.TypeOf(o.table)).Pointer()
			if err := context.ShouldBind(req); err != nil {
				Error("GetSelectFunc=>param error:%s", err.Error())
				renderError(context, err)
				return
			}
			err := db.Table(o.table.TableName()).Create(req).Error
			if err != nil {
				Error("GetSelectFunc=>Find error:%s", err.Error())
				renderError(context, err)
				return
			}
			renderOk(context, o.table)
		}
	}
	return o.addFunc
}

func (o *Option) GetEditFunc(db *gorm.DB) gin.HandlerFunc {
	if o.EditFunc == nil {
		o.EditFunc = func(context *gin.Context) {

		}
	}
	return o.EditFunc
}

func (o *Option) GetDelFunc(db *gorm.DB) gin.HandlerFunc {
	if o.DelFunc == nil {
		o.DelFunc = func(context *gin.Context) {

		}
	}
	return o.DelFunc
}

func renderError(c *gin.Context, err error) {
	c.JSON(200, map[string]interface{}{"code": 400, "msg": err.Error()})
}

func renderOk(c *gin.Context, data interface{}) {
	if data == nil {
		data = "ok"
	}
	c.JSON(200, map[string]interface{}{"code": 200, "msg": "", "data": data})
}

// *sql.Rows 转换为 []map[string]interface{}类型
func rows2maps(rows *sql.Rows) (res []map[string]interface{}) {
	defer rows.Close()
	cols, _ := rows.Columns()
	cache := make([]interface{}, len(cols))
	// 为每一列初始化一个指针
	for index, _ := range cache {
		var a interface{}
		cache[index] = &a
	}

	for rows.Next() {
		rows.Scan(cache...)
		row := make(map[string]interface{})
		for i, val := range cache {
			// 处理数据类型
			v := *val.(*interface{})
			switch v.(type) {
			case []uint8:
				v = string(v.([]uint8))
			case nil:
				v = ""
			}
			row[cols[i]] = v
		}

		res = append(res, row)
	}

	return res
}
