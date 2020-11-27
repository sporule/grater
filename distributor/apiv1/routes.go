package apiv1

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/common/queue"
)

//RegisterAPIRoutes registers all api routers
func RegisterAPIRoutes(router *gin.Engine) {
	r := router.Group("/api/v1")
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	registerEndpoints(r)
}

//registerEndpoints register the core end points
func registerEndpoints(router *gin.RouterGroup) {
	r := router.Group("/queues/")

	//queue service
	queue.InitiateRouters(r)
}
