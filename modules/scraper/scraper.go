package scraper

import (
	"bytes"
	"encoding/json"
	"errors"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"math"
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
	id                       string
	collector                *colly.Collector
	proxies                  []string
	rule                     models.Rule
	queue                    *queue.Queue
	receviedLinkIDs          []string
	scrapedRecords           []string
	pageLayoutErrors         []string
	parentLinks              map[string]string
	parentLinksMutex         sync.RWMutex
	headers                  map[string]string
	cookie                   string
	cookiesJar               []string
	useProxy                 bool
	profileChangedTimeStamp  time.Time
	profileChangedMutex      sync.RWMutex
	failedRequests           int
	pendingLinks             []string
	previousPendingLinksSize int
	failedTimes              int
}

func (scraper *scraper) updateParentLinks(link, value string) {
	scraper.parentLinksMutex.Lock()
	scraper.parentLinks[link] = value
	scraper.parentLinksMutex.Unlock()
}

//new creates new scraper
func new(id string) (*scraper, error) {

	return &scraper{
		id:          id,
		parentLinks: make(map[string]string),
		useProxy:    true,
	}, nil
}

func (scraper *scraper) saveScrapedRecords() error {
	//save records
	if len(scraper.scrapedRecords) <= 0 {
		return nil
	}
	err := models.InsertManyResults(scraper.rule.TargetLocation, scraper.scrapedRecords)
	if err != nil {
		return err
	}
	scraper.scrapedRecords = make([]string, 0)

	//save page layouts
	if len(scraper.pageLayoutErrors) <= 0 {
		return nil
	}
	err = models.InsertManyResults("PageLayoutError", scraper.pageLayoutErrors)
	if err != nil {
		return err
	}
	scraper.pageLayoutErrors = make([]string, 0)

	return nil
}

func (scraper *scraper) changeProfile(isProxy, isCookies bool) {
	scraper.profileChangedMutex.Lock()
	scraper.failedRequests++
	if isProxy {
		for len(scraper.proxies) <= 1 {
			log.Println("Waiting for proxy")
			time.Sleep(20 * time.Second)
		}
	}
	if scraper.profileChangedTimeStamp.Add(10*time.Second).Before(time.Now()) || scraper.failedRequests >= 15 {
		if isCookies {
			if len(scraper.cookiesJar) > 1 {
				scraper.cookie = scraper.cookiesJar[0]
				//reset cookies
				scraper.cookiesJar = scraper.cookiesJar[1:]
			}
			scraper.profileChangedTimeStamp = time.Now()
		}
		if isProxy {
			//reset proxies
			scraper.proxies = scraper.proxies[1:]
			scraper.profileChangedTimeStamp = time.Now()
		}

		log.Println("Profile Changed, proxies:", len(scraper.proxies), "cookies:", len(scraper.cookiesJar), "failed requests:", scraper.failedRequests)
		scraper.failedRequests = 0
	}
	scraper.profileChangedMutex.Unlock()
}

func (scraper *scraper) addCookiesToJar(cookies ...string) {
	scraper.cookie += cookies[0]
	//log.Println("Latest Cookies:", scraper.cookie)
}

func (scraper *scraper) getCookie() string {
	return scraper.cookie
}

func (scraper *scraper) setProxies() error {
	log.Println("Getting Proxies")
	proxyAPIStr := utility.GetEnv("PROXY_API", "socks5-grater-https://api.proxyscrape.com/v2/?request=getproxies&protocol=socks5&timeout=10000&country=all")
	proxyAPISlice := strings.Split(proxyAPIStr, "-grater-")
	if len(proxyAPISlice) <= 1 {
		return errors.New("Can't read proxy configuration, disabling proxy")
	}
	proxyProtocol := proxyAPISlice[0]
	proxyLink := proxyAPISlice[1]
	testLink := strings.Join(strings.Split(scraper.rule.LinkPattern, "/")[:3], "/")
	if utility.IsNil(proxyProtocol, proxyLink, testLink) {
		return errors.New("Can't read proxy configuration, disabling proxy")
	}
	proxies, cookies, err := getProxies(proxyLink, proxyProtocol, testLink)
	if err != nil {
		return err
	}
	scraper.proxies = proxies
	scraper.cookiesJar = cookies
	log.Println("Proxy obtained, size:", len(proxies), "cookies, size:", len(cookies))
	return nil
}

