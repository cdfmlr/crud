// Package controller defines a set of generic RESTful http handlers
// (i.e. controllers) for the CRUD operations:
//   - GET    /models     => GetListHandler[Model]: to retrieve a model list
//   - GET    /models/:id => GetByIDHandler[Model]: to retrieve a model by id
//   - POST   /models     => CreateHandler[Model] : to create a new model
//   - PUT    /models/:id => UpdateHandler[Model] : to update an existing model
//   - DELETE /models/:id => DeleteHandler[Model] : to delete an existing model
package controller

import "github.com/cdfmlr/crud/log"

var logger = log.ZoneLogger("crud/controller")
