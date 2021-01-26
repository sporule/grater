package models

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/sporule/grater/modules/database"
	"github.com/sporule/grater/modules/utility"
)

//Link are the links/tasks that are waiting to be scraped
type Link struct {
	ID         string `bson:"_id" json:"id,omitempty"`
	Link       string `json:"link,omitempty"`
	Status     string `json:"status,omitempty"`
	Scraper    string `json:"scraper,omitempty"`
	RuleID     string `json:"ruleID,omitempty"`
	LastUpdate time.Time
}

const linkTable = "link"

//NewLink is the constructor of Rule
func NewLink(link, ruleID string) (*Link, error) {
	if utility.IsNil(link, ruleID) {
		return nil, errors.New(utility.Enums().ErrorMessages.LackOfInfo)
	}
	id, _ := uuid.NewRandom()
	return &Link{
		ID:     id.String(),
		Link:   link,
		Status: utility.Enums().Status.Active,
		RuleID: ruleID,
	}, nil
}

//GetLinks returns links by link's ruleID, status and page
func GetLinks(ruleID, status string, page int) ([]Link, error) {
	var links []Link
	filters := map[string]interface{}{"status": status, "ruleid": ruleID}
	if ruleID == "" {
		filters = map[string]interface{}{"status": status}
	}
	return links, database.Client.GetAll(linkTable, &links, filters, nil, page)
}

//AddLinks inserts a list of links to the database
func AddLinks(links []Link) error {
	linkAsInterface := make([]interface{}, len(links))
	for i, link := range links {
		link.LastUpdate = time.Now()
		linkAsInterface[i] = link
	}
	return database.Client.InsertMany(linkTable, linkAsInterface)
}

//AddLinksRaw inserts a list of link strings to the database
func AddLinksRaw(linkStrs []string, ruleID string) error {
	var links []Link
	for _, linkStr := range linkStrs {
		link, err := NewLink(linkStr, ruleID)
		if err != nil {
			return err
		}
		links = append(links, *link)
	}
	return AddLinks(links)
}

//UpdateManyLinks updates the links by filter
func UpdateManyLinks(filters, updatesFields map[string]interface{}) error {
	updatesFields["lastupdate"] = time.Now()
	return database.Client.UpdateMany(linkTable, filters, updatesFields)
}

//UpdateLinksStatusToComplete sets links to complete status by using ids
func UpdateLinksStatusToComplete(ids []string) error {
	filters := map[string]interface{}{"_id": database.Client.InQry(ids)}
	updatesFields := map[string]interface{}{"status": utility.Enums().Status.Completed}
	return UpdateManyLinks(filters, updatesFields)
}

//AllocateLinks returns a set of new links and it will update the links with the scraper and status
func AllocateLinks(ruleID, scraper string) ([]Link, error) {
	links, err := GetLinks(ruleID, utility.Enums().Status.Active, 1)
	if err != nil {
		return nil, err
	}
	if links == nil {
		return nil, errors.New(utility.Enums().ErrorMessages.RecordNotFound)
	}
	//update links status to running
	var ids []string
	for _, link := range links {
		ids = append(ids, link.ID)
	}
	filters := map[string]interface{}{"_id": database.Client.InQry(ids)}
	updatesFields := map[string]interface{}{"status": utility.Enums().Status.Running, "scraper": scraper}
	return links, UpdateManyLinks(filters, updatesFields)
}

//ResetInactiveLinks sets tasks that are running for more than 1 hour back to Active with empty scraper
func ResetInactiveLinks() error {
	//default time is 60 minutes
	timeLimit := time.Now().Add(-time.Minute * 60)
	filters := map[string]interface{}{"lastupdate": database.Client.LessThanQry(timeLimit), "status": utility.Enums().Status.Running}
	updatesFields := map[string]interface{}{"status": utility.Enums().Status.Active, "scraper": ""}
	return UpdateManyLinks(filters, updatesFields)
}

//CancelInactiveLinks sets the incompleted links status to cancelled for given rule id
func CancelInactiveLinks(ruleID string) error {
	filters := map[string]interface{}{"ruleid": ruleID, "status": database.Client.NotEqualQry(utility.Enums().Status.Completed)}
	updatesFields := map[string]interface{}{"status": utility.Enums().Status.Cancelled, "scraper": ""}
	return UpdateManyLinks(filters, updatesFields)
}
