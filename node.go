package selina

import (
	"context"
)

//Node a node that can send and receive data
type Node struct {
	Name   string
	output Broadcaster
	input  Receiver
	w      Worker
	close  chan struct{}
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

func SafeClose(c chan<- []byte) {
	defer func() {
		_ = recover()
	}()
	close(c)
}

//Start initialize the worker, worker.Process should be called multiple times until Node is stoped
//or worker.Process return an error
func (n *Node) Start() error {
	inChan := n.input.Receive()
	outChan := make(chan []byte)
	go n.output.Broadcast(outChan)
	defer SafeClose(outChan)
	ctx := newNodeContext(context.Background(), n.close)
	if err := n.w.Process(ctx, inChan, outChan); err != nil {
		return err
	}
	return nil
}

//Stop stop worker in node, must be called after Start
func (n *Node) Stop() error {
	close(n.close)
	return nil
}

//NewNode create a new node that wraps Worker
func NewNode(name string, w Worker) *Node {
	n := &Node{w: w, Name: name}
	n.close = make(chan struct{})
	return n
}
