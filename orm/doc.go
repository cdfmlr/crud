// Package orm wraps the gorm to provide a simple ORM layer.
//
// It contains a Model interface. All your models (for crud operations) should
// implement it. And there is a BasicModel struct implements Model.
// You can embed it into your model.
//
// Call the ConnectDB() function to connect to the database.
// And call RegisterModel() to register your models.
package orm
