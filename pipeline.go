package selina

import (
	"context"

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

//NewSimplePipeline create a linear pipeline
func NewSimplePipeline(n ...*Node) Pipeliner {
	p := &SimplePipeline{}
	p.nodes = make(map[string]*Node)
	for i := 1; i < len(n); i++ {
		prev := n[i-1]
		curr := n[i]
		prev.Chain(curr)
		p.nodes[prev.ID()] = prev
		p.nodes[curr.ID()] = curr
	}
	return p
}