func getProxies(link, proxyType, testLink string) (fullProxies []string, cookies []string, err error) {
	for len(fullProxies) <= 1 {
		res, err := http.Get(link)
		if res.Body != nil {
			defer res.Body.Close()
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("Unable to obtain Proxy")
			return nil, nil, err
		}
		proxies := strings.Split(string(body), "\r\n")
		if len(proxies) <= 1 {
			proxies = strings.Split(string(body), "\n")
		}
		validatedProxies, cookies := proxyCheck(proxies, proxyType, testLink)
		log.Println("Proxies:", len(proxies), "Validated:", len(validatedProxies), "Cookies:", len(cookies))
		for _, proxy := range validatedProxies {
			fullProxies = append(fullProxies, proxyType+"://"+proxy)
		}
		if len(fullProxies) > 1 {
			return fullProxies, cookies, nil
		}
		time.Sleep(5 * time.Second)
	}
	//this code will never exectue
	return nil, nil, nil
}

func (scraper *scraper) setLinksToComplete() error {
	if api := utility.GetEnv("DISTRIBUTOR_API", "http://localhost:9999/api/v1/dist"); !utility.IsNil(api) {
		body, err := json.Marshal(map[string][]string{
			"linkids": scraper.receviedLinkIDs,
		})
		if err != nil {
			return errors.New("Error on parsing completed link IDs")
		}
		_, err = http.Post(api+"/links", "application/json", bytes.NewBuffer(body))
		if err == nil {
			//reset the completedLinkIDs
			scraper.receviedLinkIDs = make([]string, 0)
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
		res, err := http.Get(api + "/rules?isscraper=1")
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
			return errors.New("Unable to read rule information.")
		}
		scraper.rule = rule
	} else {
		return errors.New("API Not found")
	}
	return nil
}

func getLinks(ruleID string, scraperID string) (linkIDs, pendingLinks []string, err error) {
	if api := utility.GetEnv("DISTRIBUTOR_API", "http://localhost:9999/api/v1/dist"); !utility.IsNil(api) {
		//obtain the links
		res, err := http.Get(api + "/links?ruleid=" + ruleID + "&scraper=" + scraperID)
		if err != nil {
			log.Println("Unable to make request to obtain links ", err)
			return nil, nil, err
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("Unable to read links ", err)
			return nil, nil, err
		}
		var links []models.Link
		err = json.Unmarshal(body, &links)
		if err != nil {
			log.Println("Could not find links")
			return nil, nil, err
		}
		for _, link := range links {
			linkIDs = append(linkIDs, link.ID)
			pendingLinks = append(pendingLinks, link.Link)
		}
	} else {
		return nil, nil, errors.New("API Not found")
	}
	return linkIDs, pendingLinks, nil
}

func (scraper *scraper) setLinksQueue() error {
	if len(scraper.pendingLinks) > 0 {
		threadSizeStr := utility.GetEnv("THREADS", "20")
		threadSize, err := strconv.Atoi(threadSizeStr)
		if err != nil {
			threadSize = 20
		}
		scraper.queue, _ = queue.New(
			threadSize,
			&queue.InMemoryQueueStorage{MaxSize: 100000}, // Use default queue storage
		)
		for _, link := range scraper.pendingLinks {
			scraper.queue.AddURL(link)
		}
		scraper.pendingLinks = make([]string, 0)

	}
	return nil
}

