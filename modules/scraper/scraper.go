package scraper

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"runtime"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/gocolly/colly/proxy"
	"github.com/gocolly/colly/queue"
	"github.com/google/uuid"

	"github.com/sporule/grater/models"
	"github.com/sporule/grater/modules/utility"
)

//scraper is the struct for scraper
type scraper struct {
	ID              string `json:"id,omitempty"`
	Collector       *colly.Collector
	Proxies         []string
	Rule            models.Rule
	Queue           *queue.Queue
	ReceviedLinkIDs []string
	ScrapedRecords  []string
	TableName       string
}

//new creates new scraper
func new() (*scraper, error) {
	id, _ := uuid.NewRandom()
	return &scraper{
		ID: id.String(),
	}, nil
}

func (scraper *scraper) SaveScrapedRecords() error {
	err := models.InsertManyResults(scraper.TableName, scraper.ScrapedRecords)
	if err != nil {
		return err
	}
	scraper.ScrapedRecords = make([]string, 0)
	return nil
}

func (scraper *scraper) setProxies() error {
	//hard coded to obtain sock5 proxy
	proxyLink := "https://api.proxyscrape.com/v2/?request=getproxies&protocol=socks5&timeout=10000&country=all"
	res, err := http.Get(proxyLink)
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print("Unable to obtain Proxy")
		return err
	}
	proxies := strings.Split(string(body), "\r\n")
	scraper.Proxies = make([]string, 0)
	for _, proxy := range proxies {
		scraper.Proxies = append(scraper.Proxies, "socks5://"+proxy)
	}
	return nil
}

func (scraper *scraper) setLinksToComplete() error {
	if api := os.Getenv("DISTRIBUTOR_API"); !utility.IsNil(api) {
		body, err := json.Marshal(map[string][]string{
			"linkids": scraper.ReceviedLinkIDs,
		})
		if err != nil {
			return errors.New("Error on parsing completed link IDs")
		}
		_, err = http.Post(api+"/links", "application/json", bytes.NewBuffer(body))
		if err == nil {
			//reset the completedLinkIDs
			scraper.ReceviedLinkIDs = make([]string, 0)
			return nil
		}
	} else {
		return errors.New("API Not found")
	}
	return nil
}

func (scraper *scraper) setRule() error {
	if api := os.Getenv("DISTRIBUTOR_API"); !utility.IsNil(api) {
		//obtain the highest priority queue
		res, err := http.Get(api + "/rules")
		if err != nil {
			log.Print("Unable to make request to obtain rules", err)
			return err
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Print("Unable to read rules", err)
			return err
		}
		var rules []models.Rule
		err = json.Unmarshal(body, &rules)
		if len(rules) <= 0 {
			log.Print("Unable to find any rules")
			return errors.New("Unable to find any rules")
		}
		rule := rules[0]
		if err != nil {
			log.Print("Unable to parse the returned queue result", err)
			return err
		}
		if utility.IsNil(rule.ID, rule.Pattern, rule.TargetLocation) {
			log.Print("Unable to read rule information")
			return errors.New("Unable to read rule information")
		}
		scraper.Rule = rule
		scraper.TableName = rule.TargetLocation
	} else {
		return errors.New("API Not found")
	}
	return nil
}

func (scraper *scraper) setLinksQueue() error {
	if api := os.Getenv("DISTRIBUTOR_API"); !utility.IsNil(api) {
		//obtain the links
		res, err := http.Get(api + "/links?ruleid=" + scraper.Rule.ID + "&worker=" + scraper.ID)
		if err != nil {
			log.Print("Unable to make request to obtain links ", err)
			return err
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Print("Unable to read links ", err)
			return err
		}
		var links []models.Link
		err = json.Unmarshal(body, &links)
		if err != nil {
			log.Print("Unable to parse the returned links result ", err)
			return err
		}
		log.Println("CPUs: ", runtime.NumCPU())
		scraper.Queue, _ = queue.New(
			runtime.NumCPU(), // Number of consumer threads
			&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
		)
		for _, link := range links {
			scraper.ReceviedLinkIDs = append(scraper.ReceviedLinkIDs, link.ID)
			scraper.Queue.AddURL(link.Link)
		}
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
		err := json.Unmarshal([]byte(scraper.Rule.Pattern), &pattern)
		if !utility.IsNil(err) {
			log.Println("Cannot read the rule pattern", err)
			return
		}
		// parsePattern(e.DOM, pattern)
		value := parsePattern(e.DOM, pattern)
		jsonString, err := json.Marshal(value)
		scraper.ScrapedRecords = append(scraper.ScrapedRecords, string(jsonString))
		//TODO: Save in the database
		//log.Println(value)
		log.Print("Completed")
	})

	c.OnError(func(r *colly.Response, err error) {
		// log.Println("Request URL:", r.Request.URL, "failed with response:", string(r.Body), "\nError:", err)
		log.Println("Failed")
		r.Request.Retry()
	})

	rp, err := proxy.RoundRobinProxySwitcher(scraper.Proxies...)
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

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

//StartScraping fires of the scraping process
func StartScraping() error {
	log.Println("Scraper started")
	scraper, _ := new()
	err := scraper.setRule()
	if !utility.IsNil(err) {
		return err
	}
	//get new proxies every 6 minutes
	go func() {
		for {
			scraper.setProxies()
			scraper.setCollector()
			time.Sleep(6 * time.Minute)
		}
	}()
	//save data to database very minute
	go func() {
		for {
			scraper.SaveScrapedRecords()
			time.Sleep(time.Minute)
		}
	}()
	if err != nil {
		log.Println(err)
		return err
	}
	err = scraper.setLinksQueue()
	if !utility.IsNil(err) {
		return err
	}
	for len(scraper.Proxies) <= 0 {
		log.Println("Waiting for proxies")
		time.Sleep(10 * time.Second)
	}
	scraper.Queue.Run(scraper.Collector)
	scraper.setLinksToComplete()
	scraper.SaveScrapedRecords()
	log.Println("Scraper Completed")
	return nil
}
