package scraper

import (
	"bytes"
	"encoding/json"
	"errors"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/gocolly/colly/queue"
	"github.com/google/uuid"

	"github.com/sporule/grater/models"
	"github.com/sporule/grater/modules/utility"
)

//scraper is the struct for scraper
type scraper struct {
	ID               string `json:"id,omitempty"`
	Collector        *colly.Collector
	Proxies          []string
	Rule             models.Rule
	Queue            *queue.Queue
	ReceviedLinkIDs  []string
	ScrapedRecords   []string
	PageLayoutErrors []string
	TableName        string
	ParentLinks      map[string]string
	ParentLinksMutex sync.RWMutex
	Headers          map[string]string
	Cookie           string
	CookiesJar       []string
	UseProxy         bool
}

func (scraper *scraper) UpdateParentLinks(link, value string) {
	scraper.ParentLinksMutex.Lock()
	scraper.ParentLinks[link] = value
	scraper.ParentLinksMutex.Unlock()
}

//new creates new scraper
func new() (*scraper, error) {
	id, _ := uuid.NewRandom()
	return &scraper{
		ID:          id.String(),
		ParentLinks: make(map[string]string),
		UseProxy:    true,
	}, nil
}

func (scraper *scraper) SaveScrapedRecords() error {
	err := models.InsertManyResults(scraper.TableName, scraper.ScrapedRecords)
	if err != nil {
		return err
	}
	scraper.ScrapedRecords = make([]string, 0)
	err = models.InsertManyResults("PageLayoutError", scraper.PageLayoutErrors)
	if err != nil {
		return err
	}
	scraper.PageLayoutErrors = make([]string, 0)
	return nil
}

func (scraper *scraper) addCookiesToJar(cookies ...string) {
	if len(scraper.CookiesJar) >= 30 {
		//maintaining maximum 30 cookies in jar
		scraper.CookiesJar = scraper.CookiesJar[len(cookies):]
	}
	scraper.CookiesJar = append(scraper.CookiesJar, cookies...)
}

func (scraper *scraper) getCookie() string {
	if len(scraper.CookiesJar) > 0 {
		randomIndex := rand.Intn(len(scraper.CookiesJar))
		return scraper.CookiesJar[randomIndex]
	}
	return ""
}

func (scraper *scraper) setProxies() error {
	scraper.Proxies = make([]string, 0)
	proxyAPIStr := utility.GetEnv("PROXY_API", "socks5%https://api.proxyscrape.com/v2/?request=getproxies&protocol=socks5&timeout=10000&country=all")
	proxyProtocol := strings.Split(proxyAPIStr, "%")[0]
	proxyLink := strings.Split(proxyAPIStr, "%")[1]
	testLink := strings.Join(strings.Split(scraper.Rule.LinkPattern, "/")[:3], "/")
	if utility.IsNil(proxyProtocol, proxyLink, testLink) {
		return errors.New("Can't read proxy configuration, disabling proxy")
	}
	proxies, err := getProxies(proxyLink, proxyProtocol, testLink)
	if err != nil {
		return err
	}
	scraper.Proxies = append(scraper.Proxies, proxies...)
	return nil
}

func getProxies(link, proxyType, testLink string) (fullProxies []string, err error) {
	res, err := http.Get(link)
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Unable to obtain Proxy")
		return nil, err
	}
	proxies := strings.Split(string(body), "\r\n")
	validatedProxies := proxyCheck(proxies, proxyType, testLink)
	log.Println("Proxies:", len(proxies), "Validated:", len(validatedProxies))
	for _, proxy := range validatedProxies {
		fullProxies = append(fullProxies, proxyType+"://"+proxy)
	}
	return fullProxies, nil
}

