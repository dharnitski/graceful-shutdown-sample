FROM golang:1.15

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8080

# RUN go build -o app

# CMD ["go","run","main.go"]

CMD ["app"]