package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/models"
	"github.com/sporule/grater/modules/utility"
)

//InitiateDistRouters set up all distributor endpoints
func InitiateDistRouters(router *gin.RouterGroup) {

	r := router.Group("/dist")
	r.GET("/rules ", getRulesController)
	r.GET("/links ", allocateLinksController)
	r.POST("/links", completeLinksController)
}

func getRulesController(c *gin.Context) {
	res := make(chan utility.Result)
	go func() {
		//Currently only support getting all rules without pagination
		links, err := models.GetRules(nil, 0)
		if err != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.SystemError}}
			return
		}
		res <- utility.Result{Code: http.StatusOK, Obj: links}
		return
	}()
	result := <-res
	c.JSON(result.Expand())
}

func allocateLinksController(c *gin.Context) {
	cCp := c.Copy()
	res := make(chan utility.Result)
	go func() {
		ruleID := cCp.Param("ruleid")
		worker := cCp.DefaultQuery("worker", "")
		if worker == "" {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: "Could not identify the worker"}}
			return
		}
		links, err := models.AllocateLinks(ruleID, worker)
		if err != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.SystemError}}
			return
		}
		res <- utility.Result{Code: http.StatusOK, Obj: links}
		return
	}()
	result := <-res
	c.JSON(result.Expand())
}

func completeLinksController(c *gin.Context) {
	cCp := c.Copy()
	res := make(chan utility.Result)
	go func() {
		var linksMap map[string][]string
		err := cCp.ShouldBindJSON(&linksMap)
		if err != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.LackOfInfo}}
			return
		}
		linkIDs, ok := linksMap["linkids"]
		if !ok {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.LackOfInfo}}
			return
		}
		err = models.UpdateLinksStatusToComplete(linkIDs)
		if err != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.SystemError}}
			return
		}
		res <- utility.Result{Code: http.StatusOK, Obj: nil}
		return
	}()
	result := <-res
	c.JSON(result.Expand())
}
