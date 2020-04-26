package selina

import (
	"context"
	"fmt"
	"sync"
)

type Stats struct {
}

type Pipeliner interface {
	Run(context.Context) error
	Stats() map[string]string
	Nodes() []*Node
}
type MultiError struct {
	InnerErrors map[string]error
}

func (e *MultiError) Error() string {
	return fmt.Sprintf("%v", e.InnerErrors)
}

type SimplePipeline struct {
	nodes  []*Node
	wg     sync.WaitGroup
	mtx    sync.Mutex
	errMap map[string]error
}

func (p *SimplePipeline) startNode(ctx context.Context, n *Node) {
	if err := n.Start(ctx); err != nil {
		p.notify(n.Name, err) //handle this
	}
	p.wg.Done()
}

func (p *SimplePipeline) Run(ctx context.Context) error {
	p.wg.Add(len(p.nodes)) //1-Somethimes call Add from a goroutine is a panic
	for _, n := range p.nodes {
		go p.startNode(ctx, n)
	}
	p.wg.Wait() //2- this is called after goroutine call Add
	if len(p.errMap) > 0 {
		return &MultiError{InnerErrors: p.errMap}
	}
	return nil
}

func (p *SimplePipeline) Stats() map[string]string {
	return nil
}

func (p *SimplePipeline) Nodes() []*Node {
	return p.nodes
}
func (p *SimplePipeline) notify(name string, err error) {
	p.mtx.Lock()
	p.errMap[name] = err
	p.mtx.Unlock()
}

//Create a linear pipeline
func NewSimplePipeline(n ...*Node) Pipeliner {
	p := &SimplePipeline{}
	for i := 1; i < len(n); i++ {
		n[i-1].Chain(n[i])
	}
	p.nodes = n
	p.errMap = make(map[string]error)
	return p
}
