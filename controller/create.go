package controller

import (
	"github.com/cdfmlr/crud/log"
	"github.com/cdfmlr/crud/orm"
	"github.com/cdfmlr/crud/service"
	"github.com/gin-gonic/gin"
	"reflect"
)

// CreateHandler handles
//    POST /T
func CreateHandler[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var model T
		if err := c.ShouldBindJSON(&model); err != nil {
			ResponseError(c, CodeBadRequest, ErrBindFailed)
			return
		}
		log.Logger.Debugf("CreateHandler: Create %#v", model)
		err := service.Create(&model, service.IfNotExist())
		if err != nil {
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		c.JSON(200, SuccessResponseBody(model))
	}
}

// CreateNestedHandler handles
//    POST /P/:parentIDRouteParam/T
// where:
//  - P is the parent model, T is the child model
//  - parentIDRouteParam is the route param name of the parent model P
//  - field is the field name of the child model T in the parent model P
// responds with the updated parent model P
func CreateNestedHandler[P orm.Model, T orm.Model](parentIDRouteParam string, field string) gin.HandlerFunc {
	return func(c *gin.Context) {
		parentID := c.Param(parentIDRouteParam)
		if parentID == "" {
			ResponseError(c, CodeBadRequest, ErrMissingParentID)
			return
		}
		var child T
		if err := c.ShouldBindJSON(&child); err != nil {
			ResponseError(c, CodeBadRequest, ErrBindFailed)
			return
		}
		// child id exists: add to join table, but do not update child's fields
		if _, childID := child.Identity(); !reflect.ValueOf(childID).IsZero() {
			if err := service.GetByID[T](childID, &child); err != nil {
				ResponseError(c, CodeNotFound, err)
				return
			}
		}
		// else: id is not set: create new child

		var parent P
		if err := service.GetByID[P](parentID, &parent); err != nil {
			ResponseError(c, CodeNotFound, err)
			return
		}

		log.Logger.Debugf("CreateNestedHandler: Create %#v, parent=%#v", child, parent)

		//field := strings.ToUpper(field)[:1] + field[1:]
		field := NameToField(field, parent)
		err := service.Create(&child, service.NestInto(&parent, field))

		if err != nil {
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		ResponseSuccess(c, parent)
	}
}
