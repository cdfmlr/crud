package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)

// ErrorResponseBody builds the error response body:
//    { error: "error message" }
func ErrorResponseBody(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}

// SuccessResponseBody builds the success response body:
//    { `model`: { ... } }
// where the `model` will be replaced by the model's type name.
// and addition fields can add any k-v to the response body.
func SuccessResponseBody(model any, addition ...gin.H) gin.H {
	var res = gin.H{}

	if model != nil {
		modelName := getResponseModelName(model)
		if modelName != "" {
			res[modelName] = model
		}
	}

	for _, h := range addition {
		for k, v := range h {
			res[k] = v
		}
	}
	return res
}

// get a human-readable model name
func getResponseModelName(model any) string {
	var reflectType = reflect.TypeOf(model)
	var reflectValue = reflect.ValueOf(model)
	// we can only get the model name from a struct type,
	// and if model is a pointer or slice, try to get the element type.
	switch reflectType.Kind() {
	case reflect.Struct:
		return reflectType.Name()
	case reflect.Ptr:
		return getResponseModelName(reflectValue.Elem().Interface())
	case reflect.Slice, reflect.Array:
		if reflectType.Elem().Kind() == reflect.Struct {
			return reflectType.Elem().Name() + "s"
		}
		if reflectType.Elem().Kind() == reflect.Ptr && reflectType.Elem().Elem().Kind() == reflect.Struct {
			return reflectType.Elem().Elem().Name() + "s"
		}
		if reflectValue.Len() > 0 {
			return getResponseModelName(reflectValue.Index(0).Interface()) + "s"
		}
		fallthrough
	default:
		return "data"
	}
}

// ResponseError writes an error response to client in JSON.
func ResponseError(c *gin.Context, code int, err error) {
	c.JSON(code, ErrorResponseBody(err))
}

// ResponseSuccess writes a success response to client in JSON.
func ResponseSuccess(c *gin.Context, model any, addition ...gin.H) {
	c.JSON(http.StatusOK, SuccessResponseBody(model, addition...))
}

const (
	CodeSuccess       = http.StatusOK
	CodeNotFound      = http.StatusNotFound
	CodeBadRequest    = http.StatusBadRequest
	CodeProcessFailed = http.StatusUnprocessableEntity
)

var (
	ErrBindFailed      = errors.New("bind failed")
	ErrMissingID       = errors.New("missing id")
	ErrMissingParentID = errors.New("missing parent id")
	ErrUpdateID        = errors.New("id can not be updated")
)
