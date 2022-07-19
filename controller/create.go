package controller

import (
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
			logger.WithContext(c).WithError(err).
				Warn("CreateHandler: Bind failed")
			ResponseError(c, CodeBadRequest, err)
			return
		}
		logger.WithContext(c).Tracef("CreateHandler: Create %#v", model)
		err := service.Create(c, &model, service.IfNotExist())
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("CreateHandler: Create failed")
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
			logger.WithContext(c).WithError(err).
				Warn("CreateNestedHandler: Bind failed")
			ResponseError(c, CodeBadRequest, err)
			return
		}

		if _, childID := child.Identity(); !reflect.ValueOf(childID).IsZero() {
			// child id exists: add to join table, but do not update child's fields
			logger.WithField("childID", childID).Debug("CreateNestedHandler: child model has ID, add to join table, but do not update child's fields")
			if err := service.GetByID[T](c, childID, &child); err != nil {
				logger.WithContext(c).WithError(err).
					WithField("note", "try to query it because child id exists in request").
					Warn("CreateNestedHandler: GetByID[Child] failed")
				ResponseError(c, CodeNotFound, err)
				return
			}
		}
		// else: id is not set: create new child

		var parent P
		if err := service.GetByID[P](c, parentID, &parent); err != nil {
			logger.WithContext(c).WithError(err).
				Warn("CreateNestedHandler: GetByID[Parent] failed")
			ResponseError(c, CodeNotFound, err)
			return
		}

		logger.WithContext(c).
			Tracef("CreateNestedHandler: Create %#v, parent=%#v", child, parent)

		//field := strings.ToUpper(field)[:1] + field[1:]
		field := NameToField(field, parent)

		err := service.Create(c, &child, service.NestInto(&parent, field))
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("CreateNestedHandler: CreateNest failed")
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		ResponseSuccess(c, parent)
	}
}
