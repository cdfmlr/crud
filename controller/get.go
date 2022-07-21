package controller

import (
	"context"
	"fmt"
	"github.com/cdfmlr/crud/orm"
	"github.com/cdfmlr/crud/service"
	"github.com/gin-gonic/gin"
	"reflect"
)

type GetRequestBody struct {
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
//    GET /T?limit=10&offset=0&order_by=id&desc=true&filter_by=name&filter_value=John&total=true
func GetListHandler[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request GetRequestBody
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
func GetByIDHandler[T orm.Model](idParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request GetRequestBody
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
// All GetRequestBody will be conditions for the field, for example:
//    GET /user/123/order?preload=Product
// Preloads User.Order.Product instead of User.Product.
func GetFieldHandler[T orm.Model](idParam string, field string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.WithContext(c).
			WithField("model", fmt.Sprintf("%T", *new(T))).
			WithField("idParam", idParam).
			WithField("field", field).
			Trace("GetFieldHandler")

		// 0. bind request options
		var request GetRequestBody
		if err := c.ShouldBind(&request); err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetFieldHandler: bind request failed")
			ResponseError(c, CodeBadRequest, err)
			return
		}

		// 1. check the model exists?
		model, err := getModelByID[T](c, idParam)
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetFieldHandler: getModelByID failed")
			ResponseError(c, CodeProcessFailed, err)
			return
		}
		logger.WithField("model", model).Debug("GetFieldHandler: model found")

		// 2. find out the field's type F
		// If F is Struct / *Struct / []Struct,
		// then we GetAssociations to query it.
		// Otherwise, we just return field value from getModelByID.
		field := NameToField(field, model)

		// model.field
		fieldValue := reflect.ValueOf(model).
			Elem(). // because model is a pointer
			FieldByName(field)

		if fieldValue.Type().Kind() == reflect.Ptr {
			fieldValue = fieldValue.Elem() // *F => F
		}

		var elemType reflect.Type
		switch fieldValue.Type().Kind() {
		case reflect.Slice, reflect.Array:
			elemType = fieldValue.Type().Elem() // []F => F
		default:
			elemType = fieldValue.Type() // keep F
		}
		// for []*Struct, we have to unwrap it again
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem() // *F => F
		}

		//logger.WithField("fieldType", fieldValue.Type().Kind()).
		//	WithField("elemType", elemType).
		//	Debug("GetFieldHandler: elemType found")

		if elemType.Kind() != reflect.Struct {
			// not a model, return the value
			ResponseSuccess(c, fieldValue.Interface())
			return
		}

		// Slice or Struct
		// 3. GetAssociations
		dest := fieldValue.Interface()
		err = service.GetAssociations(c, model, field, &dest, buildQueryOptions(request)...)
		if err != nil {
			logger.WithContext(c).WithError(err).
				Warn("GetFieldHandler: GetAssociations failed")
			ResponseError(c, CodeProcessFailed, err)
			return
		}

		// 4. Count
		var addition []gin.H
		if request.Total {
			total, err := service.CountAssociations(c, model, field, buildQueryOptions(request)...)
			if err != nil {
				logger.WithContext(c).WithError(err).
					Warn("GetFieldHandler: CountAssociations failed")
				addition = append(addition, gin.H{"totalError": err.Error()})
			} else {
				addition = append(addition, gin.H{"total": total})
			}
		}

		ResponseSuccess(c, dest)
	}
}

func buildQueryOptions(request GetRequestBody) []service.QueryOption {
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
