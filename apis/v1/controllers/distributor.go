package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/models"
	"github.com/sporule/grater/modules/utility"
)

//getRulesController returns all rules
func getRulesController(c *gin.Context) {
	res := make(chan utility.Result)
	go func() {
		links, err := models.GetRules(nil, 1)
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

func getLinksController(c *gin.Context) {
	cCp := c.Copy()
	res := make(chan utility.Result)
	go func() {
		linkID := cCp.Param("linkId")
		worker := cCp.DefaultQuery("worker", "")
		if worker == "" {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: "Could not identify the worker"}}
			return
		}
		links, err := models.AllocateLinks(linkID, worker)
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
