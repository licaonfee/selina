# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool

all: deps utest cover

deps:
		$(GOCMD) mod download

utest:
		$(GOTEST) -race -count 1 -timeout 30s -parallel 1 -coverprofile cover.out  ./... 
cover:
		$(GOTOOL) cover -func=cover.out 

clean:
		rm cover.out