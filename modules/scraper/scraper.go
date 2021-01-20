package scraper

import (
	"bytes"
	"encoding/json"
	"errors"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

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
	ParentLinks     map[string]string
	Headers         map[string]string
}

//new creates new scraper
func new() (*scraper, error) {
	id, _ := uuid.NewRandom()
	return &scraper{
		ID:          id.String(),
		ParentLinks: make(map[string]string),
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
	//hard coded to obtain http proxy
	proxyLink := "https://api.proxyscrape.com/v2/?request=getproxies&protocol=http&timeout=10000&country=all&ssl=all&anonymity=all"
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
	validatedProxies := proxyCheck(proxies)
	//TODO, Proxy is not working
	log.Println("Proxies:", len(proxies), "Validated:", len(validatedProxies))
	scraper.Proxies = make([]string, 0)
	for _, proxy := range validatedProxies {
		scraper.Proxies = append(scraper.Proxies, "https://"+proxy)
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
		scraper.Queue, _ = queue.New(
			runtime.NumCPU()*10,
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
	c := colly.NewCollector(
		colly.MaxDepth(2),
	)
	c.DisableCookies()
	c.Limit(&colly.LimitRule{
		RandomDelay: 5 * time.Second,
	})
	extensions.RandomUserAgent(c)

	//set headers
	var headers map[string]string
	err := json.Unmarshal([]byte(scraper.Rule.Headers), &headers)
	if err == nil {
		scraper.Headers = headers
	} else {
		scraper.Headers = nil
	}

	// //set deep link
	// linkPatterns := strings.Split(scraper.Rule.DeepLinkPatterns, ",")
	// if len(linkPatterns) >= 3 {
	// 	//link pattern needs at least 3 parameters
	// 	c.OnHTML(linkPatterns[0], func(e *colly.HTMLElement) {
	// 		e.DOM.Each(func(index int, elem *goquery.Selection) {
	// 			parentValue := elem.Find(linkPatterns[1]).First().Text()
	// 			link, _ := elem.Find(linkPatterns[2]).First().Attr("href")
	// 			log.Println(link)
	// 			scraper.ParentLinks[link] = parentValue
	// 			e.Request.Visit(link)
	// 		})
	// 	})
	// }

	c.OnRequest(func(r *colly.Request) {
		if scraper.Headers != nil {
			for k, v := range scraper.Headers {
				r.Headers.Set(k, v)
			}
		}
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		requestLink := e.Request.URL.RequestURI()
		linkPatterns := strings.Split(scraper.Rule.DeepLinkPatterns, ",")
		if len(linkPatterns) >= 3 {
			// this is the parent page deeplink,link pattern needs at least 3 parameters
			//TODO, parentPageDOM is empty
			parentPageDom := e.DOM.Find(linkPatterns[0])
			if parentPageDom.Size() > 0 {
				parentPageDom.Each(func(index int, elem *goquery.Selection) {
					parentValue := elem.Find(linkPatterns[1]).First().Text()
					link, _ := elem.Find(linkPatterns[2]).First().Attr("href")
					log.Println(link)
					scraper.ParentLinks[link] = parentValue
					e.Request.Visit(link)
				})
				return
			}
		}
		var pattern map[string]interface{}
		err := json.Unmarshal([]byte(scraper.Rule.Pattern), &pattern)
		if !utility.IsNil(err) {
			log.Println("Cannot read the rule pattern", err)
			return
		}
		value := parsePattern(e.DOM, pattern, scraper.ParentLinks[requestLink])
		value["link"] = requestLink
		for _, v := range value {
			//check if the first level map contains required data
			invalid := false
			switch v.(type) {
			case string:
				if len(v.(string)) <= 0 {
					invalid = true
				}
			case map[string]interface{}:
				if len(v.(map[string]interface{})) <= 0 {
					invalid = true
				}
			default:
				invalid = true
			}
			if invalid {
				log.Println("Invalid Page")
				e.Request.Retry()
				return
			}
		}
		jsonString, err := json.Marshal(value)
		scraper.ScrapedRecords = append(scraper.ScrapedRecords, string(jsonString))
		log.Print("Completed")
	})

	c.OnError(func(r *colly.Response, err error) {
		// log.Println("Request URL:", r.Request.URL, "failed with response:", string(r.Body), "\nError:", err)
		log.Println("Failed HTTP")
		r.Request.Retry()
	})

	rp, err := proxy.RoundRobinProxySwitcher(scraper.Proxies...)
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	//create scraper collector
	scraper.Collector = c
	return nil
}

func parsePattern(s *goquery.Selection, item map[string]interface{}, parentValue string) map[string]interface{} {
	result := make(map[string]interface{})
	//flag to detect if it is a name or it is attribute object. If it contains pattern, value or children, then this is a attribute object rather than just a name
	nameFlag := true
	var dom *goquery.Selection

	//set dom
	if pattern, ok := item["pattern"]; ok {
		dom = s.Find(pattern.(string))
		nameFlag = false
	}

	//obtain value
	if val, ok := item["value"]; ok && val != "" {
		var value string
		if val.(string) == "text" {
			value = strings.TrimSpace(dom.First().Text())
		} else if attrs := strings.Split(val.(string), ":"); attrs[0] == "attr" {
			value, _ = dom.First().Attr(attrs[1])
			value = strings.TrimSpace(value)
		}

		//post process
		if postProcessJSON, ok := item["postprocess"]; ok {
			postProcess := postProcessJSON.(map[string]interface{})
			for k, v := range postProcess {
				switch strings.ToLower(k) {
				case "split":
					//it takes two parameters, first one is the char for split, the second one is the position of the split it wants to take
					parameters := strings.Split(v.(string), ",")
					if len(parameters) >= 2 {
						//split needs at least two parameters
						index, err := strconv.Atoi(parameters[1])
						if err == nil {
							results := strings.Split(value, parameters[0])
							if len(results) > index {
								value = results[index]
							}
						}
					}
				case "replace":
					//it takes two parameters, [0] is the old string and [1] is the new string
					parameters := strings.Split(v.(string), ",")
					if len(parameters) >= 2 {
						value = strings.ReplaceAll(value, parameters[0], parameters[1])
					}
				default:
				}
			}
		}

		//validation
		if validationJSON, ok := item["validation"]; ok && value != "" {
			validation := validationJSON.(map[string]interface{})
			if equation, ok := validation["equation"]; ok {
				if targetValue, ok := validation["targetValue"]; ok {
					expression := strings.Replace(strings.Replace(equation.(string), "parentValue", parentValue, -1), "value", value, -1)
					fs := token.NewFileSet()
					isValid, err := types.Eval(fs, nil, token.NoPos, expression)
					if err == nil {
						if isValid.Value.String() == "false" {
							//set the value to empty (means it is invalid)
							value = ""
						} else if isValid.Value.String() == "true" {
							//currently only support parentvalue or this item value
							if targetValue.(string) == "parentValue" {
								value = parentValue
							}
						}
					}
				}
			}
		}
		if !utility.IsNil(value) {
			//only set nameFlag to False if the value is valid
			result["value"] = value
			nameFlag = false
		}
	}

	//obtain children
	if children, ok := item["children"]; ok {
		dom.Each(func(index int, elem *goquery.Selection) {
			key := strconv.Itoa(index)
			result[key] = parsePattern(elem, children.(map[string]interface{}), parentValue)
		})
		nameFlag = false
	}

	//iterate all keys in the same level
	if nameFlag {
		for key, value := range item {
			result[key] = parsePattern(s, value.(map[string]interface{}), parentValue)
		}
	}

	return result
}

func proxyCheck(proxies []string) (validatedProxies []string) {
	c := make(chan string)
	for _, prox := range proxies {
		go func(prox string) {
			conn, err := net.DialTimeout("tcp", prox, 10*time.Second)
			if err == nil {
				defer conn.Close()
				c <- prox
			} else {
				log.Println(err)
				c <- ""
			}
		}(prox)
	}

	for i := 0; i < len(proxies); i++ {
		res := <-c
		if res != "" {
			validatedProxies = append(validatedProxies, res)
		}
	}
	return validatedProxies
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
