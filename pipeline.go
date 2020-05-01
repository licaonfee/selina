package selina

import (
	"context"
	"fmt"
	"sync"
)

//Pipeliner all implementations must meet the following conditions
// Run must call Node.Start of all Nodes
// Context passed in Run must be propagated to all Node.Start methods
// Nodes() return an slice with all instances of *Node
type Pipeliner interface {
	Run(context.Context) error
	Stats() map[string]Stats
	Nodes() []*Node
}

//MultiError and error that contains all pipeline's Node.Start error
type MultiError struct {
	InnerErrors map[string]error
}

//Error to implement "error" interface
func (e *MultiError) Error() string {
	return fmt.Sprintf("%v", e.InnerErrors)
}

//SimplePipeline default value is unusable, you must create it with NewSimplePipeline
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

//Run init pipeline proccesing, return an error!= nil if any Node fail
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

//Stats returns a map with all nodes Stats object
func (p *SimplePipeline) Stats() map[string]Stats {
	ret := make(map[string]Stats)
	for _, n := range p.nodes {
		ret[n.Name] = n.Stats()
	}
	return ret
}

//Nodes return all instances of *Node
func (p *SimplePipeline) Nodes() []*Node {
	return p.nodes
}
func (p *SimplePipeline) notify(name string, err error) {
	p.mtx.Lock()
	p.errMap[name] = err
	p.mtx.Unlock()
}

//NewSimplePipeline create a linear pipeline
func NewSimplePipeline(n ...*Node) Pipeliner {
	p := &SimplePipeline{}
	for i := 1; i < len(n); i++ {
		n[i-1].Chain(n[i])
	}
	p.nodes = n
	p.errMap = make(map[string]error)
	return p
}
