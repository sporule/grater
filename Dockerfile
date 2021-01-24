FROM golang:1.15.7


WORKDIR /go/src/grater
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...


EXPOSE 9999