func (scraper *scraper) addLinkToQueue(url string) {
	if scraper.failedTimes > 10 {
		//give up the url
		log.Println("Giving up the link:", url)
		return
	}
	scraper.pendingLinks = append(scraper.pendingLinks, url)
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
	err := json.Unmarshal([]byte(scraper.rule.Headers), &headers)
	if err == nil {
		scraper.headers = headers
	} else {
		scraper.headers = nil
	}

	c.OnRequest(func(r *colly.Request) {
		if scraper.headers != nil {
			for k, v := range scraper.headers {
				r.Headers.Set(k, v)
			}
		}
		r.Headers.Set("cookie", scraper.getCookie())
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		requestLink := e.Request.URL.String()
		linkPatterns := strings.Split(scraper.rule.DeepLinkPatterns, ",")
		if len(linkPatterns) >= 3 {
			// this is the parent page deeplink,link pattern needs at least 3 parameters
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
							scraper.updateParentLinks(link, parentValue)
							scraper.addLinkToQueue(link)
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
						scraper.updateParentLinks(link, parentValue)
						scraper.addLinkToQueue(link)
					}
				})
				return
			}
		}
		var pattern map[string]interface{}
		err := json.Unmarshal([]byte(scraper.rule.Pattern), &pattern)
		if !utility.IsNil(err) {
			log.Println("Cannot read the rule pattern", err)
			return
		}
		value, isWrongPage, invalidPage := parsePattern(e.DOM, pattern, scraper.parentLinks[requestLink], true)
		if isWrongPage {
			if len(utility.GetEnv("WRITEPAGELAYOUTERROR", "")) > 0 {
				html, _ := e.DOM.Html()
				cookie := e.Request.Headers.Get("cookie")
				scraper.pageLayoutErrors = append(scraper.pageLayoutErrors, requestLink+"*****"+cookie+"*****"+html)
			}
			//log.Println("Page layout not as expected,change cookie", requestLink)
			//change cookie and proxy
			scraper.changeProfile(true, true)
			scraper.addLinkToQueue(e.Request.URL.String())
			return
		}
		if !invalidPage {
			value["link"] = requestLink
			value["rule"] = scraper.rule.Name
			jsonString, err := json.Marshal(value)
			if err != nil {
				log.Println("Validated data is not valid json")
				return
			}
			scraper.scrapedRecords = append(scraper.scrapedRecords, string(jsonString))
			//remove item from the map
			delete(scraper.parentLinks, requestLink)
			log.Println("Scraped Success:", value)
		} else {
			log.Println("Data recevied failed on validation", requestLink)
		}

	})

	c.OnResponse(func(r *colly.Response) {
		cookie := getCookieFromRespList(r.Headers.Values("set-cookie"))
		if !utility.IsNil(cookie) {
			//get server cookie mannually
			scraper.addCookiesToJar(cookie)
		}

	})

	c.OnError(func(r *colly.Response, err error) {
		//log.Println("Failed HTTP", r.StatusCode, err, r.Request.URL)
		if r.StatusCode <= 10 {
			scraper.changeProfile(true, false)
		} else {
			scraper.changeProfile(true, true)
		}
		scraper.addLinkToQueue(r.Request.URL.String())
	})

	for len(scraper.proxies) <= 0 && scraper.useProxy {
		log.Println("Waiting for the Proxy...")
		time.Sleep(20 * time.Second)
	}
	c.SetProxyFunc(scraper.proxySwitcher)

	//create scraper collector
	scraper.collector = c
	return nil
}

