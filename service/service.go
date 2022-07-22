// Package service implements the basic CRUD operations for models.
//
// For any not-in-the-box lower level database operations, you can implement
// your own services with the orm.DB (a *gorm.DB) instance.
package service

import "github.com/cdfmlr/crud/log"

// TODO: use orm.Model instead of any

var logger = log.ZoneLogger("crud/service")
