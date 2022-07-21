package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/cdfmlr/crud/orm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Get fetch a single model T into dest.
//
// Shout out to GORM for the feature called Smart Select Fields:
//   https://gorm.io/docs/advanced_query.html#Smart-Select-Fields ,
// we can Get a specific part of fields of model T into a "view model" struct.
// So the generic type T is the type of the Model (which mapping to a scheme,
// i.e. a table in the database), while the parameter dest given the type of the
// view model (for API responses or any other usage). Of course, you can always
// use the original model struct T as its view model, in which case the dest
// parameter should be a *T.
//
// Use FilterBy and Where options to query on specific fields, and by adding
// Preload options, you can preload relationships, for example:
//     Get[User](&user, FilterBy("id", 10), Preload("Sessions")))
// means:
//     SELECT * FROM users WHERE id = 10;          // into dest
//     SELECT * FROM sessions WHERE user_id = 10;  // into user.Sessions
// Because this getting model by id is a common operation, a shortcut GetByID
// is provided. (but you still have to add Preload options if needed)
func Get[T any](ctx context.Context, dest any, options ...QueryOption) error {
	vT := *new(T)
	logger := logger.WithContext(ctx).
		WithField("model", fmt.Sprintf("%T", vT)).
		WithField("dest", fmt.Sprintf("%T", dest))

	logger.Trace("Get model into dest")

	query := orm.DB.WithContext(ctx).Model(new(T))
	for _, option := range options {
		query = option(query)
	}
	ret := query.Take(dest)

	if ret.Error != nil {
		logger.WithError(ret.Error).
			WithField("model", fmt.Sprintf("%T", vT)).
			WithField("dest", fmt.Sprintf("%T", dest)).
			Warnf("Get[%T] into %T failed", vT, dest)
	}

	return ret.Error
}

// GetByID is a shortcut for Get[T](&T, FilterBy("id", id))
//
// Notice: "id" here is the column (or field) name of the primary key of the
// model which is indicated by the Identity method of orm.Model.
// So GetByID only works for models that implement the orm.Model interface.
func GetByID[T orm.Model](ctx context.Context, id any, dest any, options ...QueryOption) error {
	logger.WithContext(ctx).WithField("model", fmt.Sprintf("%T", *new(T))).
		WithField("dest", fmt.Sprintf("%T", dest)).
		Trace("GetByID: Get model by id")

	if id == nil {
		logger.WithContext(ctx).Warn("GetByID skipped: id is nil")
		return ErrNilID
	}
	idField, _ := (*new(T)).Identity()
	if idField == "" {
		logger.WithContext(ctx).Warn("GetByID skipped: unknown id field")
		return ErrNoIdentityField
	}
	options = append(options, FilterBy(idField, id))
	return Get[T](ctx, dest, options...)
}

// GetMany returns a list of models T into dest.
// The dest should be a pointer to a slice of "view model" struct (e.g. *[]*T).
// See the documentation of Get function above for more details.
//
// Adding options parameters, you can query with specific conditions:
//   - WithPage(limit, offset) => pagination
//   - OrderBy(field, descending) => ordering
//   - FilterBy(field, value) => WHERE field=value condition
//      - Where(query, args...) => for more complicated queries
//   - Preload(field) => preload a relationship
//      - PreloadAll() => preload all associations
//
// Example:
//     GetMany[User](&users,
//                   WithPage(10, 0),
//                   OrderBy("age", true),
//                   FilterBy("name", "John"))
// means:
//     SELECT * FROM users
//         WHERE name = "John"
//         ORDER BY age desc
//         LIMIT 10 OFFSET 0;  // into users
func GetMany[T any](ctx context.Context, dest any, options ...QueryOption) error {
	logger := logger.WithContext(ctx).
		WithField("model", fmt.Sprintf("%T", *new(T))).
		WithField("dest", fmt.Sprintf("%T", dest))
	logger.Trace("GetMany: Get models into dest")

	query := orm.DB.WithContext(ctx).Model(new(T))
	for _, option := range options {
		query = option(query)
	}
	ret := query.Find(dest)
	if ret.Error != nil {
		logger.WithError(ret.Error).
			Warn("GetMany: Get models into dest failed")
	}
	return ret.Error
}

