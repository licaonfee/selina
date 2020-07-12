package main

import (
	"context"
	"log"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/licaonfee/selina"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

var availableNodes = map[string]func() NodeFacility{
	"read_file":  func() NodeFacility { return &ReadFile{} },
	"write_file": func() NodeFacility { return &WriteFile{} },
	"sql_query":  func() NodeFacility { return &SQLQuery{} },
	"regex":      func() NodeFacility { return &Regexp{} },
	"csv":        func() NodeFacility { return &CSV{} },
}

type PipeDefinition struct {
	Nodes  []GeneralOptions `yaml:"nodes"`
	Layout []string
}

const sampleDefinition = `
#employes.csv
#name,role,department,id
#jhon,student,it,1
#josh,manager,sales,2
---
nodes:
  - name: employes
    type: read_file
    args:
      filename: employes.csv
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
      ifexists: append
      mode: 0777
#Absent or empty layout assumes lineal layout
#all nodes will be chained in same order as declared
#another way is chaining nodes manually using A.Chain(B)
#layout mut be a single or multiple singleline expresions
#example:
#layout:
#  - A.Chain(B)
#  - B.Chain(C)
#  - C.Chain(D)
layout:
  - employes.Chain(filter_it)
  - filter_it.Chain(to_json)
  - to_json.Chain(it_employes)
`

//Chainer is just to hide full selina.Node signature
type Chainer struct {
	n *selina.Node
}

func (c Chainer) Chain(next Chainer) Chainer {
	c.n.Chain(next.n)
	return next
}
func layout(layout []string, nodes []*selina.Node) selina.Pipeliner {
	env := make(map[string]Chainer)
	for _, n := range nodes {
		env[n.Name()] = Chainer{n: n}
	}
	vmEnv := expr.Env(env)
	for _, l := range layout {
		prog, err := expr.Compile(l, vmEnv)
		if err != nil {
			log.Fatal(err)
		}
		_, err = expr.Run(prog, env)
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, v := range env {
		v.n = nil
	}
	return selina.FreePipeline(nodes...)
}

func loadPipeline(definition string) {
	dec := yaml.NewDecoder(strings.NewReader(definition))
	dec.SetStrict(true)
	var defined PipeDefinition
	if err := dec.Decode(&defined); err != nil {
		log.Fatal(err)
	}
	pipeNodes := make([]*selina.Node, 0, len(defined.Nodes))
	for _, n := range defined.Nodes {
		facFunc, ok := availableNodes[n.Type]
		if !ok {
			log.Fatal("unavaliable type")
		}
		facility := facFunc()
		if err := mapstructure.Decode(n.Args, &facility); err != nil {
			log.Fatal(err)
		}
		node, err := facility.Make(n.Name)
		if err != nil {
			log.Fatal(err)
		}
		pipeNodes = append(pipeNodes, node)
	}
	var p selina.Pipeliner

	if len(defined.Layout) == 0 {
		p = selina.LinealPipeline(pipeNodes...)
	} else {
		p = layout(defined.Layout, pipeNodes)
	}
	if err := p.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func main() {

	loadPipeline(sampleDefinition)
}
