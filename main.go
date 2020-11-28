package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sporule/grater/common/utility"
	"github.com/sporule/grater/distributor/apiv1"
)

func main() {

	//load configuration to global map[string]string Config
	utility.LoadConfiguration("config/dev.json")

	if utility.Config["ENV"] == "dev" {
		//set environment varilable for dev environment
		os.Setenv("distributor", "1")
		os.Setenv("scraper", "1")
	}

	if !utility.IsNil(os.Getenv("distributor")) {
		//turn on distributor mode
		router := gin.Default()
		apiv1.RegisterAPIRoutes(router)
		router.Run()
	}

	if !utility.IsNil(os.Getenv("scraper")) {
		//turn on scraper mode

	}

	// msg := &queue.Message{
	// 	ID:         "1",
	// 	Link:       "https://www.amazon.co.uk/s?k=levis&rh=p_76%3A419158031&dc&qid=1606558926&rnid=419157031&ref=sr_nr_p_76_1",
	// 	Status:     utility.Enums().Status.Active,
	// 	LastUpdate: time.Now(),
	// }
	// s, _ := scraper.New(`{"product-meta":{"pattern":"div[data-asin][data-component-type=s-search-result]","value":"","children":{"name":{"pattern":"span.a-size-base-plus.a-color-base.a-text-normal","value":"text"},"image":{"pattern":"img[data-image-latency=s-product-image]","value":"attr:src"},"url":{"pattern":"a.a-link-normal.a-text-normal","value":"attr:href"}}},"page":{"pattern":"ul.a-pagination>li.a-selected>a","value":"text"}}`)
	// s.Collector.Visit(msg.Link)
	// s.Collector.Wait()
}
