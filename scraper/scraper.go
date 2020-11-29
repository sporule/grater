package scraper

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"runtime"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/gocolly/colly/queue"
	"github.com/google/uuid"

	"github.com/sporule/grater/common/utility"
)

//scraper is the struct for scraper
type scraper struct {
	ID        string `json:"id,omitempty"`
	Collector *colly.Collector
	Proxies   []string
	Queue     map[string]interface{}
}

//new creates new scraper
func new() (*scraper, error) {
	id, _ := uuid.NewRandom()
	return &scraper{
		ID: id.String(),
	}, nil
}

//StartScraping fires of the scraping process
func StartScraping(size int) error {
	log.Println("Scraper started")
	scraper, _ := new()
	err := scraper.getQueue()
	if !utility.IsNil(err) {
		return err
	}
	scraper.setCollector()
	err = scraper.scrape(size)
	if !utility.IsNil(err) {
		return err
	}
	log.Println("Scraper Finished")
	return nil
}

func getProxies(api string) []string {
	//TODO: to be completed later
	proxies := []string{"http://39.109.123.188:3128"}
	return proxies
}

func (scraper *scraper) getQueue() error {
	if api := os.Getenv("DISTRIBUTOR_API"); !utility.IsNil(api) {
		//obtain the highest priority queue
		res, err := http.Get(api + "/queues/")
		if err != nil {
			log.Print("Unable to make request to obtain queue", err)
			return err
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Print("Unable to read queue", err)
			return err
		}
		var queues []map[string]interface{}
		err = json.Unmarshal(body, &queues)
		if len(queues) <= 0 {
			log.Print("Unable to find any queue")
			return errors.New("Unable to find any queue")
		}
		queue := queues[0]
		if err != nil {
			log.Print("Unable to parse the returned queue result", err)
			return err
		}
		if utility.IsNil(queue["pattern"], queue["id"], queue["name"], queue["status"], queue["targetLocation"]) {
			log.Print("Unable to read queue information")
			return errors.New("Unable to read queue information")
		}
		scraper.Queue = queue
	} else {
		return errors.New("API Not found")
	}
	return nil
}

func (scraper *scraper) setCollector() error {
	c := colly.NewCollector()
	extensions.RandomUserAgent(c)

	c.OnRequest(func(r *colly.Request) {
		//TODO: Config in Queue
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		var pattern map[string]interface{}
		err := json.Unmarshal([]byte(scraper.Queue["pattern"].(string)), &pattern)
		if !utility.IsNil(err) {
			log.Println("Cannot read the queue pattern", err)
			return
		}
		value := parsePattern(e.DOM, pattern)
		//TODO: Save in the database
		log.Println(value)

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
	scraper.Collector = c
	return nil
}
func parsePattern(s *goquery.Selection, item map[string]interface{}) map[string]interface{} {
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
			result[key] = parsePattern(elem, children.(map[string]interface{}))
		})
		nameFlag = false
	}

	if nameFlag {
		for key, value := range item {
			result[key] = parsePattern(s, value.(map[string]interface{}))
		}
	}

	return result
}

func (scraper *scraper) scrape(size int) error {
	if api := os.Getenv("DISTRIBUTOR_API"); !utility.IsNil(api) {
		//obtain the messages
		res, err := http.Get(api + "/queues/" + scraper.Queue["id"].(string) + "/messages/request?worker=" + scraper.ID + "&size=" + strconv.Itoa(size))
		if err != nil {
			log.Print("Unable to make request to obtain messages ", err)
			return err
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Print("Unable to read messages ", err)
			return err
		}
		var messages []map[string]string
		err = json.Unmarshal(body, &messages)
		if err != nil {
			log.Print("Unable to parse the returned messages result ", err)
			return err
		}
		log.Println("CPUs: ", runtime.NumCPU())
		q, _ := queue.New(
			runtime.NumCPU(), // Number of consumer threads
			&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
		)
		for _, msg := range messages {
			if link, ok := msg["link"]; ok {
				q.AddURL(link)
			}

		}
		q.Run(scraper.Collector)
	} else {
		return errors.New("API Not found")
	}
	return nil
}
