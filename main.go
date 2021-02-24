package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
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

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		scraping(mode)
	}()
	runAPI(mode)
	wg.Wait()
}

func scraping(mode string) {
	time.Sleep(3 * time.Second)
	if mode != "dist" {
		for {
			log.Println("Started new scraping round, current size of go routine:", runtime.NumGoroutine())
			PrintMemUsage()
			err := scraper.StartScraping()
			if err != nil {
				log.Println("error occured, wait for 60 seconds before restart:", err)
				time.Sleep(60 * time.Second)
			} else {
				log.Println("One round completed, starting next round")
			}
		}
	}
}

func runAPI(mode string) {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{utility.GetEnv("CORS", "http://127.0.0.1:8080"), utility.GetEnv("CORS2", "http://127.0.0.1:8080")},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	apiv1.RegisterAPIRoutes(router, mode)
	router.Run(":" + utility.GetEnv("PORT", "9999"))
}

//PrintMemUsage a
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("go routine Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tgo routine TotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tgo routine Sys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tgo routine NumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
