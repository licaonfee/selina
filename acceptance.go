package selina

import (
	"context"
	"testing"
	"time"
)

const stopPipelineTime = time.Millisecond * 20

func ATPipeline_Start_All(p Pipeliner, t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wait := make(chan struct{})
	go func() {
		_ = p.Run(ctx)
		close(wait)
	}()
	time.Sleep(stopPipelineTime)
	cancel()
	<-wait
	if len(p.Nodes()) == 0 {
		t.Fatalf("Run() pipeline does not have nodes")
	}
	for _, n := range p.Nodes() {
		if !n.Running() {
			t.Fatalf("Run() does not start all nodes")
		}
	}
}

func ATPipeline_Context_cancel(p Pipeliner, t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(stopPipelineTime)
		cancel()
	}()
	err := p.Run(ctx)
	if m, ok := err.(*MultiError); ok {
		for n, e := range m.InnerErrors {
			if e != context.Canceled {
				t.Fatalf("Run() node=%s, err= %v", n, err)
			}
		}
	} else {
		t.Fatalf("Run() err = %v", err)
	}
}
