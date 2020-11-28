package queue

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sporule/grater/common/utility"
)

//Queue is the struct for queues
type queue struct {
	ID             string    `json:"id,omitempty"`
	Name           string    `json:"name,omitempty"`
	Status         string    `json:"status,omitempty"`
	Pattern        string    `json:"pattern,omitempty"`
	TargetLocation string    `json:"database,omitempty"`
	Messages       []Message `json:"messages,omitempty"`
	mux            sync.Mutex
}

//queues returns the global queues
var queues []*queue

//Message is the basic Message for Queue
type Message struct {
	ID         string `json:"id,omitempty"`
	Link       string `json:"link,omitempty"`
	Status     string `json:"status,omitempty"`
	Worker     string `json:"worker,omitempty"`
	LastUpdate time.Time
}

//new creates a new queue
func new(name, pattern, targetLocation string) (*queue, error) {
	if utility.IsNil(name, pattern, targetLocation) {
		return nil, errors.New(utility.Enums().ErrorMessages.LackOfInfo)
	}
	id, _ := uuid.NewRandom()
	return &queue{
		Name:           name,
		ID:             id.String(),
		Status:         utility.Enums().Status.Active,
		Pattern:        pattern,
		TargetLocation: targetLocation,
	}, nil
}

//getQueue returns queue from the global queues
func getQueue(id string) (*queue, error) {
	for _, queue := range queues {
		if queue.ID == id {
			return queue, nil
		}
	}
	return nil, errors.New(utility.Enums().ErrorMessages.RecordNotFound)
}

//addQueue add a new queue to the queues
func addQueue(queue *queue) error {
	_, err := getQueue(queue.ID)
	if utility.IsNil(err) {
		//err is nil means the record does exist
		return errors.New(utility.Enums().ErrorMessages.RecordExist)
	}
	queues = append(queues, queue)
	return nil
}

//cancelQueue change the status of the queue to cancelled
func cancelQueue(id string) error {
	q, err := getQueue(id)
	if !utility.IsNil(err) {
		return errors.New(utility.Enums().ErrorMessages.RecordNotFound)
	}
	if q.Status == utility.Enums().Status.Cancelled {
		return errors.New(utility.Enums().ErrorMessages.RecordNotFound)
	}
	q.Status = utility.Enums().Status.Cancelled
	return nil
}

//addMessage adds a new message into the queue
func (q *queue) addMessage(link string) (*Message, error) {
	id, _ := uuid.NewRandom()
	if utility.IsNil(link) {
		return nil, errors.New(utility.Enums().ErrorMessages.LackOfInfo)
	}
	msg := &Message{
		ID:         id.String(),
		Link:       link,
		Status:     utility.Enums().Status.Active,
		LastUpdate: time.Now(),
	}
	q.Messages = append(q.Messages, *msg)
	return msg, nil
}

//addMessages adds a new message into the queue
func (q *queue) addMessages(links []string) error {
	for _, link := range links {
		id, _ := uuid.NewRandom()
		if utility.IsNil(link) {
			return errors.New(utility.Enums().ErrorMessages.LackOfInfo)
		}
		msg := &Message{
			ID:         id.String(),
			Link:       link,
			Status:     utility.Enums().Status.Active,
			LastUpdate: time.Now(),
		}
		q.Messages = append(q.Messages, *msg)
	}
	return nil
}

//updateMessage updates the message, currently only support update status
func (q *queue) updateMessage(id string, newMessage *Message) (*Message, error) {
	q.mux.Lock()
	for index, msg := range q.Messages {
		if msg.ID == id {
			q.Messages[index].Status = newMessage.Status
			q.Messages[index].LastUpdate = time.Now()
			q.mux.Unlock()
			return &q.Messages[index], nil
		}
	}
	q.mux.Unlock()
	return nil, errors.New(utility.Enums().ErrorMessages.RecordNotFound)
}

//allocateMessage returns the first active message
func (q *queue) allocateMessage(worker string) (*Message, error) {
	q.mux.Lock()
	for index, msg := range q.Messages {
		if msg.Status == utility.Enums().Status.Active {
			q.Messages[index].Status = utility.Enums().Status.Running
			q.Messages[index].Worker = worker
			q.mux.Unlock()
			return &q.Messages[index], nil
		}
	}
	q.mux.Unlock()
	return nil, errors.New(utility.Enums().ErrorMessages.RecordNotFound)
}

//allocateMessages returns the active messages
func (q *queue) allocateMessages(worker string, size int) ([]*Message, error) {
	q.mux.Lock()
	messages := []*Message{}
	for index, msg := range q.Messages {
		if msg.Status == utility.Enums().Status.Active {
			q.Messages[index].Status = utility.Enums().Status.Running
			q.Messages[index].Worker = worker
			messages = append(messages, &q.Messages[index])
		}
	}
	q.mux.Unlock()
	if len(messages) > 0 {
		return messages, nil
	}
	return nil, errors.New(utility.Enums().ErrorMessages.RecordNotFound)
}

//getMessage returns the first active message
func (q *queue) getMessage(id string) (*Message, error) {
	for _, msg := range q.Messages {
		if msg.ID == id {
			return &msg, nil
		}
	}
	return nil, errors.New(utility.Enums().ErrorMessages.RecordNotFound)
}
