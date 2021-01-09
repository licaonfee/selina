package selina_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/licaonfee/selina"
)

func startNode(ctx context.Context, n *selina.Node, t *testing.T) func() error {
	return func() error {
		t.Logf("Start %s", n.Name())
		return n.Start(ctx)
	}
}

func TestNodeChainPassMessages(t *testing.T) {
	r := &sliceReader{values: []string{"1", "2", "3"}}
	w := &sliceWriter{}
	n1 := selina.NewNode("A", r)
	n2 := selina.NewNode("B", &dummyWorker{})
	n3 := selina.NewNode("C", &dummyWorker{})
	n4 := selina.NewNode("D", w)
	n1.Chain(n2).
		Chain(n3).
		Chain(n4)

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(startNode(ctx, n1, t))
	g.Go(startNode(ctx, n2, t))
	g.Go(startNode(ctx, n3, t))
	g.Go(startNode(ctx, n4, t))
	if err := g.Wait(); err != nil {
		t.Fatalf("Chain() err = %v", err)
	}
	if !reflect.DeepEqual(r.values, w.values) {
		t.Fatalf("Chain() values differ got %v , want %v", w.values, r.values)
	}
}

func TestNodeChainPassMessagesNoDoubleChain(t *testing.T) {
	r := &sliceReader{values: []string{"1", "2", "3"}}
	w := &sliceWriter{}
	n1 := selina.NewNode("A", r)
	n2 := selina.NewNode("B", &dummyWorker{})
	n3 := selina.NewNode("C", &dummyWorker{})
	n4 := selina.NewNode("D", w)
	n1.Chain(n2).
		Chain(n3).
		Chain(n4)
	n1.Chain(n2).
		Chain(n3).
		Chain(n4)
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(startNode(ctx, n1, t))
	g.Go(startNode(ctx, n2, t))
	g.Go(startNode(ctx, n3, t))
	g.Go(startNode(ctx, n4, t))
	if err := g.Wait(); err != nil {
		t.Fatalf("Chain() err = %v", err)
	}
	if !reflect.DeepEqual(r.values, w.values) {
		t.Fatalf("Chain() values differ got %v , want %v", w.values, r.values)
	}
}

func TestNodeCheckStarted(t *testing.T) {
	const waitForStart = time.Millisecond * 20
	node := selina.NewNode("Me", &dummyWorker{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = node.Start(ctx)
	}()
	time.Sleep(waitForStart)
	if err := node.Start(ctx); err != selina.ErrAlreadyStarted {
		t.Fatalf("Start() err = %v", err)
	}
}

func TestNodeStop(t *testing.T) {
	n1 := selina.NewNode("A", &lazyWorker{})
	stoped := make(chan error, 1)
	go func() {
		err := n1.Start(context.Background())
		stoped <- err
	}()
	for !n1.Running() {

	}
	if err := n1.Stop(); err != nil {
		t.Fatalf("Stop() err = %v", err)
	}
	select {
	case err := <-stoped:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Start() err = %v", err)
		}
		return
	case <-time.After(time.Second * 5):
		t.Fatalf("Stop() is not working")
	}

}

func Benchmark_Node(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		start := selina.NewNode("start", &produceN{count: 100, message: []byte("byte")})
		end := selina.NewNode("end", &sink{})
		start.Chain(end)
		b.StartTimer()
		go func() {
			_ = start.Start(context.Background())
		}()
		_ = end.Start(context.Background())
	}
}
