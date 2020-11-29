package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/common/database"
	"github.com/sporule/grater/common/queue"
	"github.com/sporule/grater/common/utility"
	"github.com/sporule/grater/distributor/apiv1"
	"github.com/sporule/grater/scraper"
)

func main() {

	//load configuration to global map[string]string Config
	utility.LoadConfiguration("config/dev.json")

	if utility.Config["ENV"] == "dev" {
		//set environment varilable for dev environment
		os.Setenv("DISTRIBUTOR", "1")
		os.Setenv("SCRAPER", "1")
		os.Setenv("DISTRIBUTOR_API", "http://localhost:8080/api/v1")
		os.Setenv("CONNECTION_URI", "mongodb://root:example@mongo:27017/")
		os.Setenv("DATABASE_NAME", "grater")
	}

	//initiate database
	if uri, dbName := os.Getenv("CONNECTION_URI"), os.Getenv("DATABASE_NAME"); !utility.IsNil(uri, dbName) {
		if err := database.InitiateDB("mongo", uri, dbName); err != nil {
			log.Fatal("Database Connection Failed ", err)
		}
	} else {
		log.Fatal("Failed to obtain database connection information from environment variables")
	}

	//queue.InsertQueue()
	if queues, err := queue.GetQueues(); err != nil {
		log.Fatal(err)
	} else {
		log.Print(queues)
	}

	if !utility.IsNil(os.Getenv("SCRAPER")) {
		if !utility.IsNil(os.Getenv("DISTRIBUTOR")) {
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

	if !utility.IsNil(os.Getenv("DISTRIBUTOR")) {
		//turn on distributor mode
		router := gin.Default()
		apiv1.RegisterAPIRoutes(router)
		router.Run()
	}
}
