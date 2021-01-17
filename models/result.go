package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/sporule/grater/modules/database"
)

//Result is the scraping result
type Result struct {
	ID         string                 `bson:"_id" json:"id,omitempty"`
	Content    map[string]interface{} `json:"content,omitempty"`
	LastUpdate time.Time              `json:"lastUpdate,omitempty"`
}

const resultTable = "result"

//NewResult is the constructor of Result
func NewResult(content map[string]interface{}) (*Result, error) {
	id, _ := uuid.NewRandom()
	return &Result{
		ID:         id.String(),
		Content:    content,
		LastUpdate: time.Now(),
	}, nil
}

//Insert updates or inserts rule object to database, it will attach the LastUpdate time stamp to time.now()
func (result *Result) Insert() error {
	return database.Client.InsertOne(ruleTable, result)
}
