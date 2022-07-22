// Package controller implements model based generic CRUD controllers
// (i.e. http handlers) to handle create / read / update / delete requests
// from http clients.
// Since crud use [Gin] web framework as http server, a controller here
// is actually a gin.HandlerFunc:
//
//   - GET    /models     => GetListHandler[Model]: to retrieve a model list
//   - GET    /models/:id => GetByIDHandler[Model]: to retrieve a model by id
//   - POST   /models     => CreateHandler[Model] : to create a new model
//   - PUT    /models/:id => UpdateHandler[Model] : to update an existing model
//   - DELETE /models/:id => DeleteHandler[Model] : to delete an existing model
//
//   - GET    /models/:id/field => GetFieldHandler[Model]     : to retrieve a field (nested model) of a model
//   - POST   /models/:id/field => CreateNestedHandler[Model] : to create a nested model (association)
//   - DELETE /models/:id/field => DeleteNestedHandler[Model] : to delete an association record
//
// The controller are all generic functions, which is available in Go 1.18 and
// later, see [Go generics tutorial] for help if you are not familiar with this
// feature. What you need to notice is that you HAVE TO pass handles the
// type arguments to specify the model type in the function call, because
// go have no way to infer them.
//
// Notice that there is not a UpdateNestedHandler, because:
//    PUT /models/:id/field/:id == PUT /field/:id
//
// [Gin]: https://github.com/gin-gonic/gin
// [Go generics tutorial]: https://go.dev/doc/tutorial/generics
package controller

import "github.com/cdfmlr/crud/log"

var logger = log.ZoneLogger("crud/controller")
