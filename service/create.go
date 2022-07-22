package service

import (
	"context"
	"github.com/cdfmlr/crud/orm"
	"gorm.io/gorm"
)

// Create creates a model in the database.
// Nested models associated with the model will be created as well.
//
// There are two mode of creating a model:
// 	- IfNotExist: creates a model record if it does not exist.
// 	- NestInto: creates a nested model of the parent model.
//
// Note:
//
//    user := User{ profile: Profile{ ... } }
//    Create(&user, IfNotExist())     // creates user, user.profile
//
//    group := GetByID[Group](123)
//    Create(&user, NestInto(&group, "users"))
//    // user is already in the database: just add it into group.users
func Create(ctx context.Context, model any, in CreateMode) error {
	return in(ctx, model)
}

// CreateMode is the way to create a model:
//  - IfNotExist: creates a model if it does not exist.
//  - NestInto: creates a nested model of the parent model.
//
// TODO: I dont think it's reasonable to (ab)use functional option pattern here
//      to handle different kinds of creates (CreateMode). It is a temporary
//      solution and should be replaced by seperated functions, say,
//      Create(model) and CreateNested(parentID, field, child).
type CreateMode func(ctx context.Context, modelToCreate any) error

// NestInto creates a nested model of the parent model in the database.
// Say, if you have a model User and a model Profile:
//    CREATE TABLE `user` ( `id` INTEGER PRIMARY KEY AUTOINCREMENT, ...)
//    CREATE TABLE `profile` ( `id` INTEGER PRIMARY KEY AUTOINCREMENT, ...)
//    CREATE TABLE `user_profiles` (  // a join table
//        `user_id` INTEGER NOT NULL,
//        `profile_id` INTEGER NOT NULL,
//        PRIMARY KEY (`user_id`, `profile_id`)
//        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
//        FOREIGN KEY (`profile_id`) REFERENCES `profile` (`id`)
//    )
// You can create a Profile of a User (i.e. the user has the profile) by calling
//    Create(&userProfile, NestInto(&user))
// The userProfile will be created in the database and associated with the user:
//    INSERT INTO profile ...
//    INSERT INTO user_profiles (user_id, profile_id)
//
// This is useful to handle POSTs like /api/users/{user_id}/profile
func NestInto(parent any, field string) CreateMode {
	return func(ctx context.Context, modelToCreate any) error {
		logger.WithContext(ctx).
			WithField("parent", parent).
			WithField("field", field).
			WithField("modelToCreate", modelToCreate).
			Trace("Create Nested")

		return orm.DB.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).
			Model(parent).Association(field).Append(modelToCreate)
	}
}

// IfNotExist creates a model if it does not exist.
func IfNotExist() CreateMode {
	return func(ctx context.Context, modelToCreate any) error {
		logger.WithContext(ctx).
			WithField("modelToCreate", modelToCreate).
			Trace("Create IfNotExist")

		return orm.DB.WithContext(ctx).Create(modelToCreate).Error
	}
}
