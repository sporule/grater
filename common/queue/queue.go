package queue

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sporule/grater/common/queue/utility"
)

//Queue is the struct for queues
type Queue struct {
	Name     string
	Messages []Message
	mux      sync.Mutex
}

//Message is the basic message for Queue
type Message struct {
	ID           string
	Link         string
	Database     string
	Table        string
	Status       string
	LastUpdate   time.Time
	QueryStrings map[string]string
}

//New creates a new queue
func New(name string) (*Queue, error) {
	if utility.IsNil(name) {
		return nil, errors.New(utility.Enums().ErrorMessages.LackOfInfo)
	}
	return &Queue{
		Name: name,
	}, nil
}

//AddMessage adds a new message into the queue
func (q *Queue) AddMessage(link, database, table string, queryStrings map[string]string) error {
	id, _ := uuid.NewRandom()
	if utility.IsNil(link, database, table) {
		return errors.New(utility.Enums().ErrorMessages.LackOfInfo)
	}
	if utility.IsNil(queryStrings) {
		queryStrings = make(map[string]string)
	}
	message := &Message{
		ID:           id.String(),
		Link:         link,
		Database:     database,
		Table:        table,
		QueryStrings: queryStrings,
		Status:       utility.Enums().Status.Active,
		LastUpdate:   time.Now(),
	}
	q.Messages = append(q.Messages, *message)
	return nil
}

//CancelMessage updates the message status to cancelled
func (q *Queue) CancelMessage(id string) error {
	q.mux.Lock()
	for index, message := range q.Messages {
		if message.ID == id && message.Status != utility.Enums().Status.Cancelled {
			q.Messages[index].Status = utility.Enums().Status.Cancelled
			q.mux.Unlock()
			return nil
		}
	}
	q.mux.Unlock()
	return errors.New(utility.Enums().ErrorMessages.RecordNotFound)
}

//GetMessage returns the first active message
func (q *Queue) GetMessage() (*Message, error) {
	q.mux.Lock()
	for index, message := range q.Messages {
		if message.Status == utility.Enums().Status.Active {
			q.Messages[index].Status = utility.Enums().Status.Running
			q.mux.Unlock()
			return &message, nil
		}
	}
	q.mux.Unlock()
	return nil, errors.New(utility.Enums().ErrorMessages.RecordNotFound)
}
