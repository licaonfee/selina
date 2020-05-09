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
		rm coverage.out