package main

import (
	"time"

	"github.com/sporule/grater/common/queue"
	"github.com/sporule/grater/common/utility"
	"github.com/sporule/grater/scraper"
)

func main() {

	msg := &queue.Message{
		ID:         "1",
		Link:       "https://www.amazon.co.uk/s?k=levis&rh=p_76%3A419158031&dc&qid=1606558926&rnid=419157031&ref=sr_nr_p_76_1",
		Status:     utility.Enums().Status.Active,
		LastUpdate: time.Now(),
	}
	s, _ := scraper.New(`{"product-meta":{"pattern":"div[data-asin][data-component-type=s-search-result]","value":"","children":{"name":{"pattern":"span.a-size-base-plus.a-color-base.a-text-normal","value":"text"},"image":{"pattern":"img[data-image-latency=s-product-image]","value":"attr:src"},"url":{"pattern":"a.a-link-normal.a-text-normal","value":"attr:href"}}},"page":{"pattern":"ul.a-pagination>li.a-selected>a","value":"text"}}`)
	s.Collector.Visit(msg.Link)
	s.Collector.Wait()
}
