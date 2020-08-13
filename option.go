package gorm_admin

import (
	"errors"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Option struct {
	table                                  interface{}
	tablePrtType                           reflect.Type
	key                                    string
	url                                    string
	canAdd, canEdit, canDel                bool
	selectFunc, addFunc, EditFunc, DelFunc gin.HandlerFunc
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

func (o *Option) GetSelectFunc(db *gorm.DB) gin.HandlerFunc {
	if o.selectFunc == nil {
		o.selectFunc = func(context *gin.Context) {
			req := new(BaseForm)
			if err := context.ShouldBind(req); err != nil {
				Error("SelectFunc=>param error:%s", err.Error())
				renderError(context, err)
				return
			}
			if req.Limit == 0 {
				req.Limit = 10
			}

			var total int64
			list := reflect.New(reflect.SliceOf(o.tablePrtType))
			Info("%+v %+v", list.Type(), list.Interface())
			err := db.Offset(req.Offset).Limit(req.Limit).Find(list.Interface()).Count(&total).Error
			if err != nil {
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
	if o.EditFunc == nil {
		o.EditFunc = func(context *gin.Context) {
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
	return o.EditFunc
}

func (o *Option) GetDelFunc(db *gorm.DB) gin.HandlerFunc {
	if o.DelFunc == nil {
		o.DelFunc = func(context *gin.Context) {
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
