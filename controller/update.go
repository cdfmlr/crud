package controller

import (
	"github.com/cdfmlr/crud/log"
	"github.com/cdfmlr/crud/orm"
	"github.com/cdfmlr/crud/service"
	"github.com/gin-gonic/gin"
)

// UpdateHandler handles
//    PUT /T/:idParam
func UpdateHandler[T orm.Model](idParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var model T

		id := c.Param(idParam) // NOTICE: id is a string
		if id == "" {
			ResponseError(c, CodeBadRequest, ErrMissingID)
			return
		}

		if err := service.GetByID[T](id, &model); err != nil {
			ResponseError(c, CodeNotFound, err)
			return
		}

		var updatedModel = model
		if err := c.ShouldBindJSON(&updatedModel); err != nil {
			ResponseError(c, CodeBadRequest, err)
			return
		}
		log.Logger.Debugf("UpdateHandler: Update %#v, id=%v", updatedModel, id)
		_, oldID := model.Identity()
		_, newID := updatedModel.Identity()
		if oldID != newID {
			ResponseError(c, CodeBadRequest, ErrUpdateID)
			return
		}

		_, err := service.Update(&updatedModel)
		if err != nil {
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		ResponseSuccess(c, &updatedModel)
	}
}
