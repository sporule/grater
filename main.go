package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/modules/apis/apiv1"
	"github.com/sporule/grater/modules/database"
	"github.com/sporule/grater/modules/scraper"
	"github.com/sporule/grater/modules/utility"
)

func main() {

	//initiate database
	if uri, dbName := utility.GetEnv("CONNECTION_URI", "mongodb://root:example@mongo:27017/"), utility.GetEnv("DATABASE_NAME", "grater"); !utility.IsNil(uri, dbName) {
		if err := database.InitiateDB("mongo", uri, dbName); err != nil {
			log.Fatal("Database Connection Failed ", err)
			return
		}
	} else {
		log.Fatal("Failed to obtain database connection information from environment variables")
		return
	}

	//default to run on both mode
	mode := "both"
	if len(os.Args) > 0 {
		mode = os.Args[1]
	}

	switch mode {
	case "dist":
		//dist mode only runs distributor
		router := gin.Default()
		apiv1.RegisterAPIRoutes(router)
		router.Run(":" + utility.GetEnv("DISTRIBUTOR_PORT", "9999"))
	case "scraper":
		for {
			//scraper mode only runs scraper
			err := scraper.StartScraping()
			if err != nil {
				log.Println("error occured, wait for 60 seconds before restart:", err)
				time.Sleep(60 * time.Second)
			}
		}
	case "both":
		//both mode runs both distributor and scraper
		go func() {
			time.Sleep(5 * time.Second)
			for {
				//turn on scraper mode
				err := scraper.StartScraping()
				if err != nil {
					log.Println("error occured, wait for 60 seconds before restart", err)
					time.Sleep(60 * time.Second)
				}
			}
		}()
		router := gin.Default()
		apiv1.RegisterAPIRoutes(router)
		router.Run(":" + utility.GetEnv("DISTRIBUTOR_PORT", "9999"))
	}
}
