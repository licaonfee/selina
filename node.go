package selina

import (
	"context"
	"errors"
	"sync"
)

//Node a node that can send and receive data
type Node struct {
	Name    string
	output  Broadcaster
	input   Receiver
	w       Worker
	close   chan struct{}
	running bool
	opMx    sync.RWMutex
}

//Running true if Start() method was called
func (n *Node) Running() bool {
	n.opMx.RLock()
	defer n.opMx.RUnlock()
	return n.running
}

//Chain send messages emitted by worker to next node, it returns next node to be chained again
func (n *Node) Chain(next *Node) *Node {
	c := n.output.Client()
	next.input.Watch(c)
	return next
}

type nodeContext struct {
	context.Context
}

func newNodeContext(ctx context.Context, close <-chan struct{}) context.Context {
	parent, cancel := context.WithCancel(ctx)
	me := nodeContext{Context: parent}
	go func() {
		<-close
		cancel()
	}()
	return me
}

func safeClose(c chan<- []byte) {
	defer func() {
		_ = recover()
	}()
	close(c)
}

//ErrAlreadyStarted returned if Start method is called more than once
var ErrAlreadyStarted = errors.New("node already started")

func (n *Node) checkStart() error {
	n.opMx.Lock()
	defer n.opMx.Unlock()
	if n.running {
		return ErrAlreadyStarted
	}
	n.running = true
	return nil
}

//Start initialize the worker, worker.Process should be called multiple times until Node is stoped
//or worker.Process return an error
func (n *Node) Start(ctx context.Context) error {
	if err := n.checkStart(); err != nil {
		return err
	}
	inChan := n.input.Receive()
	outChan := make(chan []byte)
	go n.output.Broadcast(outChan)
	defer safeClose(outChan)
	inCtx := newNodeContext(ctx, n.close)

	return n.w.Process(inCtx, inChan, outChan)
}

//ErrStopNotStarted returned when Stop is called before Start method
var ErrStopNotStarted = errors.New("stopping a not started worker")

//Stop stop worker in node, must be called after Start
func (n *Node) Stop() error {
	n.opMx.RLock()
	defer n.opMx.RUnlock()
	if n.running {
		close(n.close)
		return nil
	}
	return ErrStopNotStarted
}

//NewNode create a new node that wraps Worker
func NewNode(name string, w Worker) *Node {
	n := &Node{w: w, Name: name}
	n.close = make(chan struct{})
	return n
}
