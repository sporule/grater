package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/models"
	"github.com/sporule/grater/modules/utility"
)

//InitiateAdminRouters set up all distributor endpoints
func InitiateAdminRouters(router *gin.RouterGroup) {

	r := router.Group("/admin")
	r.GET("/results", getResultsController)
}

func getResultsController(c *gin.Context) {
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
