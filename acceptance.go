package selina

import (
	"context"
	"fmt"
	"time"
)

const stopPipelineTime = time.Millisecond * 20

//ATPipelineStartAll all Nodes in a pipeline mus be started when pipeline.Start is called
func ATPipelineStartAll(p Pipeliner) error {
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
		return fmt.Errorf("Run() pipeline does not have nodes")
	}
	for _, n := range p.Nodes() {
		if !n.Running() {
			return fmt.Errorf("Run() does not start all nodes")
		}
	}
	return nil
}

//ATPipelineContextCancel context must be propagated to all Nodes
func ATPipelineContextCancel(p Pipeliner) error {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(stopPipelineTime)
		cancel()
	}()
	err := p.Run(ctx)
	if err != context.Canceled {
		return fmt.Errorf("Run() err = %v", err)
	}
	return nil
}

func ATPipelineStats(p Pipeliner) error {
	if err := p.Run(context.Background()); err != nil {
		return fmt.Errorf("Stats() err = %v", err)
	}
	stats := p.Stats()
	if len(stats) != len(p.Nodes()) {
		return fmt.Errorf("Stats() missing nodes got = %d , want = %d", len(stats), len(p.Nodes()))
	}
	return nil

}
