package database

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/sporule/grater/modules/database/mgoqry"
	"github.com/sporule/grater/modules/utility"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//MongoDB is the mongodb implementation
type MongoDB struct {
	client *mongo.Database
}

//Connect creates a connection pool to mongodb
func (db *MongoDB) Connect(uri, databaseName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		// Can't connect to Mongo server
		log.Fatal("Can't connect to Mongo server ", err)
	}
	db.client = client.Database(databaseName)
	return nil
}

//GetOne returns one result
func (db *MongoDB) GetOne(table string, item interface{}, filtersMap map[string]interface{}) error {
	//convert filters map to filter bson.M
	filters := mgoqry.Bsons(filtersMap)
	return db.client.Collection(table).FindOne(context.TODO(), filters).Decode(item)
}

//GetAll returns all result
func (db *MongoDB) GetAll(table string, items interface{}, filtersMap map[string]interface{}, sortByMap map[string]interface{}, page int) error {
	//set pagination
	itemPerPageStr := utility.GetEnv("ITEM_PER_PAGE", "10")
	itemPerPage, _ := strconv.Atoi(itemPerPageStr)
	skipSize := (page - 1) * itemPerPage
	if page == 0 {
		//set unlimited item per page
		itemPerPage = 99999999
		skipSize = 0
	}
	options := &options.FindOptions{}
	options.SetSkip(int64(skipSize))
	options.SetLimit(int64(itemPerPage))
	if sortByMap != nil {
		options.SetSort(mgoqry.Bsons(sortByMap))
	}

	//convert filters map to filter bson.M
	filters := mgoqry.Bsons(filtersMap)

	//get from the db
	cursor, err := db.client.Collection(table).Find(context.TODO(), filters, options)
	if err != nil {
		return err
	}
	return cursor.All(context.Background(), items)
}

//InsertOne inserts one item to the database
func (db *MongoDB) InsertOne(table string, item interface{}) error {
	_, err := db.client.Collection(table).InsertOne(context.TODO(), item)
	return err
}

//InsertMany inserts many items to the database
func (db *MongoDB) InsertMany(table string, items []interface{}) error {
	_, err := db.client.Collection(table).InsertMany(context.TODO(), items)
	return err
}

//UpdateOne updates one item
func (db *MongoDB) UpdateOne(table string, filtersMap map[string]interface{}, updatedItem interface{}) error {
	filters := mgoqry.Bsons(filtersMap)
	_, err := db.client.Collection(table).UpdateOne(context.TODO(), filters, mgoqry.Bson("$set", updatedItem))
	return err
}

//UpsertOne updates or inserts one item
func (db *MongoDB) UpsertOne(table string, filtersMap map[string]interface{}, updatedItem interface{}) error {
	filters := mgoqry.Bsons(filtersMap)
	_, err := db.client.Collection(table).UpdateOne(context.TODO(), filters, mgoqry.Bson("$set", updatedItem), options.Update().SetUpsert(true))
	return err
}

//UpdateMany updates many items
func (db *MongoDB) UpdateMany(table string, filtersMap map[string]interface{}, updatesFieldsMap map[string]interface{}) error {
	filters := mgoqry.Bsons(filtersMap)
	updatesFields := mgoqry.Bsons(updatesFieldsMap)
	_, err := db.client.Collection(table).UpdateMany(context.TODO(), filters, mgoqry.Bson("$set", updatesFields))
	return err
}

//UpsertMany updates or inserts many items
func (db *MongoDB) UpsertMany(table string, filtersMap map[string]interface{}, updatesFieldsMap map[string]interface{}) error {
	filters := mgoqry.Bsons(filtersMap)
	updatesFields := mgoqry.Bsons(updatesFieldsMap)
	_, err := db.client.Collection(table).UpdateMany(context.TODO(), filters, mgoqry.Bson("$set", updatesFields), options.Update().SetUpsert(true))
	return err
}

//InQry takes list of values and returns "In" query
func (db *MongoDB) InQry(values interface{}) interface{} {
	return mgoqry.Bson("$in", values)
}

//NotInQry takes list of values and returns "NotIn" query
func (db *MongoDB) NotInQry(values interface{}) interface{} {
	return mgoqry.Bson("$nin", values)
}

//GreaterThanQry take a value and returns "gt" query
func (db *MongoDB) GreaterThanQry(value interface{}) interface{} {
	return mgoqry.Bson("$gt", value)
}

//LessThanQry take a value and returns "lt" query
func (db *MongoDB) LessThanQry(value interface{}) interface{} {
	return mgoqry.Bson("$lt", value)
}

//NotEqualQry take a value and returns "not" query
func (db *MongoDB) NotEqualQry(value interface{}) interface{} {
	return mgoqry.Bson("$ne", value)
}
