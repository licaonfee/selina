package selina

import (
	"context"
	"errors"
	"time"
)

const stopPipelineTime = time.Millisecond * 20

var ErrNotHaveNodes = errors.New("Pipeliner does not have nodes")
var ErrInconsistentStart = errors.New("Pipeliner does not start all nodes")
var ErrMissingStats = errors.New("missing nodes in Stats map")

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
		return ErrNotHaveNodes
	}
	for _, n := range p.Nodes() {
		if !n.Running() {
			return ErrInconsistentStart
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
		return err
	}
	return nil
}

func ATPipelineStats(p Pipeliner) error {
	if err := p.Run(context.Background()); err != nil {
		return err
	}
	stats := p.Stats()
	if len(stats) != len(p.Nodes()) {
		return ErrMissingStats
	}
	return nil
}
