package controller

import (
	"reflect"
	"strings"
)

// nameToField converts name to the right field name in the structure.
// For example:
//    type User struct {
//        ID int
//        Name string
//    }
//    fieldId   := NameToFieldOf[User]("id")   // fieldId   == "ID"
//    fieldName := NameToFieldOf[User]("name") // fieldName == "Name"
func nameToField(name string, structure any) string {
	reflectType := reflect.TypeOf(structure)

	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	if reflectType.Kind() != reflect.Struct {
		return name
	}

	name = strings.ToLower(name)
	name = strings.Replace(name, " ", "", -1)
	name = strings.Replace(name, "-", "", -1)
	name = strings.Replace(name, "_", "", -1)
	name = strings.Replace(name, "/", "", -1)

	for i := 0; i < reflectType.NumField(); i++ {
		fieldName := strings.ToLower(reflectType.Field(i).Name)
		if name == fieldName {
			return reflectType.Field(i).Name
		}
	}

	return name
}
