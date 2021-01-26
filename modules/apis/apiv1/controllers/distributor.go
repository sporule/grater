package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/models"
	"github.com/sporule/grater/modules/utility"
)

//InitiateDistRouters set up all distributor endpoints
func InitiateDistRouters(router *gin.RouterGroup) {

	r := router.Group("/dist")
	r.GET("/rules", getRulesController)
	r.POST("/rules", AddRuleController)
	r.GET("/links", allocateLinksController)
	r.POST("/links", completeLinksController)
}

func getRulesController(c *gin.Context) {
	cCp := c.Copy()
	res := make(chan utility.Result)
	go func() {
		//Currently only support getting all rules without pagination
		isScraper := cCp.DefaultQuery("isscraper", "0")
		var rules []models.Rule
		var err error
		if isScraper == "0" {
			pageStr := cCp.DefaultQuery("page", "1")
			page, err := strconv.Atoi(pageStr)
			if err != nil {
				page = 1
			}
			rules, err = models.GetRules(nil, page)
		} else {
			rules, err = models.GetRulesWithActiveLinks()
		}
		if err != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.SystemError}}
			return
		}
		res <- utility.Result{Code: http.StatusOK, Obj: rules}
		return
	}()
	result := <-res
	c.JSON(result.Expand())
}

func allocateLinksController(c *gin.Context) {
	cCp := c.Copy()
	res := make(chan utility.Result)
	go func() {
		ruleID := cCp.DefaultQuery("ruleid", "")
		scraper := cCp.DefaultQuery("scraper", "")
		if utility.IsNil(scraper, ruleID) {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.LackOfInfo}}
			return
		}
		links, err := models.AllocateLinks(ruleID, scraper)
		if err != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: err.Error()}}
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

//AddRuleController add new rule to the database
func AddRuleController(c *gin.Context) {
	cCp := c.Copy()
	res := make(chan utility.Result)
	go func() {
		var rule models.Rule
		err := cCp.ShouldBindJSON(&rule)
		if err != nil {
			res <- utility.Result{Code: http.StatusOK, Obj: &utility.Error{Error: utility.Enums().ErrorMessages.LackOfInfo}}
			return
		}
		err = rule.Upsert()
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
