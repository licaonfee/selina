package selina

import (
	"context"
	"sync"
)

type Stats struct {
}

type Pipeliner interface {
	Run(context.Context) error
	Stats()
}

type Pipeline struct {
	nodes []*Node
	wg    sync.WaitGroup
}

func (p *Pipeline) startNode(ctx context.Context, n *Node) {
	if err := n.Start(); err != nil {
		panic(err) //handle this
	}
	p.wg.Done()
}

func (p *Pipeline) Run(ctx context.Context) error {
	p.wg.Add(len(p.nodes)) //1-Somethimes call Add from a goroutine is a panic
	for _, n := range p.nodes {
		go p.startNode(ctx, n)
	}
	p.wg.Wait() //2- this is called after goroutine call Add
	return nil
}

func (p *Pipeline) Stats() {
}

func NewPipeline(n ...*Node) Pipeliner {
	p := &Pipeline{}
	for i := 1; i < len(n); i++ {
		n[i-1].Chain(n[i])
	}
	return p
}
