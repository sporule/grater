package queue

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sporule/grater/common/utility"
	"github.com/stretchr/testify/assert"
)

//prepareQueue inject some messages to the queue
func prepareQueue() *queue {
	id, _ := uuid.NewRandom()
	q := &queue{Name: "test", ID: id.String(), Status: utility.Enums().Status.Active}
	link, database, table := "https://google.co.uk", "grater", "test"
	q.addMessage(link, database, table)
	q.addMessage(link, database, table)
	q.addMessage(link, database, table)
	q.Messages[0].Status = utility.Enums().Status.Cancelled
	return q
}

func cleanEnvironment() {
	//clean global queues
	queues = queues[:0]
}

func TestAddMessage(t *testing.T) {
	q := &queue{Name: "test"}
	link, database, table := "https://google.co.uk", "grater", "test"
	q.addMessage(link, database, table)
	assert.Equal(t, 1, len(q.Messages), "Queue should contain 1 message")
	assert.NotNil(t, q.Messages[0].ID, "Message id should not be empty")
	assert.NotNil(t, q.Messages[0].Link, "Message Link should not be empty")
	assert.NotNil(t, q.Messages[0].LastUpdate, "Message LastUpdate should not be empty")
}

func TestUpdateMessage(t *testing.T) {
	q := prepareQueue()
	id := q.Messages[1].ID
	newMessage := q.Messages[1]
	newMessage.Status = utility.Enums().Status.Cancelled
	_, err := q.updateMessage(id, &newMessage)
	assert.Nil(t, err, "Error should be nil as the id is in the messages")
	assert.Equal(t, newMessage.Status, q.Messages[0].Status, "Message status should be cancelled")
	_, errC := q.updateMessage("fakeid", &newMessage)
	assert.Equal(t, errors.New(utility.Enums().ErrorMessages.RecordNotFound), errC, "Error should not be record not found as the id is fake")
}

func TestGetMessage(t *testing.T) {
	q := prepareQueue()
	msg, err := q.allocateMessage("worker1")
	assert.Nil(t, err, "It should not return any error")
	assert.NotNil(t, msg, "It should return a message")
	assert.Equal(t, "worker1", msg.Worker, "The message should contain worker information")
	q.allocateMessage("worker1")
	msg, errB := q.allocateMessage("worker2")
	assert.NotNil(t, errB, "It should return an error because there is no item in message queue")
	assert.Nil(t, msg, "It should not return any message as there is only 3 message in the queue")

}

func TestGetMessageInfo(t *testing.T) {
	q := prepareQueue()
	msg, err := q.getMessage(q.Messages[0].ID)
	assert.Nil(t, err, "It should not return any error")
	assert.NotNil(t, msg, "It should return a message")
	assert.Equal(t, q.Messages[0].ID, msg.ID, "The ID should be the same")
}

func TestGetQueue(t *testing.T) {
	q := prepareQueue()
	queues = append(queues, q)
	defer cleanEnvironment()
	queue, err := getQueue(q.ID)
	assert.Equal(t, q.ID, queue.ID, "It should return the same id")
	assert.Nil(t, err, "It should not return any error")
}

func TestAddQueue(t *testing.T) {
	q := prepareQueue()
	err := addQueue(q)
	assert.Nil(t, err, "It should not return any error")
	assert.Equal(t, len(queues), 1, "The queues size should be 1 after adding one q into queues")
	errB := addQueue(q)
	assert.NotNil(t, errB, "It should return error because the queue with same name is already in the queues")
	assert.Equal(t, len(queues), 1, "The queues size should still be 1 because the previous function is still active")
}

func TestCancelQueue(t *testing.T) {
	q := prepareQueue()
	queues = append(queues, q)
	err := cancelQueue(q.ID)
	assert.Nil(t, err, "It should not return an error because it can find the ID")
	errB := cancelQueue(q.ID)
	assert.NotNil(t, errB, "It should return an error because the queue is already deactivated")
}
