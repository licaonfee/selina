package selina_test

import (
	"testing"

	"github.com/licaonfee/selina"
)

func TestSimplePipelineStartAll(t *testing.T) {
	p := selina.NewSimplePipeline(
		selina.NewNode("n1", &lazyWorker{}),
		selina.NewNode("n2", &lazyWorker{}))
	if err := selina.ATPipelineStartAll(p); err != nil {
		t.Fatal(err)
	}
}

func TestSimplePipelineCancel(t *testing.T) {
	p := selina.NewSimplePipeline(
		selina.NewNode("n1", &lazyWorker{}),
		selina.NewNode("n2", &lazyWorker{}))
	if err := selina.ATPipelineContextCancel(p); err != nil {
		t.Fatal(err)
	}
}

func TestSimplePipelineStats(t *testing.T) {
	p := selina.NewSimplePipeline(
		selina.NewNode("n1", &produceN{count: 10, message: []byte("b")}),
		selina.NewNode("n2", &dummyWorker{}),
		selina.NewNode("n3", &sink{}))
	if err := selina.ATPipelineStats(p); err != nil {
		t.Fatal(err)
	}
}
