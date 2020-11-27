package scraper

import (
	"fmt"

	"github.com/gocolly/colly"
	"github.com/sporule/grater/common/queue"
)

//scraper is the struct for scraper
type scraper struct {
	ID        string `json:"id,omitempty"`
	Collector *colly.Collector
}

func New(message *queue.Message) (*scraper, error) {
	c := colly.NewCollector()
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	c.OnHTML("body", func(e *colly.HTMLElement) {
		// text := e.ChildText("div[class=elementor-text-editor]>p")
		// fmt.Printf("hELLO", text)
		// e.ForEach("div[class=elementor-text-editor]>p", func(_ int, elem *colly.HTMLElement) {
		// 	if strings.Contains(elem.Text, "golang") {
		// 		fmt.Println(elem.Text)
		// 	}
		// })
		//text := e.ChildText("div.elementor-text-editor")
		//match multiple soup.select('div#top div.foo.bar')
		e.ForEach("div.elementor-text-editor.elementor-clearfix", func(_ int, elem *colly.HTMLElement) {
			fmt.Println(elem.Text)
		})
	})
	s := &scraper{
		ID:        "1",
		Collector: c,
	}
	return s, nil
}
