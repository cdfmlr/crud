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
			logger.WithContext(c).WithField("idParam", idParam).
				Warn("UpdateHandler: Missing id")
			ResponseError(c, CodeBadRequest, ErrMissingID)
			return
		}

		if err := service.GetByID[T](c, id, &model); err != nil {
			logger.WithContext(c).WithError(err).
				Warn("UpdateHandler: GetByID failed")
			ResponseError(c, CodeNotFound, err)
			return
		}

		var updatedModel = model
		if err := c.ShouldBindJSON(&updatedModel); err != nil {
			logger.WithContext(c).WithError(err).
				Warn("UpdateHandler: Bind failed")
			ResponseError(c, CodeBadRequest, err)
			return
		}

		log.Logger.Tracef("UpdateHandler: Update %#v, id=%v", updatedModel, id)

		_, oldID := model.Identity()
		_, newID := updatedModel.Identity()
		if oldID != newID {
			logger.WithContext(c).WithField("idParam", idParam).
				WithField("oldID", oldID).
				WithField("newID", newID).
				Warn("UpdateHandler: id mismatch: cannot update id")
			ResponseError(c, CodeBadRequest, ErrUpdateID)
			return
		}

		_, err := service.Update(c, &updatedModel)
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("UpdateHandler: Update failed")
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		ResponseSuccess(c, &updatedModel)
	}
}
