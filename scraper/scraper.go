package scraper

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

//scraper is the struct for scraper
type scraper struct {
	ID        string `json:"id,omitempty"`
	Collector *colly.Collector
	Proxies   []string
}

func New(pattern string) (*scraper, error) {
	c := colly.NewCollector()
	extensions.RandomUserAgent(c)

	var patternObj map[string]interface{}
	json.Unmarshal([]byte(pattern), &patternObj)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("referer", "https://www.amazon.co.uk/")
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		getValues(e.DOM, patternObj)
		//TODO: Save in the database

	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// set proxy
	// proxies := getProxies("")
	// rp, err := proxy.RoundRobinProxySwitcher(proxies...)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// c.SetProxyFunc(rp)

	//create scraper
	s := &scraper{
		ID:        "1",
		Collector: c,
	}
	return s, nil
}

func getValues(s *goquery.Selection, item map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	//flag to detect if it is a name or it is attribute object. If it contains pattern, value or children, then this is a attribute object rather than just a name
	nameFlag := true
	var dom *goquery.Selection

	if pattern, ok := item["pattern"]; ok {
		dom = s.Find(pattern.(string))
		nameFlag = false
	}

	if val, ok := item["value"]; ok && val != "" {
		if val.(string) == "text" {
			result["value"] = dom.First().Text()
		} else if attrs := strings.Split(val.(string), ":"); attrs[0] == "attr" {
			result["value"], _ = dom.First().Attr(attrs[1])
		}
		nameFlag = false
	}

	if children, ok := item["children"]; ok {
		dom.Each(func(index int, elem *goquery.Selection) {
			key := strconv.Itoa(index)
			result[key] = getValues(elem, children.(map[string]interface{}))
		})
		nameFlag = false
	}

	if nameFlag {
		for key, value := range item {
			result[key] = getValues(s, value.(map[string]interface{}))
		}
	}

	return result
}

func getProxies(api string) []string {
	//TODO: to be completed later
	proxies := []string{"http://39.109.123.188:3128"}
	return proxies
}
