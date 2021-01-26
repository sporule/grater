package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/sporule/grater/modules/database"
)

//Result is the scraping result
type Result struct {
	ID         string    `bson:"_id" json:"id,omitempty"`
	Content    string    `json:"content,omitempty"`
	LastUpdate time.Time `json:"lastUpdate,omitempty"`
}

//NewResult is the constructor of Result
func NewResult(content string) (*Result, error) {
	id, _ := uuid.NewRandom()
	return &Result{
		ID:         id.String(),
		Content:    content,
		LastUpdate: time.Now(),
	}, nil
}

//InsertManyResults inserts results to the target table
func InsertManyResults(tableName string, results []string) error {
	resultsInterface := make([]interface{}, len(results))
	for i, result := range results {
		resultTemp, _ := NewResult(result)
		resultsInterface[i] = resultTemp
	}
	return database.Client.InsertMany(tableName, resultsInterface)
}

//GetResults returns results by fitlers
func GetResults(tableName string, filtersMap map[string]interface{}, sortByMap map[string]interface{}, page int) ([]Result, error) {
	var results []Result
	err := database.Client.GetAll(tableName, &results, filtersMap, sortByMap, page)
	return results, err
}
