# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool

all: deps utest cover

deps:
		$(GOCMD) mod download

utest:
		$(GOTEST) -race -count 1 -timeout 30s  -coverprofile coverage.out  ./... 
cover:
		$(GOTOOL) cover -func=coverage.out 

clean:
		rm -f ./coverage.out
		rm -rf ./bin 
		rm -f schema.json

build:
		mkdir -p bin
		CGO_ENABLED=0 $(GOBUILD) -o bin/selina cmd/*.go 

schema:
	go run cmd/*.go -schema > schema.json
