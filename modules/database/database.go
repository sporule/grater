package database

//Database is the interface for storage layer
type Database interface {
	Connect(uri, databaseName string) error
	GetOne(table string, item interface{}, filtersMap map[string]interface{}) error
	GetAll(table string, items interface{}, filtersMap map[string]interface{}, sortByMap map[string]interface{}, page int) error
	InsertOne(table string, item interface{}) error
	InsertMany(table string, items []interface{}) error
	UpdateOne(table string, filtersMap map[string]interface{}, updatedItem interface{}) error
	UpsertOne(table string, filtersMap map[string]interface{}, updatedItem interface{}) error
	UpdateMany(table string, filtersMap map[string]interface{}, updatesFieldsMap map[string]interface{}) error
	UpsertMany(table string, filtersMap map[string]interface{}, updatesFieldsMap map[string]interface{}) error
	InQry(values interface{}) interface{}
	NotInQry(values interface{}) interface{}
	GreaterThanQry(value interface{}) interface{}
	LessThanQry(value interface{}) interface{}
	NotEqualQry(value interface{}) interface{}
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
