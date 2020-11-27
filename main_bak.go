package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/distributor/apiv1"
)

func main() {

	router := gin.Default()
	apiv1.RegisterAPIRoutes(router)

	router.Run()
}
