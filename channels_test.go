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

func TestBroadcasterBroadcastPanic(t *testing.T) {
	const waitForbroascast = time.Millisecond * 20
	b := selina.Broadcaster{}
	_ = b.Client() // drop a client
	in := make(chan *bytes.Buffer)
	go func() {
		b.Broadcast(in)
	}()
	time.Sleep(waitForbroascast)
	panicCall := func() (err error) {
		defer func() {
			err = fmt.Errorf("%v", recover())
		}()
		_ = b.Client()
		return
	}
	if err := panicCall(); err == nil || err.Error() != "call Client after Broadcast" {
		t.Fatalf("Broadcast allow add clients while running")
	}

}
func TestBroadcasterBroadcast(t *testing.T) {
	const clientCount = 2
	inChan := selina.SliceAsChannelOfBuffer([]string{"foo", "bar", "baz"}, true)
	want := [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}
	b := selina.Broadcaster{}
	var out []<-chan *bytes.Buffer
	for i := 0; i < clientCount; i++ {
		c := b.Client()
		out = append(out, c)
	}
	go b.Broadcast(inChan)
	var wg sync.WaitGroup
	for _, c := range out {
		wg.Add(1)
		go func(can <-chan *bytes.Buffer) {
			received := make([][]byte, 0)
			for d := range can {
				received = append(received, d.Bytes())
				selina.FreeBuffer(d)
			}
			if !sameElements(received, want) {
				t.Errorf("Branch() got = %v, want = %v", received, want)
			}
			wg.Done()
		}(c)
	}
	wg.Wait()
}

func TestReceiverReceive(t *testing.T) {
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
		r.Watch(selina.SliceAsChannelOfBuffer(values, true))
	}
	recv := r.Receive()
	var got [][]byte
	for msg := range recv {
		got = append(got, msg.Bytes())
	}
	if !sameElements(got, want) {
		t.Fatalf("Receive() got = %v , want = %v", got, want)
	}
}

func drainChan[T any](in <-chan T) {
	for range in {
	}
}

func BenchmarkBroadcaster(b *testing.B) {
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
				go drainChan(broad.Client())
			}
			input := make(chan *bytes.Buffer)
			go func() {
				for i := 0; i < b.N; i++ {
					buff := bytes.NewBuffer(nil)
					buff.Write(bench.msg)
					input <- buff
				}
				close(input)
			}()
			broad.Broadcast(input)
		})
	}
}

func BenchmarkReceiver(b *testing.B) {
	// This function need to be improved to acquire more precise data
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
				data := make([]*bytes.Buffer, bench.messageCount)
				for j := 0; j < bench.messageCount; j++ {
					x := bytes.NewBuffer(nil)
					x.Write(bench.msg)
					data[j] = x
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
	outA := make(chan *bytes.Buffer, 1)
	msg := bytes.NewBuffer([]byte("foo"))
	// Case 1: message delivered
	if err := selina.SendContext(context.Background(), msg, outA); err != nil {
		t.Fatalf("SendContext() err = %v", err)
	}
	got := <-outA
	if !bytes.Equal(got.Bytes(), msg.Bytes()) {
		t.Fatalf("SendContext() message corupted got = %v, want = %v", got, msg)
	}
	// Case 2: context canceled
	const cancelAfter = time.Millisecond * 50
	outB := make(chan *bytes.Buffer)
	ctx, cancel := context.WithTimeout(context.Background(), cancelAfter)
	defer cancel()
	if err := selina.SendContext(ctx, msg, outB); err != context.DeadlineExceeded {
		t.Fatalf("SendContext() unexpected err = %v", err)
	}
}
