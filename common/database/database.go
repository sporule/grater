package database

//Database is the interface for storage layer
type Database interface {
	Connect(uri, databaseName string) error
	GetOne(table string, item, filter interface{}) error
	GetAll(table string, items, filter interface{}, skip, limit int64) error
	InsertOne(table string, item interface{}) error
	InsertMany(table string, items []interface{}) error
	UpdateOne(table string, updateQuery, filter interface{}) error
	UpdateMany(table string, updateQuery, filter interface{}) error
}

//DBInstance is the global database instance
var DBInstance Database

//InitiateDB creates database object
func InitiateDB(databaseType, uri, databaseName string) error {
	switch databaseType {
	default:
		//default is mongodb
		DBInstance = &MongoDB{}
	}
	err := DBInstance.Connect(uri, databaseName)
	return err
}
