package selina_test

import (
	"testing"

	"github.com/licaonfee/selina"
)

func TestSimplePipeline_StartAll(t *testing.T) {
	p := selina.NewSimplePipeline(
		selina.NewNode("n1", &lazyWorker{}),
		selina.NewNode("n2", &lazyWorker{}))
	selina.ATPipelineStartAll(p, t)
}

func TestSimplePipeline_Cancel(t *testing.T) {
	p := selina.NewSimplePipeline(
		selina.NewNode("n1", &lazyWorker{}),
		selina.NewNode("n2", &lazyWorker{}))
	selina.ATPipelineContextCancel(p, t)
}