func (scraper *scraper) proxySwitcher(pr *http.Request) (*url.URL, error) {
	for len(scraper.proxies) <= 0 {
		log.Println("Proxy switcher is waiting for proxy, sleep for 5 seconds")
		time.Sleep(5 * time.Second)
	}
	proxyStr := scraper.proxies[0] //always return the first proxy
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
func proxyCheck(proxies []string, proxyType string, testLink string) (validatedProxies []string, cookies []string) {
	c := make(chan string)
	timeout := math.Max(float64(len(proxies))*0.01, 10.0)
	log.Println("Validating Proxies, it could take:", timeout, "seconds")
	for _, prox := range proxies {
		go func(prox string) {
			proxyURL, err := url.Parse(proxyType + "://" + prox)
			client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}, Timeout: time.Duration(timeout) * time.Second}
			req, err := http.NewRequest("GET", testLink, nil)
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.96 Safari/537.36 Edg/88.0.705.50")
			req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
			req.Header.Set("accept-encoding", "gzip, deflate, br")
			req.Header.Set("accept-language", "en-GB,en;q=0.9,en-US;q=0.8,zh-CN;q=0.7,zh-TW;q=0.6,zh;q=0.5")
			req.Header.Set("cache-control", "max-age=100")
			res, err := client.Do(req)
			if err == nil && res.StatusCode <= 299 {
				cookie := getCookieFromRespList(res.Header.Values("set-cookie"))
				if len(cookie) > 0 {
					cookies = append(cookies, cookie)
				}
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
	return validatedProxies, cookies
}

func getCookieFromRespList(respCookies []string) string {
	cookie := ""
	for _, v := range respCookies {
		cookie += strings.Split(v, ";")[0] + "; "
	}
	return cookie
}

//runOneScraper fires of the scraping process
func runOneScraper(id string) error {
	log.Println("Scraper started")
	flag := true
	scraper, _ := new(id)
	//Get Rule
	err := scraper.setRule()
	//Get Links from Rule
	linkIDs, pendingLinks, err := getLinks(scraper.rule.ID, scraper.id)
	if !utility.IsNil(err) {
		return err
	}
	scraper.receviedLinkIDs = linkIDs
	scraper.pendingLinks = pendingLinks
	//get new proxies periodically
	go func() {
		for flag {
			scraper.setProxies()
			time.Sleep(6 * time.Minute)
		}
	}()
	//save data to database very minute
	go func() {
		for flag {
			scraper.saveScrapedRecords()
			time.Sleep(30 * time.Second)
		}
	}()
	if err != nil {
		log.Println(err)
		return err
	}
	for len(scraper.pendingLinks) > 0 {
		if len(scraper.pendingLinks) == scraper.previousPendingLinksSize {
			scraper.failedTimes++
		} else {
			scraper.failedTimes = 0
		}
		scraper.previousPendingLinksSize = len(scraper.pendingLinks)
		coolDownDelay := rand.Int31n(int32(math.Max(float64(len(scraper.pendingLinks)), 60)))
		if utility.GetEnv("ISCOOLDOWN", "") == "" {
			//set cooldown to 5 if it is disabled
			coolDownDelay = 10
		}
		log.Println("Refreshing collector,queue,proxies and cookies,sleep for ", coolDownDelay, "seconds. Size of Links:", len(scraper.pendingLinks), "Total Failed Time:", scraper.failedTimes)
		time.Sleep(time.Duration(coolDownDelay) * time.Second)
		log.Println("Setting collector")
		scraper.setCollector()
		log.Println("Setting links queue")
		err = scraper.setLinksQueue()
		if !utility.IsNil(err) {
			return err
		}

		for scraper.collector == nil {
			log.Println("Waiting for collector to be ready")
			time.Sleep(10 * time.Second)
		}
		log.Println("SStart running the collector")
		scraper.queue.Run(scraper.collector)
	}
	err = scraper.setLinksToComplete()
	if err != nil {
		log.Println("Set Links to Complete failed:", err)
	}
	log.Println("Link set to completed")
	scraper.saveScrapedRecords()
	flag = false //removing those loops
	log.Println("Scraper Completed")
	return nil
}

//StartScraping fires of the scraping process
func StartScraping() (err error) {
	scrapersStr := utility.GetEnv("SCRAPERS", "3")
	scrapers, err := strconv.Atoi(scrapersStr)
	if err != nil {
		scrapers = 3
	}
	id, _ := uuid.NewRandom()
	name := utility.GetEnv("NAME", id.String())
	errs := make(chan error)
	for i := 1; i <= scrapers; i++ {
		go func() {
			errs <- runOneScraper(name)
		}()
		go func() {
			time.Sleep(15 * time.Minute)
			errs <- nil
		}()
		//random delay before running the next scraper
		time.Sleep(time.Duration(rand.Intn(60)) * time.Second)
	}

	for i := 1; i <= scrapers; i++ {
		tempErr := <-errs
		if tempErr == nil {
			err = nil
		}
	}
	errs = nil
	return err
}