// Count returns the number of models.
func Count[T any](ctx context.Context, options ...QueryOption) (count int64, err error) {
	logger := logger.WithContext(ctx).
		WithField("model", fmt.Sprintf("%T", *new(T)))
	logger.Trace("Count: Count models")

	query := orm.DB.WithContext(ctx).Model(new(T))
	for _, option := range options {
		query = option(query)
	}
	ret := query.Count(&count)
	if ret.Error != nil {
		logger.WithError(ret.Error).Warn("Count: Count models failed")
		err = ret.Error
	}
	return count, ret.Error
}

// GetAssociations find matched associations (model.field) into dest.
func GetAssociations(ctx context.Context, model any, field string, dest any, options ...QueryOption) error {
	logger := logger.WithContext(ctx).
		WithField("model", fmt.Sprintf("%T", model)).
		WithField("field", field).
		WithField("dest", fmt.Sprintf("%T", dest))

	logger.Trace("GetAssociation: Get association into dest")

	err := associationQuery(ctx, model, field, options...).Find(dest)
	if err != nil {
		logger.WithError(err).
			Warn("GetAssociation: Get association into dest failed")
	}
	return err
}

// CountAssociations count matched associations (model.field).
func CountAssociations(ctx context.Context, model any, field string, options ...QueryOption) (count int64, err error) {
	logger.WithContext(ctx).
		WithField("model", fmt.Sprintf("%T", model)).
		WithField("field", field).
		Trace("CountAssociations: Count associations")

	count = associationQuery(ctx, model, field, options...).Count()
	return count, err
}

// associationQuery builds a gorm association query
func associationQuery(ctx context.Context, model any, field string, options ...QueryOption) *gorm.Association {
	query := orm.DB.WithContext(ctx).Model(model)
	for _, option := range options {
		query = option(query)
	}
	return query.Association(field)
}

// QueryOption is a function that can be used to construct a query.
type QueryOption func(tx *gorm.DB) *gorm.DB

// Preload preloads a relationship (eager loading).
// It can be applied multiple times (for multiple preloads).
// And nested preloads (like "User.Sessions") are supported.
//
// Passing QueryOptions to custom preloading SQL, see
// https://gorm.io/docs/preload.html#Custom-Preloading-SQL
func Preload(field string, options ...QueryOption) QueryOption {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Preload(field, func(tx *gorm.DB) *gorm.DB {
			for _, option := range options {
				tx = option(tx)
			}
			return tx
		})
	}
}

// PreloadAll to Preload all associations.
// clause.Associations won’t preload nested associations!
func PreloadAll() QueryOption {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Preload(clause.Associations)
	}
}

// WithPage is a query option that sets pagination for GetMany.
func WithPage(limit int, offset int) QueryOption {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Limit(limit).Offset(offset)
	}
}

// OrderBy is a query option that sets ordering for GetMany.
// It can be applied multiple times (for multiple orders).
func OrderBy(field string, descending bool) QueryOption {
	order := field
	if descending {
		order += " desc"
	}
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Order(order)
	}
}

// FilterBy is a query option that sets WHERE field=value condition for GetMany.
// It can be applied multiple times (for multiple conditions).
//
// Example:
//     GetMany[User](&users, FilterBy("name", "John"), FilterBy("age", 10))
// means:
//     SELECT * FROM users WHERE name = "John" AND age = 10 ;  // into users
func FilterBy(field string, value any) QueryOption {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(map[string]any{field: value})
	}
}

// Where offers a more flexible way to set WHERE conditions.
// Equivalent to gorm.DB.Where(...), see:
//   https://gorm.io/docs/query.html#Conditions
//
// Example:
//     GetMany[User](&users, Where("name = ? AND age > ?", "John", 10))
// means:
//     SELECT * FROM users WHERE name = "John" AND age > 10 ;  // into users
func Where(query any, args ...any) QueryOption {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(query, args...)
	}
}

var (
	ErrNoIdentityField = errors.New("no identity field found")
	ErrNilID           = errors.New("id is nil")
)
