FROM golang:latest

RUN mkdir /go/src/server
ADD main.go /go/src/server/main.go

CMD ["go", "run", "/go/src/server/main.go"]