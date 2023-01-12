package selina

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

// Stats contain node overall statistics
// Counters are garanted to be consistent only when node finalize
type Stats struct {
	Time          time.Time
	Sent          int64
	SentBytes     int64
	Received      int64
	ReceivedBytes int64
}

// Node a node that can send and receive data
type Node struct {
	id      string
	name    string
	output  Broadcaster
	input   Receiver
	w       Worker
	close   chan struct{}
	running bool
	opMx    sync.RWMutex
	chained map[string]struct{}
}

// ID return a unique identifier for this node
func (n *Node) ID() string {
	return n.id
}

// Name return node name this value is not unique
func (n *Node) Name() string {
	return n.name
}

// Running true if Start() method was called
func (n *Node) Running() bool {
	n.opMx.RLock()
	defer n.opMx.RUnlock()
	return n.running
}

// Chain send messages emitted by worker to next node,
// it returns next node to be chained again
// if next is already chained this operation does nothing
func (n *Node) Chain(next *Node) *Node {
	if n.IsChained(next) {
		return next
	}
	c := n.output.Client()
	next.input.Watch(c)
	n.chained[next.ID()] = struct{}{}
	return next
}

// Next returns nodes id chained to current node
func (n *Node) Next() []string {
	ret := make([]string, 0, len(n.chained))
	for k := range n.chained {
		ret = append(ret, k)
	}
	return ret
}

// IsChained returns true if Chain was called before with other
func (n *Node) IsChained(other *Node) bool {
	_, ok := n.chained[other.ID()]
	return ok
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

func safeCloseChan[T any](c chan<- T) {
	defer func() {
		_ = recover()
	}()
	close(c)
}

// ErrAlreadyStarted returned if Start method is called more than once
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

// Start initialize the worker, worker.Process should be called multiple times until Node is stoped
// or worker.Process return an error
func (n *Node) Start(ctx context.Context) error {
	if err := n.checkStart(); err != nil {
		return err
	}
	inChan := n.input.Receive()
	outChan := make(chan *bytes.Buffer)
	go n.output.Broadcast(outChan)
	defer safeCloseChan(outChan)
	inCtx := newNodeContext(ctx, n.close)
	args := ProcessArgs{Input: inChan, Output: outChan}
	err := n.w.Process(inCtx, args)
	if err != nil {
		return fmt.Errorf("%s : %w", n.name, err)
	}
	return nil
}

// ErrStopNotStarted returned when Stop is called before Start method
var ErrStopNotStarted = errors.New("stopping a not started worker")

// Stop stop worker in node, must be called after Start
// successive calls to Stop does nothing
func (n *Node) Stop() error {
	n.opMx.RLock()
	defer n.opMx.RUnlock()
	if n.running {
		safeCloseChan(n.close)
		return nil
	}
	return ErrStopNotStarted
}

// Stats return Worker channels stats
func (n *Node) Stats() Stats {
	oc, ob := n.output.Stats()
	ic, ib := n.input.Stats()
	return Stats{Sent: oc, SentBytes: ob, Received: ic, ReceivedBytes: ib}
}

func getID() string {
	return ulid.Make().String()
}

// NewNode create a new node that wraps Worker
// name is a user defined identifier, internally
// Node generates an unique id
func NewNode(name string, w Worker) *Node {
	id := getID()
	n := &Node{id: id, w: w, name: name}
	n.chained = make(map[string]struct{})
	n.close = make(chan struct{})
	return n
}
