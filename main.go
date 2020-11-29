package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/common/utility"
	"github.com/sporule/grater/distributor/apiv1"
	"github.com/sporule/grater/scraper"
)

func main() {

	//load configuration to global map[string]string Config
	utility.LoadConfiguration("config/dev.json")

	if utility.Config["ENV"] == "dev" {
		//set environment varilable for dev environment
		os.Setenv("distributor", "1")
		os.Setenv("scraper", "1")
	}

	if !utility.IsNil(os.Getenv("scraper")) {
		if !utility.IsNil(os.Getenv("distributor")) {
			go func() {
				for {
					//turn on scraper mode
					err := scraper.StartScraping(2)
					if !utility.IsNil(err) {
						log.Println("error occured, wait for 10 seconds before restart")
						time.Sleep(10 * time.Second)
					}
				}
			}()

		} else {
			for {
				//turn on scraper mode
				err := scraper.StartScraping(2)
				if !utility.IsNil(err) {
					log.Println("error occured, wait for 10 seconds before restart")
					time.Sleep(10 * time.Second)
				}
			}
		}
	}

	if !utility.IsNil(os.Getenv("distributor")) {
		//turn on distributor mode
		router := gin.Default()
		apiv1.RegisterAPIRoutes(router)
		router.Run()
	}
}
