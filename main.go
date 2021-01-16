package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/modules/apis/apiv1"
	"github.com/sporule/grater/modules/database"
	"github.com/sporule/grater/modules/utility"
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

	// rule, _ := models.NewRule("testRule", "testStorage", "a.text", "https://blog.golang.org", 5)
	// err := rule.Upsert()
	// if err != nil {
	// 	fmt.Printf("%+v\n", err)
	// }

	// rules, err := models.GetRules(nil, 1)
	// if err != nil {
	// 	fmt.Printf("%+v\n", err)
	// }
	// links, err := rules[0].GenerateLinks()
	// err = models.AddLinksRaw(links, rules[0].ID)
	// fmt.Printf("%+v\n", err)

	// links, err := models.AllocateLinks("535b5e1f-6447-4408-bedd-62d3992f3c3e", "worker1")
	// links, err := models.GetLinks("535b5e1f-6447-4408-bedd-62d3992f3c3e", "Running", 1)
	// var linkIDs []string
	// for _, link := range links {
	// 	linkIDs = append(linkIDs, link.ID)
	// }
	// models.UpdateLinksStatusToComplete(linkIDs)
	// fmt.Printf("%+v\n", err)
	// fmt.Printf("%+v\n", links)

	// //queue.InsertQueue()
	// if queues, err := queue.GetQueues(); err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	log.Print(queues)
	// }

	// if !utility.IsNil(os.Getenv("SCRAPER")) {
	// 	if !utility.IsNil(os.Getenv("DISTRIBUTOR")) {
	// 		go func() {
	// 			for {
	// 				//turn on scraper mode
	// 				err := scraper.StartScraping(2)
	// 				if !utility.IsNil(err) {
	// 					log.Println("error occured, wait for 10 seconds before restart")
	// 					time.Sleep(10 * time.Second)
	// 				}
	// 			}
	// 		}()

	// 	} else {
	// 		for {
	// 			//turn on scraper mode
	// 			err := scraper.StartScraping(2)
	// 			if !utility.IsNil(err) {
	// 				log.Println("error occured, wait for 10 seconds before restart")
	// 				time.Sleep(10 * time.Second)
	// 			}
	// 		}
	// 	}
	// }

	// if !utility.IsNil(os.Getenv("DISTRIBUTOR")) {
	// 	//turn on distributor mode
	// router := gin.Default()
	// apiv1.RegisterAPIRoutes(router)
	// router.Run()
	// }
	router := gin.Default()
	apiv1.RegisterAPIRoutes(router)
	router.Run()
}
