FROM golang:1.14.4-alpine

RUN apk update && apk add --no-cache git

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download 

COPY . .

EXPOSE 8080

RUN go get github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon -log-prefix=false -build="go build ." -command="./golang-task"