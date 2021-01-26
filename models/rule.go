package models

import (
	"errors"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sporule/grater/modules/database"
	"github.com/sporule/grater/modules/utility"
)

//Rule sets the scraper pattern for all links
type Rule struct {
	ID               string    `bson:"_id" json:"id,omitempty"`
	Name             string    `json:"name,omitempty"`
	Status           string    `json:"status,omitempty"`
	Pattern          string    `json:"pattern,omitempty"`
	Priority         int       `json:"priorty,omitempty"`
	TargetLocation   string    `json:"targetLocation,omitempty"`
	LinkPattern      string    `json:"linkPattern,omitempty"`
	DeepLinkPatterns string    `json:"deeplinkPatterns,omitempty"`
	TotalPages       int       `json:"totalPages,omitempty"`
	LastUpdate       time.Time `json:"lastUpdate,omitempty"`
	Headers          string    `json:"headers,omitempty"`
	Frequency        int       `json:"frequency,omitempty"`
}

const ruleTable = "rule"

//NewRule is the constructor of Rule
func NewRule(name, targetLocation, pattern, linkPattern, deeplinkPatterns, headers string, totalPages int) (*Rule, error) {
	if utility.IsNil(name, pattern, targetLocation) {
		return nil, errors.New(utility.Enums().ErrorMessages.LackOfInfo)
	}
	id, _ := uuid.NewRandom()
	return &Rule{
		ID:               id.String(),
		Name:             name,
		Status:           utility.Enums().Status.Active,
		Pattern:          pattern,
		Priority:         0,
		LinkPattern:      linkPattern,
		DeepLinkPatterns: deeplinkPatterns,
		TotalPages:       totalPages,
		Headers:          headers,
		TargetLocation:   targetLocation,
	}, nil
}

//Upsert updates or inserts rule object to database, it will attach the LastUpdate time stamp to time.now()
func (rule *Rule) Upsert() error {
	if utility.IsNil(rule.ID) {
		id, _ := uuid.NewRandom()
		rule.ID = id.String()
	}
	filters := map[string]interface{}{"_id": rule.ID}
	rule.LastUpdate = time.Now()
	return database.Client.UpsertOne(ruleTable, filters, rule)
}

//GenerateLinks generates links based on Link Pattern and Page
func (rule *Rule) GenerateLinks() ([]string, error) {
	//page pattern is {page}
	var links []string
	pagePattern := "{page}"
	if !strings.Contains(rule.LinkPattern, pagePattern) {
		return nil, errors.New(utility.Enums().ErrorMessages.LackOfInfo)
	}
	page := 1
	for page <= rule.TotalPages {
		links = append(links, strings.ReplaceAll(rule.LinkPattern, pagePattern, strconv.Itoa(page)))
		page++
	}
	return links, nil
}

//GenerateAndInsertLinks generates links and Add it to the database, it also resets the incompleted links
func (rule *Rule) GenerateAndInsertLinks() error {
	log.Println("Generating links for rule:", rule.Name)
	CancelInactiveLinks(rule.ID)
	links, err := rule.GenerateLinks()
	if err != nil {
		return err
	}
	err = AddLinksRaw(links, rule.ID)
	return err
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
	err := database.Client.GetAll(ruleTable, &rules, filtersMap, nil, page)
	return rules, err
}

//GetRulesWithActiveLinks returns rules with active link, it returns a slice but with only 1 item
func GetRulesWithActiveLinks() ([]Rule, error) {
	var rules []Rule
	links, err := GetLinks("", "Active", 1)
	if err != nil {
		return nil, err
	}
	index := rand.Intn(len(links))
	//return a random rule with active links
	rule, err := GetRule(links[index].RuleID)
	rules = append(rules, *rule) // return a list because of lazy
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
