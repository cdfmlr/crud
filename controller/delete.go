package controller

import (
	"github.com/cdfmlr/crud/log"
	"github.com/cdfmlr/crud/orm"
	"github.com/cdfmlr/crud/service"
	"github.com/gin-gonic/gin"
)

// DeleteHandler handles
//    DELETE /T/:idParam
func DeleteHandler[T orm.Model](idParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param(idParam)
		if id == "" {
			ResponseError(c, CodeBadRequest, ErrMissingID)
			return
		}
		log.Logger.Debugf("DeleteHandler: Delete %T, id=%v", *new(T), id)

		_, err := service.DeleteByID[T](id)
		if err != nil {
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		ResponseSuccess(c, nil, gin.H{"deleted": true})
	}
}

// DeleteNestedHandler handles
//    DELETE /P/:parentIdParam/T/:childIdParam
// where:
//  - P is the parent model, T is the child model
//  - parentIdParam is the route param name of the parent model P
//  - childIdParam is the route param name of the child model T in the parent model P
//  - field is the field name of the child model T in the parent model P
func DeleteNestedHandler[P orm.Model, T orm.Model](parentIdParam string, field string, childIdParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		parentId := c.Param(parentIdParam)
		if parentId == "" {
			ResponseError(c, CodeBadRequest, ErrMissingParentID)
			return
		}
		childId := c.Param(childIdParam)
		if childId == "" {
			ResponseError(c, CodeBadRequest, ErrMissingID)
			return
		}
		//field := strings.ToUpper(field)[:1] + field[1:]
		field := NameToField(field, new(P))
		log.Logger.Debugf("DeleteNestedHandler: Delete %v of %v, parentId=%v, field=%v, childId=%v", *new(T), *new(P), parentId, field, childId)
		err := service.DeleteNestedByID[P, T](parentId, field, childId)
		if err != nil {
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		ResponseSuccess(c, nil, gin.H{"deleted": true})
	}
}
