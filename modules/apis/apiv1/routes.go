package apiv1

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/modules/apis/apiv1/controllers"
)

//RegisterAPIRoutes registers all api routers
func RegisterAPIRoutes(router *gin.Engine) {
	r := router.Group("/api/v1")
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})
	registerEndpoints(r)
}

//registerEndpoints register the core end points
func registerEndpoints(router *gin.RouterGroup) {
	controllers.InitiateDistRouters(router)
}
