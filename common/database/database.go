package database

//Database is the interface for storage layer
type Database interface {
	Connect(uri, databaseName string) error
	GetOne(table string, item interface{}, filtersMap map[string]interface{}) error
	GetAll(table string, items interface{}, filter map[string]interface{}, page int) error
	InsertOne(table string, item interface{}) error
	InsertMany(table string, items []interface{}) error
	UpdateOne(table string, updateQuery, filter interface{}) error
	UpdateMany(table string, updateQuery, filter interface{}) error
}

//Client is the global database instance
var Client Database

//InitiateDB creates database object
func InitiateDB(databaseType, uri, databaseName string) error {
	switch databaseType {
	default:
		//default is mongodb
		Client = &MongoDB{}
	}
	err := Client.Connect(uri, databaseName)
	return err
}