func (scraper *scraper) setLinksToComplete() error {
	if api := utility.GetEnv("DISTRIBUTOR_API", "http://localhost:9999/api/v1/dist"); !utility.IsNil(api) {
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
	if api := utility.GetEnv("DISTRIBUTOR_API", "http://localhost:9999/api/v1/dist"); !utility.IsNil(api) {
		//obtain the highest priority queue
		res, err := http.Get(api + "/rules")
		if err != nil {
			log.Println("Unable to make request to obtain rules", err)
			return err
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("Unable to read rules", err)
			return err
		}
		var rules []models.Rule
		err = json.Unmarshal(body, &rules)
		if err != nil {
			log.Println("Unable to parse the returned queue result", err)
			return err
		}
		if len(rules) <= 0 {
			log.Println("Unable to find any rules")
			return errors.New("Unable to find any rules")
		}
		rule := rules[0]
		if utility.IsNil(rule.ID, rule.Pattern, rule.TargetLocation) {
			log.Println("Unable to read rule information")
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
	if api := utility.GetEnv("DISTRIBUTOR_API", "http://localhost:9999/api/v1/dist"); !utility.IsNil(api) {
		//obtain the links
		res, err := http.Get(api + "/links?ruleid=" + scraper.Rule.ID + "&scraper=" + scraper.ID)
		if err != nil {
			log.Println("Unable to make request to obtain links ", err)
			return err
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("Unable to read links ", err)
			return err
		}
		var links []models.Link
		err = json.Unmarshal(body, &links)
		if err != nil {
			log.Println("Could not find links")
			return err
		}
		threadSizeStr := utility.GetEnv("THREADS", "3")
		threadSize, err := strconv.Atoi(threadSizeStr)
		if err != nil {
			threadSize = 3
		}
		scraper.Queue, _ = queue.New(
			threadSize,
			&queue.InMemoryQueueStorage{MaxSize: 100000}, // Use default queue storage
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

func (scraper *scraper) AddLinkToQueue(url string) {
	scraper.Queue.AddURL(url)
}

func (scraper *scraper) setCollector() error {
	c := colly.NewCollector(
		colly.MaxDepth(2),
	)
	c.Limit(&colly.LimitRule{
		RandomDelay: 10 * time.Second,
	})
	extensions.RandomUserAgent(c)
	c.DisableCookies()
	c.IgnoreRobotsTxt = true
	c.AllowURLRevisit = true
	//set headers
	var headers map[string]string
	err := json.Unmarshal([]byte(scraper.Rule.Headers), &headers)
	if err == nil {
		scraper.Headers = headers
	} else {
		scraper.Headers = nil
	}

	c.OnRequest(func(r *colly.Request) {
		if scraper.Headers != nil {
			for k, v := range scraper.Headers {
				r.Headers.Set(k, v)
			}
		}
		if time.Now().Second()%2 == 0 || time.Now().Second()%3 == 0 {
			//only set the server cookie in a percentage
			r.Headers.Set("cookie", scraper.getCookie())
		}
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		requestLink := e.Request.URL.String()
		// cookie := scraper.Collector.Cookies(e.Request.URL.String())
		// scraper.SaveCookie(cookie)
		linkPatterns := strings.Split(scraper.Rule.DeepLinkPatterns, ",")
		if len(linkPatterns) >= 3 {
			// this is the parent page deeplink,link pattern needs at least 3 parameters
			//TODO, parentPageDOM is empty
			parentPageDom := e.DOM.Find(linkPatterns[0])
			if parentPageDom.Size() > 0 {
				parentPageDom.Each(func(index int, elem *goquery.Selection) {
					parentValue := elem.Find(linkPatterns[1]).First().Text()
					link, _ := elem.Find(linkPatterns[2]).First().Attr("href")
					if len(linkPatterns) >= 4 {
						//remove query string
						if linkPatterns[3] == "removeQueryString" {
							link = strings.Split(link, "?")[0]
						}
					}
					if len(linkPatterns) >= 5 {
						//skip the link if it contains keyword in this list
						if !strings.Contains(link, linkPatterns[4]) {
							testLink := link[:4]
							if strings.ToLower(testLink) != "http" {
								if link[:1] == "/" {
									link = e.Request.URL.Scheme + "://" + e.Request.URL.Host + link
								} else {
									link = e.Request.URL.Scheme + "://" + e.Request.URL.Host + "/" + link
								}
							}
							scraper.UpdateParentLinks(link, parentValue)
							scraper.AddLinkToQueue(link)
						} else {
							log.Println("Not visting the link:", link, "because it contains:", linkPatterns[4])
						}
					} else {
						if strings.ToLower(link[:4]) != "http" {
							if link[:1] == "/" {
								link = e.Request.URL.Scheme + "://" + e.Request.URL.Host + link
							} else {
								link = e.Request.URL.Scheme + "://" + e.Request.URL.Host + "/" + link
							}
						}
						scraper.UpdateParentLinks(link, parentValue)
						scraper.AddLinkToQueue(link)
					}
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
		value, isWrongPage, invalidPage := parsePattern(e.DOM, pattern, scraper.ParentLinks[requestLink], true)
		if isWrongPage {
			html, _ := e.DOM.Html()
			cookie := e.Request.Headers.Get("cookie")
			scraper.PageLayoutErrors = append(scraper.PageLayoutErrors, string(requestLink+"\n"+cookie+"\n"+html))
			log.Println("Page layout not as expected", requestLink)
			scraper.AddLinkToQueue(e.Request.URL.String())
			return
		}
		if !invalidPage {
			value["link"] = requestLink
			jsonString, err := json.Marshal(value)
			if err != nil {
				log.Println("Validated data is not valid json")
				return
			}
			scraper.ScrapedRecords = append(scraper.ScrapedRecords, string(jsonString))
			log.Println("Scraped Success:", value)
		} else {
			log.Println("Data recevied failed on validation", requestLink)
		}

	})

	c.OnResponse(func(r *colly.Response) {
		cookie := r.Headers.Get("set-cookie")
		if !utility.IsNil(cookie) {
			//get server cookie mannually
			scraper.addCookiesToJar(cookie)
		}

	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Failed HTTP", r.StatusCode, err, r.Request.URL)
		scraper.AddLinkToQueue(r.Request.URL.String())
		time.Sleep(time.Duration(rand.Int31n(30)) * time.Second)
	})

	for len(scraper.Proxies) <= 0 && scraper.UseProxy {
		log.Println("Waiting for the Proxy...")
		time.Sleep(5 * time.Second)
	}
	c.SetProxyFunc(scraper.ProxySwitcher)

	//create scraper collector
	scraper.Collector = c
	return nil
}

func (scraper *scraper) ProxySwitcher(pr *http.Request) (*url.URL, error) {
	for len(scraper.Proxies) <= 0 {
		time.Sleep(1 * time.Second)
	}
	proxyStr := scraper.Proxies[rand.Intn(len(scraper.Proxies))]
	proxy := &url.URL{Scheme: strings.Split(proxyStr, "://")[0], Host: strings.Split(proxyStr, "://")[1]}
	return proxy, nil
}

func parsePattern(s *goquery.Selection, item map[string]interface{}, parentValue string, isTopLevel bool) (map[string]interface{}, bool, bool) {
	result := make(map[string]interface{})
	wrongPage := false
	invalid := false
	var dom *goquery.Selection

	//set dom
	if pattern, ok := item["pattern"]; ok {
		dom = s.Find(pattern.(string))
		if dom.Size() <= 0 {
			//can't find the dom, return wrong page
			wrongPage = true
		} else {
			//obtain value
			if val, ok := item["value"]; ok && val != "" {
				var value string
				if val.(string) == "text" {
					//value from text property
					value = strings.TrimSpace(dom.First().Text())
				} else if attrs := strings.Split(val.(string), ":"); attrs[0] == "attr" {
					//value from attr property
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
							equationStr := equation.(string)
							expression := strings.Replace(strings.Replace(equationStr, "parentValue", parentValue, -1), "value", value, -1)
							fs := token.NewFileSet()
							isValid, err := types.Eval(fs, nil, token.NoPos, expression)
							if err == nil {
								if isValid.Value.String() == "true" {
									//currently only support parentvalue or this item value
									if targetValue.(string) == "parentValue" {
										value = parentValue
										invalid = false
									}
								} else {
									invalid = true
								}
							} else {
								invalid = true
							}
						}
					}
				}
				if !utility.IsNil(value) && !invalid {
					//only set nameFlag to False if the value is valid
					result["value"] = value
				} else {
					invalid = true
				}
			}

			//obtain children
			if children, ok := item["children"]; ok {
				dom.Each(func(index int, elem *goquery.Selection) {
					if !wrongPage && !invalid {
						key := strconv.Itoa(index)
						childResult, wrongPageChild, invalidChild := parsePattern(elem, children.(map[string]interface{}), parentValue, false)
						if wrongPageChild {
							//set global wrongPage to rue if its child dom is invalid
							wrongPage = wrongPageChild
						} else if !invalidChild {
							//add the value if it is valid content
							result[key] = childResult
						} else {
							invalid = invalidChild
							result = make(map[string]interface{})
						}
					}
				})
			}
		}

	} else {
		for key, value := range item {
			if !wrongPage && !invalid {
				childResult, wrongPageChild, invalidChild := parsePattern(s, value.(map[string]interface{}), parentValue, false)
				if wrongPageChild {
					//set global wrongPage to rue if its child dom is invalid
					wrongPage = wrongPageChild
				} else if !invalidChild {
					//add the value if it is valid content
					result[key] = childResult
				} else {
					invalid = invalidChild
					result = make(map[string]interface{})
				}
			}
		}
	}

	if len(result) <= 0 {
		invalid = true
	}

	return result, wrongPage, invalid
}

//proxyCheck code from https://github.com/asm-jaime/go-proxycheck
func proxyCheck(proxies []string, proxyType string, testLink string) (validatedProxies []string) {
	c := make(chan string)
	for _, prox := range proxies {
		go func(prox string) {
			proxyURL, err := url.Parse(proxyType + "://" + prox)
			client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}, Timeout: 5 * time.Second}
			res, err := client.Get(testLink)
			if err == nil && res.StatusCode <= 299 {
				c <- prox
			} else {
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
	//get new proxies every 5 minutes
	go func() {
		for {
			scraper.setProxies()
			scraper.CookiesJar = make([]string, 0)
			time.Sleep(5 * time.Minute)
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
	scraper.setCollector()
	err = scraper.setLinksQueue()
	if !utility.IsNil(err) {
		return err
	}

	for scraper.Collector == nil {
		log.Println("Waiting for collector to be ready")
		time.Sleep(10 * time.Second)
	}
	err = scraper.Queue.Run(scraper.Collector)
	log.Println(err)
	scraper.setLinksToComplete()
	scraper.SaveScrapedRecords()
	log.Println("Scraper Completed")
	return nil
}
