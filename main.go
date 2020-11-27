package main

import (
	"log"
	"time"

	"github.com/sporule/grater/common/queue"
	"github.com/sporule/grater/common/utility"
	"github.com/sporule/grater/scraper"
)

func main() {

	msg := &queue.Message{
		ID:         "1",
		Link:       "https://bromleyplumbersltd.co.uk/services/drainage-waste/",
		Database:   "database",
		Table:      "table",
		Status:     utility.Enums().Status.Active,
		LastUpdate: time.Now(),
	}
	s, _ := scraper.New(msg)
	s.Collector.Visit(msg.Link)
	s.Collector.Wait()
	log.Println(s)
}
