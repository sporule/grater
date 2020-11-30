
# Queues

## Get

GET http://localhost:8080/api/v1/queues/ HTTP/1.1

## Add Queue

POST http://localhost:8080/api/v1/queues/ HTTP/1.1
content-type: application/json

{
    "Name":"Test Queue",
    "ID": "1",
    "Status":"Active",
    "Pattern":"{\"pattern\":\"h2.entry-title\",\"children\":{\"title\":{\"pattern\":\"a[rel=bookmark]\",\"value\":\"text\"}}}",
	"Priority":5,
	"TargetLocation":"target"
}

# Messages

## Add Messages

POST http://localhost:8080/api/v1/queues/1/messages
content-type: application/json

[
    {
        "Link":"https://www.bloggingbasics101.com/"
    },
        {
        "Link":"https://www.bloggingbasics101.com/"
    },
        {
        "Link":"https://www.bloggingbasics101.com/"
    },
        {
        "Link":"https://www.bloggingbasics101.com/"
    }
]


## Request Messages

GET http://localhost:8080/api/v1/queues/1/messages/request?worker=worker1&size=2 HTTP/1.1



# Examples

POST https://example.com/comments HTTP/1.1
content-type: application/json
User-Agent: rest-client
Accept-Language: en-GB,en-US;q=0.8,en;q=0.6,zh-CN;q=0.4
Content-Type: application/json
