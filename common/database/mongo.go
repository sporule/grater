package database

import (
	"context"
	"log"
	"time"

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
func (db *MongoDB) GetOne(table string, item, filter interface{}) error {
	return db.client.Collection(table).FindOne(context.TODO(), filter).Decode(item)
}

//GetAll returns all result
func (db *MongoDB) GetAll(table string, items interface{}, filter interface{}, skip, limit int64) error {
	options := &options.FindOptions{}
	options.SetSkip(skip)
	options.SetLimit(limit)
	cursor, err := db.client.Collection(table).Find(context.TODO(), filter, options)
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
func (db *MongoDB) UpdateOne(table string, updateQuery, filter interface{}) error {
	_, err := db.client.Collection(table).UpdateOne(context.TODO(), filter, updateQuery)
	return err
}

//UpdateMany updates many item
func (db *MongoDB) UpdateMany(table string, updateQuery, filter interface{}) error {
	_, err := db.client.Collection(table).UpdateMany(context.TODO(), filter, updateQuery)
	return err
}
