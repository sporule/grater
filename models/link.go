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
	ID         string `json:"id,omitempty"`
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

//GetAllLinks returns links by link's ruleID, status and page
func GetAllLinks(ruleID, status string, page int) ([]Link, error) {
	var links []Link
	filters := map[string]interface{}{"status": status, "ruleID": ruleID}
	return links, database.Client.GetAll(ruleTable, &links, filters, page)
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

//UpdateManyLinks updates the links by filter
func UpdateManyLinks(filters, updatesFields map[string]interface{}) error {
	updatesFields["LastUpdate"] = time.Now()
	return database.Client.UpdateMany(linkTable, filters, updatesFields)
}

//AllocateLinks returns a set of new links and it will update the links with the worker and status
func AllocateLinks(ruleID, worker string) ([]Link, error) {
	links, err := GetAllLinks(utility.Enums().Status.Active, ruleID, 1)
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
	filters := map[string]interface{}{"LastUpdate": database.Client.LessThanQry(timeLimit), "status": database.Client.NotEqualQry(utility.Enums().Status.Active)}
	updatesFields := map[string]interface{}{"status": utility.Enums().Status.Active, "worker": ""}
	return UpdateManyLinks(filters, updatesFields)
}
