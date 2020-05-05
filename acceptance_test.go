package selina_test

import (
	"context"
	"errors"
	"testing"

	"github.com/licaonfee/selina"
	"golang.org/x/sync/errgroup"
)

var _ (selina.Pipeliner) = (*idealPipeline)(nil)

type idealPipeline struct {
	nodes []*selina.Node
}

func (i *idealPipeline) Nodes() []*selina.Node {
	return i.nodes
}

func (i *idealPipeline) Run(ctx context.Context) error {
	if len(i.nodes) < 2 {
		return errors.New("not enough")
	}
	for j := 1; j < len(i.nodes); j++ {
		curr := i.nodes[j-1]
		next := i.nodes[j]
		curr.Chain(next)
	}
	g, ct := errgroup.WithContext(ctx)
	for _, n := range i.nodes {
		node := n
		g.Go(func() error {
			return node.Start(ct)
		})
	}
	return g.Wait()
}

func (i *idealPipeline) Stats() map[string]selina.Stats {
	ret := make(map[string]selina.Stats)
	for _, n := range i.nodes {
		ret[n.ID()] = n.Stats()
	}
	return ret
}

var _ selina.Pipeliner = (*noStartPipeline)(nil)

type noStartPipeline struct {
	nodes []*selina.Node
}

func (n *noStartPipeline) Run(ctx context.Context) error {
	return nil
}

func (n *noStartPipeline) Nodes() []*selina.Node {
	return n.nodes
}
func (n *noStartPipeline) Stats() map[string]selina.Stats {
	return nil
}

var _ selina.Pipeliner = (*noCancelPipeline)(nil)

var errNocancel = errors.New("no cancel error")

type noCancelPipeline struct {
	nodes []*selina.Node
}

func (n *noCancelPipeline) Run(ctx context.Context) error {
	return errNocancel
}
func (n *noCancelPipeline) Nodes() []*selina.Node {
	return n.nodes
}
func (n *noCancelPipeline) Stats() map[string]selina.Stats {
	return nil
}

var _ selina.Pipeliner = (*missingStatsPipeline)(nil)

type missingStatsPipeline struct {
	nodes []*selina.Node
}

func (n *missingStatsPipeline) Run(ctx context.Context) error {
	return nil
}
func (n *missingStatsPipeline) Nodes() []*selina.Node {
	return n.nodes
}
func (n *missingStatsPipeline) Stats() map[string]selina.Stats {
	return make(map[string]selina.Stats)
}

func TestATPipelineAcceptance(t *testing.T) {
	tests := []struct {
		name    string
		pipe    selina.Pipeliner
		test    func(selina.Pipeliner) error
		wantErr error
	}{
		{
			name: "Start All Success",
			pipe: &idealPipeline{nodes: []*selina.Node{
				selina.NewNode("n1", &lazyWorker{}),
				selina.NewNode("n2", &lazyWorker{}),
			}},
			test:    selina.ATPipelineStartAll,
			wantErr: nil,
		},
		{
			name:    "Start All Without nodes",
			pipe:    &idealPipeline{},
			test:    selina.ATPipelineStartAll,
			wantErr: selina.ErrNotHaveNodes,
		},
		{
			name: "Start All No start all nodes",
			pipe: &noStartPipeline{nodes: []*selina.Node{
				selina.NewNode("n1", &lazyWorker{}),
				selina.NewNode("n2", &lazyWorker{}),
			}},
			test:    selina.ATPipelineStartAll,
			wantErr: selina.ErrInconsistentStart,
		},
		{
			name: "Context Cancel Success",
			pipe: &idealPipeline{nodes: []*selina.Node{
				selina.NewNode("n1", &lazyWorker{}),
				selina.NewNode("n2", &lazyWorker{}),
			}},
			test:    selina.ATPipelineContextCancel,
			wantErr: nil,
		},
		{
			name: "Context Not Cancel propagation",
			pipe: &noCancelPipeline{nodes: []*selina.Node{
				selina.NewNode("n1", &lazyWorker{}),
				selina.NewNode("n2", &lazyWorker{})},
			},
			test:    selina.ATPipelineContextCancel,
			wantErr: errNocancel,
		},
		{
			name: "Stats failed Pipeliner", //Use same test case as no cancel
			pipe: &noCancelPipeline{nodes: []*selina.Node{
				selina.NewNode("n1", &lazyWorker{}),
				selina.NewNode("n2", &lazyWorker{})},
			},
			test:    selina.ATPipelineStats,
			wantErr: errNocancel,
		},
		{
			name: "Stats success",
			pipe: &idealPipeline{nodes: []*selina.Node{
				selina.NewNode("n1", &produceN{count: 10, message: []byte("ninebytes")}),
				selina.NewNode("n2", &sink{})},
			},
			test:    selina.ATPipelineStats,
			wantErr: nil,
		},
		{
			name: "Stats missing",
			pipe: &missingStatsPipeline{nodes: []*selina.Node{
				selina.NewNode("n1", &produceN{count: 10, message: []byte("ninebytes")}),
				selina.NewNode("n2", &sink{})},
			},
			test:    selina.ATPipelineStats,
			wantErr: selina.ErrMissingStats,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.test(tt.pipe); err != tt.wantErr {
				t.Fatal(err)
			}
		})
	}
}
