package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/licaonfee/selina"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

var availableNodes = map[string]NewFacility{
	"read_file":  NewReadFile,
	"write_file": NewWriteFile,
	"sql_query":  NewSQLQuery,
	"sql_insert": NewSQLInsert,
	"regex":      NewRegexp,
	"csv":        NewCSV,
	"cron":       NewCron,
}

type PipeDefinition struct {
	NodeDefs []GeneralOptions `yaml:"nodes"`
	nodes    []*selina.Node   `yaml:"-"`
}

func layout(def *PipeDefinition) (selina.Pipeliner, error) {
	nodes := make(map[string]*selina.Node, len(def.nodes))
	var usefetch bool
	for i := 0; i < len(def.nodes); i++ {
		nodes[def.nodes[i].Name()] = def.nodes[i]
		if len(def.NodeDefs[i].Fetch) > 0 {
			usefetch = true
		}
	}
	if !usefetch {
		return selina.LinealPipeline(def.nodes...), nil
	}
	chained := make(map[string]struct{})
	for _, d := range def.NodeDefs {
		me := nodes[d.Name]
		for _, f := range d.Fetch {
			prev, ok := nodes[f]
			if !ok {
				return nil, errors.New("missing node")
			}
			prev.Chain(me)
			chained[me.Name()] = struct{}{}
			chained[prev.Name()] = struct{}{}
		}
	}
	for _, d := range def.NodeDefs {
		_, ok := chained[d.Name]
		if !ok {
			return nil, errors.New("not chained")
		}
	}

	return selina.FreePipeline(def.nodes...), nil
}

func loadDefinition(definition io.Reader) (*PipeDefinition, error) {
	dec := yaml.NewDecoder(definition)
	dec.SetStrict(true)
	var defined PipeDefinition
	if err := dec.Decode(&defined); err != nil {
		return nil, err
	}

	pipeNodes := make([]*selina.Node, 0, len(defined.NodeDefs))
	for _, n := range defined.NodeDefs {
		facFunc, ok := availableNodes[n.Type]
		if !ok {
			return nil, errors.New("unavaliable type")
		}
		facility := facFunc()
		if err := mapstructure.Decode(n.Args, &facility); err != nil {
			return nil, err
		}
		node, err := facility.Make(n.Name)
		if err != nil {
			return nil, err
		}
		pipeNodes = append(pipeNodes, node)
	}
	defined.nodes = pipeNodes
	return &defined, nil
}

func createPipeline(defined *PipeDefinition) (selina.Pipeliner, error) {
	p, err := layout(defined)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func main() {
	filename := flag.String("file", "", "pipeline definition file")
	timeout := flag.Duration("timeout", time.Duration(0), "maximum time to run, default limitless")
	printSchema := flag.Bool("schema", false, "print jsonschema for yaml LSP")
	flag.Parse()

	if *printSchema {
		fmt.Println(schema())
		return
	}

	if *filename == "" {
		log.Fatal("file is mandatory")
	}

	data, err := ioutil.ReadFile(*filename)
	if err != nil {
		log.Fatal(err)
	}

	def, err := loadDefinition(bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	p, err := createPipeline(def)
	if err != nil {
		log.Fatal(err)
	}
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, os.Kill)
	var ctx context.Context
	var cancel context.CancelFunc
	if *timeout == time.Duration(0) {
		ctx, cancel = context.WithCancel(context.Background())
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), *timeout)
	}
	go func() {
		<-s
		cancel()
	}()
	if err := p.Run(ctx); err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			log.Fatalf("timeout after %v", *timeout)
		}
		log.Fatal(err)
	}
}
