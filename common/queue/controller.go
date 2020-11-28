package queue

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/common/utility"
)

//InitiateRouters register all routes
func InitiateRouters(r *gin.RouterGroup) {

	//Queues Management
	r.GET("/", getQueuesController)
	r.POST("/", addQueueController)

	//Message Management
	r.GET("/:qid/messages/request", requestMessagesController)
	r.POST("/:qid/messages/:mid", updateMessageController)
	r.POST("/:qid/messages/", addMessageController)
}

//getQueuesController returns all queues
func getQueuesController(c *gin.Context) {
	res := make(chan utility.Result)
	go func() {
		res <- utility.Result{Code: http.StatusOK, Obj: queues}
		return
	}()
	result := <-res
	c.JSON(result.Expand())
}

//addQueueController adds a new queue
func addQueueController(c *gin.Context) {
	cCp := c.Copy()
	res := make(chan utility.Result)
	go func() {
		var q queue
		bodyErr := cCp.BindJSON(&q)
		if bodyErr != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.LackOfInfo}}
			return
		}
		qErr := addQueue(&q)
		if qErr != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.RecordExist}}
			return
		}
		res <- utility.Result{Code: http.StatusOK, Obj: &q}
	}()
	result := <-res
	c.JSON(result.Expand())
}

//requestMessagesController returns a message to client
func requestMessagesController(c *gin.Context) {
	cCp := c.Copy()
	res := make(chan utility.Result)
	go func() {
		qID := cCp.Param("qid")
		worker := cCp.DefaultQuery("worker", "")
		sizeStr := cCp.DefaultQuery("size", "30")
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.SystemError}}
			return
		}
		if worker == "" {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: "Could not identify the worker"}}
			return
		}
		q, qErr := getQueue(qID)
		if qErr != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.RecordNotFound}}
			return
		}
		msgs, msgErr := q.allocateMessages(worker, size)
		if msgErr != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.RecordNotFound}}
			return
		}
		res <- utility.Result{Code: http.StatusOK, Obj: msgs}
		return
	}()
	result := <-res
	c.JSON(result.Expand())
}

//updateMessageController updates the status of the message
func updateMessageController(c *gin.Context) {
	cCp := c.Copy()
	res := make(chan utility.Result)
	go func() {
		qID := cCp.Param("qid")
		mID := cCp.Param("mid")
		q, qErr := getQueue(qID)
		if qErr != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.RecordNotFound}}
			return
		}
		var newMessage Message
		bodyErr := cCp.BindJSON(&newMessage)
		if bodyErr != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.LackOfInfo}}
			return
		}
		currentMessage, cmErr := q.getMessage(mID)
		if cmErr != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.RecordExist}}
			return
		}
		updatedMessage, umErr := q.updateMessage(currentMessage.ID, &newMessage)
		if umErr != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: umErr.Error()}}
			return
		}
		res <- utility.Result{Code: http.StatusOK, Obj: updatedMessage}
	}()
	result := <-res
	c.JSON(result.Expand())
}

//addMessageController adds the message to the queue
func addMessageController(c *gin.Context) {
	cCp := c.Copy()
	res := make(chan utility.Result)
	go func() {
		qID := cCp.Param("qid")
		q, qErr := getQueue(qID)
		if qErr != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.RecordNotFound}}
			return
		}
		var newMessage Message
		bodyErr := cCp.BindJSON(&newMessage)
		if bodyErr != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.LackOfInfo}}
			return
		}
		insertedMessage, umErr := q.addMessage(newMessage.Link)
		if umErr != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: umErr.Error()}}
			return
		}
		res <- utility.Result{Code: http.StatusOK, Obj: insertedMessage}
	}()
	result := <-res
	c.JSON(result.Expand())
}
