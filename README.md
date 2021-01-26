# Grater

Grate is designed to be a distributed scraping tool with a rich ui to achieve high performance scraping from multiple locations and nodes.


## Update

### 2021/01/23

- Alpha version of application that supports basic functionalities

## TODO List

- [ ] Basic documentation about the rules
- [ ] Logging mechanism
- [ ] API EndPoint to generate links
- [ ] Unit Testing
- [ ] Basic Admin Panel to control the rules
- [ ] Basic Docker-Compose file
- [ ] Basic Helm Chart

## Configuration

### Argument

You can pass arguments to choose mode such as `grater both`

| Name | Default | Usage                                                                                          |
| ---- | ------- | ---------------------------------------------------------------------------------------------- |
| mode | both    | It provides the running mode. Valid values are: dist:distributor, scraper:scraper or both:both |


### Environment Variable

| Name                 | Default                                                                                             | Usage                                                                                                                                                                                                          | Type        |
| -------------------- | --------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- |
| NAME                 | random UUID                                                                                         | identify the name of the host                                                                                                                                                                                  | both        |
| ENV                  | dev                                                                                                 | It will run gin in release mode if it is set to anything other than dev                                                                                                                                        | both        |
| CONNECTION_URI       | mongodb://root:example@mongo:27017/                                                                 | connection string to the database                                                                                                                                                                              | both        |
| DATABASE_NAME        | grater                                                                                              | name of the database                                                                                                                                                                                           | both        |
| PORT                 | 9999                                                                                                | port of the api                                                                                                                                                                                                | distributor |
| ITEM_PER_PAGE        | 10                                                                                                  | Items will be returned per page from API, it means the scraper will get 10 links every time                                                                                                                    | distributor |
| DISTRIBUTOR_API      | http://localhost:9999/api/v1/dist                                                                   | Address for the distributor                                                                                                                                                                                    | scraper     |
| PROXY_API            | socks5%https://api.proxyscrape.com/v2/?request=getproxies&protocol=socks5&timeout=10000&country=all | It should be in the format  `http/tcp%<Link>`, for example `http%www.api.com`. The api should return a list of proxies in the format of ip:port. You can leave this empty and it will not use proxy by default | scraper     |
| THREADS              | 20                                                                                                  | The size of threads for signle scraper                                                                                                                                                                         | scraper     |
| SCRAPERS             | 5                                                                                                   | The number of scrapers in one node. With default setting, the total threads per node will be 5 * 20 = 100. It means 100 requests will be running in parallel.                                                                                                                                                                             | scraper     |
| ISCOOLDOWN           |                                                                                                     | It will have a random cool down time if this variable is not empty.                                                                                                                                            | scraper     |
| WRITEPAGELAYOUTERROR |                                                                                                     | It will write the page layout error to a table call `PageLayoutError` if this value is not empty                                                                                                               | scraper     |



## Rules

You can find the json payload in example folder.

Example:

```json
{
    "linkPattern": "https://www.ebay.co.uk/sch/i.html?_from=R40&_nkw=ps5&_sacat=0&LH_Auction=1&_sop=1&_pgn={page}",
    "name": "eBay PS5 Auction",
    "pattern": "{\"name\":{\"pattern\":\"h1.it-ttl\",\"value\":\"text\"},\"price\":{\"pattern\":\"div.val.vi-price span.notranslate\",\"value\":\"text\",\"postprocess\":{\"replace\":\"£,\"},\"validation\":{\"equation\":\"300 <= value\",\"targetValue\":\"value\"}}}",
    "deeplinkPatterns": "li.s-item.s-item--watch-at-corner,span.s-item__bids.s-item__bidCount,a.s-item__link,removeQueryString,redirect",
    "targetLocation": "PS5",
    "headers": "{\"accept-encoding\":\"gzip, deflate, br\",\"accept-language\":\"en-US,en;q=0.9\",\"referer\":\"https://www.ebay.co.uk/\"}",
    "totalPages": 5,
    "frequency": 86400
}
```

### linkPattern

Link Pattern currently only supports page variable {page}. This will be used to generate the actually links. It is used with totalPages. If totalPages is 5, it will generate 5 links by replacing {page} with 1,2,3,4,5.

### name

This is a simple metadata

### pattern

It is in jsonstring format.

### deeplinkPatterns

Sometimes you may want to go to the second level rather than staying in the first level.  This is the option you want to set

### headers

This is the headers will be attached to the request

### targetLocation

This is the target database name that will be used to stored the scraped data

### frequency

This sets how many seconds will this rule regenerate all links