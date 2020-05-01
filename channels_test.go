package selina_test

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"context"

	"github.com/licaonfee/selina"
)

func sameElements(a [][]byte, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	var matches = 0
	for _, v := range a {
		for _, w := range b {
			if bytes.Equal(v, w) {
				matches++
				break
			}
		}
	}
	return matches == len(a)
}

func TestBroadcaster_Broadcast(t *testing.T) {
	const clientCount = 2
	inChan := selina.SliceAsChannel([]string{"foo", "bar", "baz"}, true)
	want := [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}
	b := selina.Broadcaster{}
	var out []<-chan []byte
	for i := 0; i < clientCount; i++ {
		c := b.Client()
		out = append(out, c)
	}
	go b.Broadcast(inChan)
	var wg sync.WaitGroup
	for _, c := range out {
		wg.Add(1)
		go func(can <-chan []byte) {
			received := make([][]byte, 0)
			for d := range can {
				received = append(received, d)
			}
			if !sameElements(received, want) {
				t.Errorf("Branch() got = %s, want = %s", received, want)
			}
			wg.Done()
		}(c)
	}
	wg.Wait()
}

func TestReceiver_Receive(t *testing.T) {
	const serversCount = 3
	const messageCount = 3
	var want [][]byte
	r := selina.Receiver{}

	for i := 1; i <= serversCount; i++ {
		var values []string
		for j := 1; j <= messageCount; j++ {
			msg := strconv.Itoa(i * j)
			values = append(values, msg)
			want = append(want, []byte(msg))
		}
		r.Watch(selina.SliceAsChannel(values, true))
	}
	recv := r.Receive()
	var got [][]byte
	for msg := range recv {
		got = append(got, msg)
	}
	if !sameElements(got, want) {
		t.Fatalf("Receive() got = %v , want = %v", got, want)
	}
}

func readChan(in <-chan []byte) {
	for range in {
	}
}

func Benchmark_Broadcaster(b *testing.B) {
	benchMarks := []struct {
		name        string
		clientCount int
		msg         []byte
	}{
		{
			name:        "Empty",
			clientCount: 4,
			msg:         []byte{},
		},
		{
			name:        "Little ",
			clientCount: 4,
			msg:         []byte("This is a little message of 32 b"),
		},
		{
			name:        "Little",
			clientCount: 8,
			msg:         []byte("This is a little message of 32 b"),
		},
	}

	for _, bench := range benchMarks {
		b.Run(fmt.Sprintf("%s_c(%d)_b(%d)", bench.name, bench.clientCount, len(bench.msg)), func(b *testing.B) {
			broad := selina.Broadcaster{}
			for i := 0; i < bench.clientCount; i++ {
				go readChan(broad.Client())
			}
			input := make(chan []byte)
			go func() {
				for i := 0; i < b.N; i++ {
					input <- bench.msg
				}
				close(input)
			}()
			broad.Broadcast(input)
		})
	}
}

func Benchmark_Receiver(b *testing.B) {
	//This function need to be improved to acquire more precise data
	benchMarks := []struct {
		name         string
		serverCount  int
		messageCount int
		msg          []byte
	}{
		{
			name:         "Empty",
			serverCount:  4,
			messageCount: 10,
			msg:          []byte{},
		},
		{
			name:         "Little ",
			serverCount:  4,
			messageCount: 10,
			msg:          []byte("This is a little message of 32 b"),
		},
		{
			name:         "Little",
			serverCount:  8,
			messageCount: 10,
			msg:          []byte("This is a little message of 32 b"),
		},
	}

	for _, bench := range benchMarks {
		b.Run(fmt.Sprintf("%s_c(%d)_b(%d)", bench.name, bench.serverCount, len(bench.msg)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				recv := selina.Receiver{}
				data := make([][]byte, bench.messageCount)
				for j := 0; j < bench.messageCount; j++ {
					data[j] = []byte(string(bench.msg))
				}
				recv.Watch(selina.SliceAsChannelRaw(data, true))

				b.StartTimer()
				for range recv.Receive() {
				}
			}
		})
	}
}

func TestSendContext(t *testing.T) {
	outA := make(chan []byte, 1)
	msg := []byte("foo")
	//Case 1: message delivered
	if err := selina.SendContext(context.Background(), msg, outA); err != nil {
		t.Fatalf("SendContext() err = %v", err)
	}
	got := <-outA
	if !bytes.Equal(got, msg) {
		t.Fatalf("SendContext() message corupted got = %v, want = %v", got, msg)
	}
	//Case 2: context canceled
	const cancelAfter = time.Millisecond * 50
	outB := make(chan []byte)
	ctx, cancel := context.WithTimeout(context.Background(), cancelAfter)
	defer cancel()
	if err := selina.SendContext(ctx, msg, outB); err != context.DeadlineExceeded {
		t.Fatalf("SendContext() unexpected err = %v", err)
	}
}
