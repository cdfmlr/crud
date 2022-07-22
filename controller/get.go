package controller

import (
	"context"
	"github.com/cdfmlr/crud/orm"
	"github.com/cdfmlr/crud/service"
	"github.com/gin-gonic/gin"
	"reflect"
)

// GetRequestOptions is the query options (?opt=val) for GET requests:
//
//     limit=10&offset=4&                 # pagination
//     order_by=id&desc=true&             # ordering
//     filter_by=name&filter_value=John&  # filtering
//     total=true&                        # return total count (all available records under the filter, ignoring pagination)
//     preload=Product&preload=Product.Manufacturer  # preloading: loads nested models as well
//
// It is used in GetListHandler, GetByIDHandler and GetFieldHandler, to bind
// the query parameters in the GET request url.
type GetRequestOptions struct {
	Limit       int      `form:"limit"`
	Offset      int      `form:"offset"`
	OrderBy     string   `form:"order_by"`
	Descending  bool     `form:"desc"`
	FilterBy    string   `form:"filter_by"`
	FilterValue string   `form:"filter_value"`
	Preload     []string `form:"preload"` // fields to preload
	Total       bool     `form:"total"`   // return total count ?
}

// GetListHandler handles
//    GET /T
// It returns a list of models.
//
// QueryOptions (See GetRequestOptions for more details):
//    limit, offset, order_by, desc, filter_by, filter_value, preload, total.
//
// Response:
//  - 200 OK: { Ts: [{...}, ...] }
//  - 400 Bad Request: { error: "request band failed" }
//  - 422 Unprocessable Entity: { error: "get process failed" }
func GetListHandler[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request GetRequestOptions
		if err := c.ShouldBind(&request); err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetListHandler: bind request failed")
			ResponseError(c, CodeBadRequest, err)
			return
		}

		options := buildQueryOptions(request)

		var dest []*T
		err := service.GetMany[T](c, &dest, options...)
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetListHandler: GetMany failed")
			ResponseError(c, CodeProcessFailed, err)
			return
		}

		var addition []gin.H
		if request.Total {
			total, err := getCount[T](c, request.FilterBy, request.FilterValue)
			if err != nil {
				logger.WithContext(c).WithError(err).
					Warn("GetListHandler: getCount failed")
				addition = append(addition, gin.H{"totalError": err.Error()})
			} else {
				addition = append(addition, gin.H{"total": total})
			}
		}
		ResponseSuccess(c, dest, addition...)
	}
}

// GetByIDHandler handles
//    GET /T/:idParam
//
// QueryOptions (See GetRequestOptions for more details): preload
//
// Response:
//  - 200 OK: { T: {...} }
//  - 400 Bad Request: { error: "request band failed" }
//  - 422 Unprocessable Entity: { error: "get process failed" }
func GetByIDHandler[T orm.Model](idParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request GetRequestOptions
		if err := c.ShouldBind(&request); err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetByIDHandler: bind request failed")
			ResponseError(c, CodeBadRequest, err)
			return
		}

		options := buildQueryOptions(request)

		dest, err := getModelByID[T](c, idParam, options...)
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetByIDHandler: getModelByID failed")
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		ResponseSuccess(c, dest)
	}
}

// GetFieldHandler handles
//    GET /T/:idParam/field
//
// QueryOptions (See GetRequestOptions for more details):
//    limit, offset, order_by, desc, filter_by, filter_value, preload, total.
// Notice, all GetRequestOptions will be conditions for the field, for example:
//    GET /user/123/order?preload=Product
// Preloads User.Order.Product instead of User.Product.
//
// Response:
//  - 200 OK: { Fs: [{...}, ...] }  // field models
//  - 400 Bad Request: { error: "request band failed" }
//  - 422 Unprocessable Entity: { error: "get process failed" }
func GetFieldHandler[T orm.Model](idParam string, field string) gin.HandlerFunc {
	field = nameToField(field, *new(T))

	return func(c *gin.Context) {
		var request GetRequestOptions
		if err := c.ShouldBind(&request); err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetFieldHandler: bind request failed")
			ResponseError(c, CodeBadRequest, err)
			return
		}
		options := buildQueryOptions(request)

		model, err := getModelByID[T](c, idParam, service.Preload(field, options...))
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetFieldHandler: getModelByID failed")
			ResponseError(c, CodeProcessFailed, err)
			return
		}

		fieldValue := reflect.ValueOf(model).
			Elem(). // because model is a pointer
			FieldByName(field)

		var addition []gin.H
		if request.Total && fieldValue.Kind() == reflect.Slice {
			total, err := getAssociationCount(c, model, field, request.FilterBy, request.FilterValue)
			if err != nil {
				logger.WithContext(c).WithError(err).
					Warn("GetFieldHandler: getAssociationCount failed")
				addition = append(addition, gin.H{"totalError": err.Error()})
			} else {
				addition = append(addition, gin.H{"total": total})
			}
		}

		ResponseSuccess(c, fieldValue.Interface(), addition...)
	}
}

func buildQueryOptions(request GetRequestOptions) []service.QueryOption {
	var options []service.QueryOption
	if request.Limit > 0 {
		options = append(options, service.WithPage(request.Limit, request.Offset))
	}
	if request.OrderBy != "" {
		options = append(options, service.OrderBy(request.OrderBy, request.Descending))
	}
	if request.FilterBy != "" && request.FilterValue != "" {
		options = append(options, service.FilterBy(request.FilterBy, request.FilterValue))
	}
	for _, field := range request.Preload {
		// logger.WithField("field", field).Debug("Preload field")
		options = append(options, service.Preload(field))
	}
	return options
}

// getModelByID gets idParam from url and get model from database
func getModelByID[T orm.Model](c *gin.Context, idParam string, options ...service.QueryOption) (*T, error) {
	var model T

	id := c.Param(idParam)
	if id == "" {
		logger.WithContext(c).WithField("idParam", idParam).
			Warn("getModelByID: id is empty")
		return &model, ErrMissingID
	}

	err := service.GetByID[T](c, id, &model, options...)
	return &model, err
}

func getCount[T any](ctx context.Context, filterBy string, filterValue any) (total int64, err error) {
	var option []service.QueryOption
	if filterBy != "" && filterValue != "" {
		option = append(option, service.FilterBy(filterBy, filterValue))
	}
	total, err = service.Count[T](ctx, option...)
	return total, err
}

func getAssociationCount(ctx context.Context, model any, field string, filterBy string, filterValue any) (total int64, err error) {
	var options []service.QueryOption
	if filterBy != "" && filterValue != "" {
		options = append(options, service.FilterBy(filterBy, filterValue))
	}
	count, err := service.CountAssociations(ctx, model, field, options...)
	return count, err
}
