package controllers

import (
	"net/http"
	"strconv"

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
	cCp := c.Copy()
	res := make(chan utility.Result)
	go func() {
		tableName := cCp.DefaultQuery("tablename", "")
		pageStr := cCp.DefaultQuery("page", "1")
		if tableName == "" {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.LackOfInfo}}
			return
		}
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			//default page is 1
			page = 1
		}
		//sort by lastupdate
		sortbyMap := make(map[string]interface{})
		sortbyMap["lastupdate"] = -1
		results, err := models.GetResults(tableName, nil, sortbyMap, page)
		if err != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.SystemError}}
			return
		}
		res <- utility.Result{Code: http.StatusOK, Obj: results}
		return
	}()
	result := <-res
	c.JSON(result.Expand())
}
