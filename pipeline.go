package selina

import (
	"context"
	"fmt"
	"io"

	"golang.org/x/sync/errgroup"
)

//Pipeliner all implementations must meet the following conditions
// Run must call Node.Start of all Nodes
// Context passed in Run must be propagated to all Node.Start methods
// Nodes() return an slice with all instances of *Nod
type Pipeliner interface {
	Run(context.Context) error
	Stats() map[string]Stats
	Nodes() []*Node
}

//SimplePipeline default value is unusable, you must create it with NewSimplePipeline
type SimplePipeline struct {
	nodes map[string]*Node
}

//Run init pipeline proccesing, return an error!= nil if any Node fail
func (p *SimplePipeline) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, n := range p.nodes {
		node := n
		g.Go(func() error {
			return node.Start(ctx)
		})
	}
	return g.Wait()
}

//Stats returns a map with all nodes Stats object
func (p *SimplePipeline) Stats() map[string]Stats {
	ret := make(map[string]Stats)
	for _, n := range p.nodes {
		ret[n.ID()] = n.Stats()
	}
	return ret
}

//Nodes return all instances of *Node
func (p *SimplePipeline) Nodes() []*Node {
	ret := make([]*Node, 0, len(p.nodes))
	for _, v := range p.nodes {
		ret = append(ret, v)
	}
	return ret
}

//LinealPipeline creates a Pipeliner
//Nodes in "nodes" are chained in a slingle branch Pipeline
//Node(0)->Node(1)->Node(2)->....Node(n)
func LinealPipeline(nodes ...*Node) Pipeliner {
	p := &SimplePipeline{}
	p.nodes = make(map[string]*Node)
	for i := 1; i < len(nodes); i++ {
		prev := nodes[i-1]
		curr := nodes[i]
		prev.Chain(curr)
		p.nodes[prev.ID()] = prev
		p.nodes[curr.ID()] = curr
	}
	return p
}

//FreePipeline provide a method to run arbitrary chained Nodes
//this method does not call Node.Chain
func FreePipeline(nodes ...*Node) Pipeliner {
	p := &SimplePipeline{}
	p.nodes = make(map[string]*Node)
	for _, n := range nodes {
		p.nodes[n.ID()] = n

	}
	return p
}

func Graph(p Pipeliner, w io.Writer) error {
	_, err := fmt.Fprintln(w, "digraph {\n\trankdir=LR;")
	if err != nil {
		return err
	}
	for _, n := range p.Nodes() {
		_, err := fmt.Fprintf(w, "\tX%s[label=\"%s\"];\n", n.ID(), n.Name())
		if err != nil {
			return err
		}
	}
	for _, n := range p.Nodes() {
		for _, id := range n.Next() {
			_, err := fmt.Fprintf(w, "\tX%s -> X%s;\n", n.ID(), id)
			if err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintln(w, "}"); err != nil {
		return err
	}
	return nil
}
