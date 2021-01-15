package models

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/sporule/grater/modules/database"
	"github.com/sporule/grater/modules/utility"
)

//Rule sets the scraper pattern for all links
type Rule struct {
	ID             string    `bson:"_id" json:"id,omitempty"`
	Name           string    `json:"name,omitempty"`
	Status         string    `json:"status,omitempty"`
	Pattern        string    `json:"pattern,omitempty"`
	Priority       int       `json:"priorty,omitempty"`
	TargetLocation string    `json:"targetLocation,omitempty"`
	LinkPattern    string    `json:"linkPattern,omitempty"`
	Pages          int       `json:"pages,omitempty"`
	LastUpdate     time.Time `json:"lastUpdate,omitempty"`
}

const ruleTable = "rule"

//NewRule is the constructor of Rule
func NewRule(name, targetLocation, pattern, linkPattern string, pages int) (*Rule, error) {
	if utility.IsNil(name, pattern, targetLocation) {
		return nil, errors.New(utility.Enums().ErrorMessages.LackOfInfo)
	}
	id, _ := uuid.NewRandom()
	return &Rule{
		ID:             id.String(),
		Name:           name,
		Status:         utility.Enums().Status.Active,
		Pattern:        pattern,
		LinkPattern:    linkPattern,
		Pages:          pages,
		TargetLocation: targetLocation,
	}, nil
}

//Upsert updates or inserts rule object to database, it will attach the LastUpdate time stamp to time.now()
func (rule *Rule) Upsert() error {
	filters := map[string]interface{}{"_id": rule.ID}
	rule.LastUpdate = time.Now()
	return database.Client.UpsertOne(ruleTable, filters, rule)
}

//GetRule returns rule by ID
func GetRule(id string) (*Rule, error) {
	var rule Rule
	filters := map[string]interface{}{"_id": id}
	err := database.Client.GetOne(ruleTable, &rule, filters)
	return &rule, err
}

//GetRules returns rule by fitlers
func GetRules(filtersMap map[string]interface{}, page int) ([]Rule, error) {
	var rules []Rule
	err := database.Client.GetAll(ruleTable, &rules, filtersMap, page)
	return rules, err
}

//CancelRule Sets the rule status to cancel by ID
func CancelRule(id string) error {
	rule, err := GetRule(id)
	if err != nil {
		return errors.New(utility.Enums().ErrorMessages.RecordNotFound)
	}
	if rule.Status == utility.Enums().Status.Cancelled {
		return errors.New(utility.Enums().ErrorMessages.RecordNotFound)
	}
	rule.Status = utility.Enums().Status.Cancelled
	if rule.Upsert() != nil {
		return errors.New(utility.Enums().ErrorMessages.SystemError)
	}
	//TODO: Cancel All Links under the Rule
	return nil
}
