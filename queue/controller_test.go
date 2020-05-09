package queue

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/common/utility"
	"github.com/stretchr/testify/assert"
)

//PerformRequest is a helper function for testing routers
func PerformRequest(r http.Handler, method, path string, body *bytes.Buffer) *httptest.ResponseRecorder {
	method = strings.ToUpper(method)
	var req *http.Request
	if method == "GET" {
		req, _ = http.NewRequest(method, path, nil)
	} else {
		req, _ = http.NewRequest(method, path, body)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRequestMessageController(t *testing.T) {
	//prepare the queues
	q := prepareQueue()
	addQueue(q)
	router := gin.Default()
	InitiateRouters(router)
	w := PerformRequest(router, "GET", "/queues/"+q.ID+"/messages/request?worker=test-node_127.0.0.1", nil)
	assert.Equal(t, http.StatusOK, w.Code, "The http code should return 200")
	var response map[string]string
	err := json.Unmarshal([]byte(w.Body.String()), &response)
	assert.Nil(t, err, "It should not return any error message")
	id, exists := response["id"]
	assert.Equal(t, true, exists, "The response object should have id property")
	testQueue, _ := getQueue(q.ID)
	message, _ := testQueue.getMessage(id)
	assert.Equal(t, "test-node_127.0.0.1", message.Worker, "The worker should not be test-node_127.0.0.1")

	//Testing Scenario Queue ID not exist
	wB := PerformRequest(router, "GET", "/queues/test-wrong-id/messages/request?worker=test-node_127.0.0.1", nil)
	var responseB map[string]string
	json.Unmarshal([]byte(wB.Body.String()), &responseB)
	_, existsB := responseB["id"]
	assert.Equal(t, false, existsB, "It should return error message as the id does not exist")
	_, existsC := responseB["error"]
	assert.Equal(t, true, existsC, "It should return error message")
}

func TestUpdateMessageController(t *testing.T) {
	//prepare the queues
	q := prepareQueue()
	addQueue(q)
	router := gin.Default()
	InitiateRouters(router)
	newMessage := q.Messages[0]
	newMessage.Status = utility.Enums().Status.Finished
	postBody, _ := json.Marshal(newMessage)
	w := PerformRequest(router, "POST", "/queues/"+q.ID+"/messages/"+newMessage.ID, bytes.NewBuffer(postBody))
	assert.Equal(t, http.StatusOK, w.Code, "The http code should return 200")
	var response map[string]string
	err := json.Unmarshal([]byte(w.Body.String()), &response)
	assert.Nil(t, err, "It should not return any error message")
	id, _ := response["id"]
	assert.Equal(t, newMessage.ID, id, "The response object should have the same id")
	assert.Equal(t, utility.Enums().Status.Finished, q.Messages[0].Status, "The status should be updated to finished")

}

// func TestUpdateMessageController(t *testing.T) {
// 	//prepare the queues
// 	q := prepareQueue()
// 	addQueue(q)
// 	router := gin.Default()
// 	InitiateRouters(router)
// 	w := PerformRequest(router, "POST", "/queues/"+q.ID+"/message?worker=test-node_127.0.0.1")
// 	assert.Equal(t, http.StatusOK, w.Code, "The http code should return 200")
// 	var response map[string]string
// 	err := json.Unmarshal([]byte(w.Body.String()), &response)
// 	assert.Nil(t, err, "It should not return any error message")
// 	id, exists := response["id"]
// 	assert.Equal(t, true, exists, "The response object should have link property")
// 	testQueue, _ := getQueue(q.ID)
// 	message, _ := testQueue.getMessageInfo(id)
// 	assert.Equal(t, "test-node_127.0.0.1", message.Worker, "The worker should not be test-node_127.0.0.1")

// 	//Testing Scenario Queue ID not exist
// 	wB := PerformRequest(router, "GET", "/queues/test-wrong-id/message?worker=test-node_127.0.0.1")
// 	var responseB map[string]string
// 	json.Unmarshal([]byte(wB.Body.String()), &responseB)
// 	_, existsB := responseB["id"]
// 	assert.Equal(t, false, existsB, "It should return error message as the id does not exist")
// 	_, existsC := responseB["error"]
// 	assert.Equal(t, true, existsC, "It should return error message")
// }
