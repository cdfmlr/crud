package controller

import (
	"github.com/cdfmlr/crud/orm"
	"github.com/cdfmlr/crud/service"
	"github.com/gin-gonic/gin"
)

// DeleteHandler handles
//    DELETE /T/:idParam
// Deletes the model T with the given id.
//
// Request body: none
//
// Response:
//  - 200 OK: { deleted: true }
//  - 400 Bad Request: { error: "missing id" }
//  - 422 Unprocessable Entity: { error: "delete process failed" }
func DeleteHandler[T orm.Model](idParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param(idParam)
		if id == "" {
			logger.WithContext(c).
				WithField("idParam", idParam).
				Warn("DeleteHandler: read id param failed")
			ResponseError(c, CodeBadRequest, ErrMissingID)
			return
		}
		logger.WithContext(c).
			Tracef("DeleteHandler: Delete %T, id=%v", *new(T), id)

		_, err := service.DeleteByID[T](c, id)
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
//
// Request body: none
//
// Response:
//  - 200 OK: { deleted: true }
//  - 400 Bad Request: { error: "missing id" }
//  - 422 Unprocessable Entity: { error: "delete process failed" }
func DeleteNestedHandler[P orm.Model, T orm.Model](parentIdParam string, field string, childIdParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		parentId := c.Param(parentIdParam)
		if parentId == "" {
			logger.WithContext(c).
				WithField("parentIdParam", parentIdParam).
				Warn("DeleteNestedHandler: read id param failed")
			ResponseError(c, CodeBadRequest, ErrMissingParentID)
			return
		}
		childId := c.Param(childIdParam)
		if childId == "" {
			logger.WithContext(c).
				WithField("childIdParam", childIdParam).
				Warn("DeleteNestedHandler: read id param failed")
			ResponseError(c, CodeBadRequest, ErrMissingID)
			return
		}
		//field := strings.ToUpper(field)[:1] + field[1:]
		field := nameToField(field, new(P))

		logger.WithContext(c).
			Tracef("DeleteNestedHandler: Delete %v of %v, parentId=%v, field=%v, childId=%v", *new(T), *new(P), parentId, field, childId)

		err := service.DeleteNestedByID[P, T](c, parentId, field, childId)
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("DeleteNestedHandler: Delete failed")
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		ResponseSuccess(c, nil, gin.H{"deleted": true})
	}
}
