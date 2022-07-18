package orm

import (
	"gorm.io/gorm"
)

type Model interface {
	// Identity returns the primary key field of the model.
	// A very common case is that the primary key field is ID.
	Identity() (fieldName string, value any)
}

type BasicModel gorm.Model

func (m BasicModel) Identity() (fieldName string, value any) {
	return "ID", m.ID
}
