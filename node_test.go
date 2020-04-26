package selina_test

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/licaonfee/selina"
)

func startNode(wg *sync.WaitGroup, n *selina.Node) {
	wg.Add(1)
	go func() {
		_ = n.Start(context.Background())
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

func TestNode_Stop(t *testing.T) {
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
		if err != context.Canceled {
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
