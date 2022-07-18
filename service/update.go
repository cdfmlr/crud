package service

import (
	"errors"
	"github.com/cdfmlr/crud/log"
	"github.com/cdfmlr/crud/orm"
)

// Update all fields of an existing model in database.
func Update(model any) (rowsAffected int64, err error) {
	if model == nil {
		return 0, ErrNoRecord
	}
	log.Logger.Debugf("service Update: %#v", model)

	result := orm.DB.Save(model)
	return result.RowsAffected, result.Error
}

var (
	ErrNoRecord        = errors.New("no record found")
	ErrMultipleRecords = errors.New("multiple records found")
)

// UpdateField updates a single fields of an existing model in database.
// It will try to GetByID first, to make sure the model exists, before updating.
func UpdateField[T orm.Model](id any, field string, value interface{}) (rowsAffected int64, err error) {
	var record T
	if err := GetByID[T](id, &record); err != nil {
		return 0, err
	}
	result := orm.DB.Model(&record).Update(field, value)
	return result.RowsAffected, result.Error
}
