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
	Worker     string `json:"worker,omitempty"`
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
	return links, database.Client.GetAll(linkTable, &links, filters, page)
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

//AllocateLinks returns a set of new links and it will update the links with the worker and status
func AllocateLinks(ruleID, worker string) ([]Link, error) {
	links, err := GetLinks(ruleID, utility.Enums().Status.Active, 1)
	if err != nil {
		return nil, err
	}
	//update links status to running
	var ids []string
	for _, link := range links {
		ids = append(ids, link.ID)
	}
	filters := map[string]interface{}{"_id": database.Client.InQry(ids)}
	updatesFields := map[string]interface{}{"status": utility.Enums().Status.Running, "worker": worker}
	return links, UpdateManyLinks(filters, updatesFields)
}

//ResetInactiveLinks sets tasks that are inactive for more than 30 minutes back to Active with empty worker
func ResetInactiveLinks() error {
	//default time is 30 minutes
	timeLimit := time.Now().Add(-time.Minute * 30)
	filters := map[string]interface{}{"lastupdate": database.Client.LessThanQry(timeLimit), "status": database.Client.NotEqualQry(utility.Enums().Status.Active)}
	updatesFields := map[string]interface{}{"status": utility.Enums().Status.Active, "worker": ""}
	return UpdateManyLinks(filters, updatesFields)
}
