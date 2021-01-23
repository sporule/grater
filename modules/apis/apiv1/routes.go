package apiv1

import (
	"log"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/sporule/grater/models"
	"github.com/sporule/grater/modules/apis/apiv1/controllers"
)

//RegisterAPIRoutes registers all api routers
func RegisterAPIRoutes(router *gin.Engine) {
	r := router.Group("/api/v1")
	router.NoRoute(func(c *gin.Context) {
		c.HTML(404, "404.html", gin.H{})
	})
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	registerEndpoints(r)
	runTimerJobs()
}

//registerEndpoints register the core end points
func registerEndpoints(router *gin.RouterGroup) {
	controllers.InitiateDistRouters(router)
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
			scheduler.Every(uint64(rule.Frequency)).Seconds().Do(rule.GenerateAndInsertLinks)
		}
	}
	//reset dead running links
	scheduler.Every(30).Minutes().Do(models.ResetInactiveLinks)
	scheduler.StartAsync()
	log.Println("Timer Jobs registered.")
}
