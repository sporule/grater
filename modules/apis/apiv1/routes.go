package apiv1

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/sporule/grater/models"
	"github.com/sporule/grater/modules/apis/apiv1/controllers"
	"github.com/sporule/grater/modules/timerjob"
)

//RegisterAPIRoutes registers all api routers
func RegisterAPIRoutes(router *gin.Engine, mode string) {
	r := router.Group("/api/v1")
	r.GET("/heartbeat", heartbeatController)
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	if mode != "scraper" {
		registerEndpoints(r)
		runTimerJobs()
	}
}

//registerEndpoints register the core end points
func registerEndpoints(router *gin.RouterGroup) {
	controllers.InitiateDistRouters(router)
	controllers.InitiateAdminRouters(router)

}

func heartbeatController(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": "hello world"})
}

//runTimerJobs runs timerjobs to refresh the links
func runTimerJobs() {
	scheduler := gocron.NewScheduler(time.Local)
	rules, err := models.GetRules(nil, 0)
	if err != nil {
		log.Fatal("Can't get rules:", err)
	}
	for _, rule := range rules {
		//setting timer jobs
		if rule.Frequency > 0 {
			//generate rules
			scheduler.Every(uint64(rule.Frequency)).Seconds().Do(timerjob.GenerateLinks, rule)
		}
	}
	//reset dead running links
	scheduler.Every(60).Minutes().StartAt(time.Now().Add(time.Duration(60 * time.Minute))).Do(models.ResetInactiveLinks)
	scheduler.StartAsync()
	log.Println("Timer Jobs registered.")
}
