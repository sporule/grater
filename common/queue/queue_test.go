package queue

import (
	"errors"
	"testing"

	"github.com/sporule/grater/common/queue/utility"
	"github.com/stretchr/testify/assert"
)

func PrepareQueue() *Queue {
	q := &Queue{Name: "test-queue"}
	link, database, table := "https://google.co.uk", "grater", "test-queue"
	q.AddMessage(link, database, table, nil)
	q.AddMessage(link, database, table, nil)
	q.AddMessage(link, database, table, nil)
	q.Messages[0].Status = utility.Enums().Status.Cancelled
	return q
}

func TestAddMessage(t *testing.T) {
	q := &Queue{Name: "test-queue"}
	link, database, table := "https://google.co.uk", "grater", "test-queue"
	q.AddMessage(link, database, table, nil)
	assert.Equal(t, 1, len(q.Messages), "Queue should contain 1 message")
	assert.NotNil(t, q.Messages[0].ID, "Message id should not be empty")
	assert.NotNil(t, q.Messages[0].Link, "Message Link should not be empty")
	assert.NotNil(t, q.Messages[0].LastUpdate, "Message LastUpdate should not be empty")
}

func TestCancelMessage(t *testing.T) {
	q := PrepareQueue()
	id := q.Messages[1].ID
	err := q.CancelMessage(id)
	assert.Nil(t, err, "Error should be nil as the id is in the messages")
	assert.Equal(t, utility.Enums().Status.Cancelled, q.Messages[0].Status, "Message status should be cancelled")
	errB := q.CancelMessage(id)
	assert.Equal(t, errors.New(utility.Enums().ErrorMessages.RecordNotFound), errB, "Error should not be record not found as the message is already cancelled")
	errC := q.CancelMessage("fakeid")
	assert.Equal(t, errors.New(utility.Enums().ErrorMessages.RecordNotFound), errC, "Error should not be record not found as the id is fake")
}

func TestGetMessage(t *testing.T) {
	q := PrepareQueue()
	message, err := q.GetMessage()
	assert.Nil(t, err, "It should not return any error")
	assert.NotNil(t, message, "It should return a message")
	q.GetMessage()
	messageB, errB := q.GetMessage()
	assert.NotNil(t, errB, "It should return an error because there is no item in message queue")
	assert.Nil(t, messageB, "It should not return any message as there is only 3 message in the queue")
}
