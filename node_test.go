package selina_test

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/licaonfee/selina"
)

type dummyWorker struct{}

func (d *dummyWorker) Process(ctx context.Context, in <-chan []byte, out chan<- []byte) error {
	for msg := range in {
		out <- msg
	}
	return nil
}

type sliceReader struct {
	values []string
}

func (r *sliceReader) Process(ctx context.Context, in <-chan []byte, out chan<- []byte) error {
	rd := selina.SliceAsChannel(r.values, true)
	for msg := range rd {
		out <- msg
	}
	return nil
}

type sliceWriter struct {
	values []string
}

func (w *sliceWriter) Process(ctx context.Context, in <-chan []byte, out chan<- []byte) error {
	for msg := range in {
		w.values = append(w.values, string(msg))
	}

	return nil
}

func (w *sliceWriter) Finish() error {
	return nil
}

func startNode(wg *sync.WaitGroup, n *selina.Node) {
	wg.Add(1)
	go func() {
		_ = n.Start()
		wg.Done()
	}()
}

func TestNode_Chain(t *testing.T) {
	r := &sliceReader{values: []string{"1", "2", "3"}}
	w := &sliceWriter{}
	n1 := selina.NewNode("A", r)
	n2 := selina.NewNode("B", &dummyWorker{})
	n3 := selina.NewNode("C", &dummyWorker{})
	n4 := selina.NewNode("D", w)
	n1.Chain(n2).
		Chain(n3).
		Chain(n4)

	wg := &sync.WaitGroup{}
	startNode(wg, n1)
	startNode(wg, n2)
	startNode(wg, n3)
	startNode(wg, n4)
	wg.Wait()
	if !reflect.DeepEqual(r.values, w.values) {
		t.Fatalf("Chain() values differ got %v , want %v", w.values, r.values)
	}
}

type lazyWorker struct{}

func (l *lazyWorker) Process(ctx context.Context, in <-chan []byte, out chan<- []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Hour):
		return nil
	}
}

func TestNode_Stop(t *testing.T) {
	n1 := selina.NewNode("A", &lazyWorker{})
	stoped := make(chan error, 1)
	go func() {
		err := n1.Start()
		stoped <- err
	}()
	_ = n1.Stop()
	select {
	case err := <-stoped:
		if err != context.Canceled {
			t.Fatalf("Start() err = %v", err)
		}
		return
	case <-time.After(time.Second * 5):
		t.Fatalf("Stop() is not working")
	}

}

type produceN struct {
	count   int
	message []byte
}

func (p *produceN) Process(ctx context.Context, input <-chan []byte, output chan<- []byte) error {
	for i := 0; i < p.count; i++ {
		output <- p.message
	}
	close(output)
	return nil
}

type sink struct{}

func (s *sink) Process(ctx context.Context, input <-chan []byte, output chan<- []byte) error {
	for range input {
	}
	close(output)
	return nil
}

func Benchmark_Node(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		start := selina.NewNode("start", &produceN{count: 100, message: []byte("byte")})
		end := selina.NewNode("end", &sink{})
		start.Chain(end)
		b.StartTimer()
		go func() {
			_ = start.Start()
		}()
		_ = end.Start()
	}
}
