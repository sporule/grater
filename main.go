package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
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

	env := utility.GetEnv("ENV", "dev")
	if env != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}
	//default to run on both mode
	mode := "both"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	switch mode {
	case "dist":
		//dist mode only runs distributor
		runAPI()
	case "scraper":
		for {
			//scraper mode only runs scraper
			scraping()
		}
	case "both":
		//both mode runs both distributor and scraper
		go func() {
			time.Sleep(5 * time.Second)
			for {
				//turn on scraper mode
				scraping()
			}
		}()
		runAPI()
	}
}

func scraping() {
	err := scraper.StartScraping()
	if err != nil {
		log.Println("error occured, wait for 60 seconds before restart:", err)
		time.Sleep(60 * time.Second)
	}
}

func runAPI() {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://www.dealself.com", "http://127.0.0.1:8080"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	apiv1.RegisterAPIRoutes(router)
	router.Run(":" + utility.GetEnv("PORT", "9999"))
}
