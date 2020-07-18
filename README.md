# selina

![Test](https://github.com/licaonfee/selina/workflows/Run%20test/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/licaonfee/selina)](https://goreportcard.com/report/github.com/licaonfee/selina)
[![Coverage Status](https://coveralls.io/repos/github/licaonfee/selina/badge.svg?branch=master)](https://coveralls.io/github/licaonfee/selina?branch=master)
[![godoc](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/licaonfee/selina?tab=doc)

Simple Pipeline for go, inspired on ratchet <https://github.com/dailyburn/ratchet>

Unstable API, please use go modules

- [selina](#selina)
  - [Installation](#installation)
  - [Usage](#usage)
  - [Builtin workers](#builtin-workers)
  - [Design](#design)
    - [Pipeline](#pipeline)
    - [Node](#node)
    - [Worker](#worker)
  - [Commandline Usage](#commandline-usage)

## Installation

```bash
go get github.com/licaonfee/selina
```

## Usage

```go
package main

import (
    "fmt"
    "os"
    "strings"
    "context"

    "github.com/licaonfee/selina"
    "github.com/licaonfee/selina/workers/regex"
    "github.com/licaonfee/selina/workers/text"
)

const sample = `this is a sample text
this is second line
#lines started with # will be skipped
this line pass`

func main() {
    rd := strings.NewReader(sample)
    input := selina.NewNode("Read", text.NewReader(text.ReaderOptions{Reader: rd}))
    //https://regex101.com/r/7ZS3Uw/1
    filter := selina.NewNode("Filter", regex.NewFilter(regex.FilterOptions{Pattern: "^[^#].+"}))
    output := selina.NewNode("Write", text.NewWriter(text.WriterOptions{Writer: os.Stdout}))
    pipe := selina.NewSimplePipeline(input, filter, output)
    if err := pipe.Run(context.Background()); err != nil {
        fmt.Printf("ERR: %v\n", err)
    }
    for name, stat := range pipe.Stats(){
        fmt.Printf("Node:%s=%v\n", name, stat)
    }
}
```

## Builtin workers

By default selina has this workers implemented

- csv.Encoder : Transform data from json to csv
- csv.Decoder : Transform csv data into json
- custom.Function : Allow to execute custom functions into a pipeline node
- ops.Cron : Allow scheduled messages into a pipeline
- random.Random : Generate random byte slices
- regex.Filter : Filter data using a regular expresion
- remote.Server : Listen for remote data
- remote.Client : Send data to a remote pipeline
- sql.Reader : Execute a query against a database and return its rows as json objects
- sql.Writer : Insert rows into a table from json objects
- text.Reader : Use any io.Reader and read its contents as text
- text.Writer : Write text data into any io.Writer

## Design

Selina have three main components

- Pipeline
- Node
- Worker

Some utility functions are provided to build pipelines, ```LinealPipeline(n ... Node)*Pipeliner``` chain all nodes in same order as their are passed. ```FreePipeline(n ...Node)*Pipeliner``` Just runs all nodes without chain them so you can build any pipeline, including ciclic graphs or aciclic graphs

### Pipeline

Start data processing and manage all chained nodes in a single object

### Node

Contains methods to pass data from Worker to Worker and get metrics

### Worker

All data Extraction/Transformation/Load logic is encapsulated in a Worker instance

## Commandline Usage

```bash
selina -file pipeline.yml -timeout 10h
```

Where pipeline.yml is

```yaml
---
nodes:
  - name: employes
    type: read_file
    args:
      filename: sample/employes.csv
  - name: filter_it
    type: regex
    args:
      pattern: '^.*,it,.*$'
  - name: to_json
    type: csv
    args:
      mode: decode
      header: [name,role,department,id]    
  - name: it_employes
    type: write_file
    args:
      filename: it_employes.txt
      ifexists: overwrite
      mode: 0777
layout: employes
  .Chain(filter_it)
  .Chain(to_json)
  .Chain(it_employes)
```
