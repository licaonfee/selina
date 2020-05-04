package selina_test

import (
	"testing"

	"github.com/licaonfee/selina"
)

func TestLinealPipelineStartAll(t *testing.T) {
	p := selina.LinealPipeline(
		selina.NewNode("n1", &lazyWorker{}),
		selina.NewNode("n2", &lazyWorker{}))
	if err := selina.ATPipelineStartAll(p); err != nil {
		t.Fatal(err)
	}
}

func TestLinealPipelineCancel(t *testing.T) {
	p := selina.LinealPipeline(
		selina.NewNode("n1", &lazyWorker{}),
		selina.NewNode("n2", &lazyWorker{}))
	if err := selina.ATPipelineContextCancel(p); err != nil {
		t.Fatal(err)
	}
}

func TestLinealPipelineStats(t *testing.T) {
	n1 := selina.NewNode("n1", &produceN{count: 10, message: []byte("b")})
	n2 := selina.NewNode("n2", &dummyWorker{})
	n3 := selina.NewNode("n3", &sink{})
	n1.Chain(n2).Chain(n3)
	p := selina.FreePipeline(n1, n2, n3)
	if err := selina.ATPipelineStats(p); err != nil {
		t.Fatal(err)
	}
}

func TestFreePipelineStartAll(t *testing.T) {
	n1 := selina.NewNode("n1", &lazyWorker{})
	n2 := selina.NewNode("n2", &lazyWorker{})
	n3 := selina.NewNode("n3", &lazyWorker{})
	n1.Chain(n2).Chain(n3)
	p := selina.FreePipeline(n1, n2, n3)
	if err := selina.ATPipelineStartAll(p); err != nil {
		t.Fatal(err)
	}
}

func TestFreePipelineCancel(t *testing.T) {
	n1 := selina.NewNode("n1", &lazyWorker{})
	n2 := selina.NewNode("n2", &lazyWorker{})
	n3 := selina.NewNode("n3", &lazyWorker{})
	n1.Chain(n2).Chain(n3)
	p := selina.FreePipeline(n1, n2, n3)
	if err := selina.ATPipelineContextCancel(p); err != nil {
		t.Fatal(err)
	}
}

func TestFreePipelineStats(t *testing.T) {
	p := selina.LinealPipeline(
		selina.NewNode("n1", &produceN{count: 10, message: []byte("b")}),
		selina.NewNode("n2", &dummyWorker{}),
		selina.NewNode("n3", &sink{}))
	if err := selina.ATPipelineStats(p); err != nil {
		t.Fatal(err)
	}
}
