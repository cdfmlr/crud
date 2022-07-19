package service

import (
	"context"
	"github.com/cdfmlr/crud/orm"
)

// Delete a model from database.
func Delete(ctx context.Context, model any) (rowsAffected int64, err error) {
	logger.WithContext(ctx).
		WithField("model", model).Trace("Delete model")
	result := orm.DB.WithContext(ctx).Delete(model)
	return result.RowsAffected, result.Error
}

// DeleteByID deletes a model from database by its ID.
func DeleteByID[T orm.Model](ctx context.Context, id any) (rowsAffected int64, err error) {
	logger.WithContext(ctx).
		WithField("id", id).
		Trace("DeleteByID: Delete model by ID")

	var model T
	if err := GetByID[T](ctx, id, &model); err != nil {
		logger.WithContext(ctx).
			WithField("id", id).WithError(err).
			Warn("DeleteByID: GetByID failed")
		return 0, err
	}
	result := orm.DB.WithContext(ctx).Delete(&model)
	if result.Error != nil {
		logger.WithContext(ctx).
			WithError(result.Error).Warn("DeleteByID: failed")
	}
	return result.RowsAffected, result.Error
}

// DeleteNested remove the association between parent and child.
func DeleteNested[P orm.Model, T any](ctx context.Context, parent *P, field string, child *T) error {
	err := orm.DB.WithContext(ctx).Model(parent).Association(field).Delete(child)
	if err != nil {
		logger.WithContext(ctx).
			WithError(err).Warn("DeleteNested: failed")
		return err
	}
	// TODO: check if child has no more parents, if none, delete it
	// ™️ 砂仁还要猪心啊 /( ◕‿‿◕ )\
	return err
}

// DeleteNestedByID remove the association between parent and child.
func DeleteNestedByID[P orm.Model, T orm.Model](ctx context.Context, parentID any, field string, childID any) error {
	logger.WithContext(ctx).
		WithField("parentID", parentID).
		WithField("field", field).
		WithField("childID", childID).
		Trace("DeleteNestedByID")

	var parent P
	if err := GetByID[P](ctx, parentID, &parent); err != nil {
		logger.WithContext(ctx).
			WithField("parentID", parentID).WithError(err).
			Warn("DeleteNestedByID: GetByID[Parent] failed")
		return err
	}

	var child T
	if err := GetByID[T](ctx, childID, &child); err != nil {
		logger.WithContext(ctx).
			WithField("childID", childID).WithError(err).
			Warn("DeleteNestedByID: GetByID[Child] failed")
		return err
	}

	return DeleteNested(ctx, &parent, field, &child)
}
