package selina_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/licaonfee/selina"
)

func TestCounterSumData(t *testing.T) {
	tests := []struct {
		clients int
		count   int
		message []byte
	}{
		{
			clients: 4,
			count:   100,
			message: []byte("message"),
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("clients_%d_count_%d", tt.clients, tt.count), func(t *testing.T) {
			dc := selina.DataCounter{}
			wg := sync.WaitGroup{}
			wg.Add(tt.clients)
			for i := 0; i < tt.clients; i++ {
				go func(d *selina.DataCounter, count int, msg []byte, wg *sync.WaitGroup) {
					for j := 0; j < count; j++ {
						dc.SumData(msg)
					}
					wg.Done()
				}(&dc, tt.count, tt.message, &wg)
			}
			wg.Wait()
			wantCount := int64(tt.count * tt.clients)
			wantData := int64(len(tt.message) * tt.count * tt.clients)

			gotCount, gotData := dc.Stats()
			if wantCount != gotCount || gotData != wantData {
				t.Fatalf("SumData() got = (%d , %d), want = (%d, %d) ", gotCount, gotData, wantCount, wantData)
			}

		})
	}
}
