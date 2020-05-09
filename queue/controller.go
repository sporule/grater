package queue

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/common/utility"
)

//InitiateRouters register all routes
func InitiateRouters(router *gin.Engine) {
	r := router.Group("/queues/")

	r.GET("/", getQueuesController)
	r.GET("/:qid/messages/request", requestMessageController)

	r.POST("/:qid/messages/:mid", updateMessageController)
}

//getQueuesController returns all queues
func getQueuesController(c *gin.Context) {
	res := make(chan utility.Result)
	go func() {
		res <- utility.Result{http.StatusOK, queues}
		return
	}()
	result := <-res
	c.JSON(result.Expand())
}

//requestMessageController returns a message to client
func requestMessageController(c *gin.Context) {
	cCp := c.Copy()
	res := make(chan utility.Result)
	go func() {
		qID := cCp.Param("qid")
		worker := cCp.DefaultQuery("worker", "")
		if worker == "" {
			res <- utility.Result{http.StatusOK, &utility.Error{"Could not identify the worker"}}
			return
		}
		q, qErr := getQueue(qID)
		if qErr != nil {
			res <- utility.Result{http.StatusOK, &utility.Error{utility.Enums().ErrorMessages.RecordNotFound}}
			return
		}
		msg, msgErr := q.allocateMessage(worker)
		if msgErr != nil {
			res <- utility.Result{http.StatusOK, &utility.Error{utility.Enums().ErrorMessages.RecordNotFound}}
			return
		}
		res <- utility.Result{http.StatusOK, msg}
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
			res <- utility.Result{http.StatusOK, &utility.Error{utility.Enums().ErrorMessages.RecordNotFound}}
			return
		}
		var newMessage message
		bodyErr := cCp.BindJSON(&newMessage)
		if bodyErr != nil {
			res <- utility.Result{http.StatusOK, &utility.Error{utility.Enums().ErrorMessages.LackOfInfo}}
			return
		}
		currentMessage, cmErr := q.getMessage(mID)
		if cmErr != nil {
			res <- utility.Result{http.StatusOK, &utility.Error{utility.Enums().ErrorMessages.RecordExist}}
			return
		}
		updatedMessage, umErr := q.updateMessage(currentMessage.ID, &newMessage)
		if umErr != nil {
			res <- utility.Result{http.StatusOK, &utility.Error{umErr.Error()}}
			return
		}
		res <- utility.Result{http.StatusOK, updatedMessage}
	}()
	result := <-res
	c.JSON(result.Expand())
}
