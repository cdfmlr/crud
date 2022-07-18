package service

import "github.com/cdfmlr/crud/orm"

// Delete a model from database.
func Delete(model any) (rowsAffected int64, err error) {
	result := orm.DB.Delete(model)
	return result.RowsAffected, result.Error
}

// DeleteByID deletes a model from database by its ID.
func DeleteByID[T orm.Model](id any) (rowsAffected int64, err error) {
	var model T
	if err := GetByID[T](id, &model); err != nil {
		return 0, err
	}
	result := orm.DB.Delete(&model)
	return result.RowsAffected, result.Error
}

// DeleteNested remove the association between parent and child.
func DeleteNested[P orm.Model, T any](parent *P, field string, child *T) error {
	err := orm.DB.Model(parent).Association(field).Delete(child)
	// TODO: check if child has no more parents, if none, delete it
	// ™️ 砂仁还要猪心啊 /( ◕‿‿◕ )\
	return err
}

// DeleteNestedByID remove the association between parent and child.
func DeleteNestedByID[P orm.Model, T orm.Model](parentID any, field string, childID any) error {
	var parent P
	if err := GetByID[P](parentID, &parent); err != nil {
		return err
	}

	var child T
	if err := GetByID[T](childID, &child); err != nil {
		return err
	}

	return DeleteNested(&parent, field, &child)
}
