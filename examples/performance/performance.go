package main

import (
	"context"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/gops/agent"
	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers/csv"
	"github.com/licaonfee/selina/workers/text"
)

const sample = `{"name":"Jimmy", "age": 22, "pet":"cat"}
{"name":"Anne Mc", "age":35, "pet":"dog"}
{"name":"Andrew; Don", "age":"19.7", "pet":"fish"}
`

type infiniteReader struct {
	sample string
	rd     *strings.Reader
}

func (i *infiniteReader) Read(b []byte) (int, error) {
	n, err := i.rd.Read(b)
	if err == io.EOF {
		i.rd.Reset(sample)
		return i.Read(b)
	}
	return n, err
}

func newInfiniteReader(sample string) *infiniteReader {
	rd := &infiniteReader{}
	rd.sample = sample
	rd.rd = strings.NewReader(sample)
	return rd
}

func main() {
	const defaultDuration = time.Second * 5
	duration := flag.Duration("duration", defaultDuration, "Run for X time")
	flag.Parse()
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatal(err)
	}
	rd := newInfiniteReader(sample)
	input := selina.NewNode("Read", text.NewReader(text.ReaderOptions{Reader: rd}))
	//Just print name and pet
	filter := selina.NewNode("CSV", csv.NewEncoder(csv.EncoderOptions{Comma: ';', UseCRLF: false, Header: []string{"name", "pet"}}))
	output := selina.NewNode("Write", text.NewWriter(text.WriterOptions{Writer: ioutil.Discard, SkipNewLine: true}))
	pipe := selina.LinealPipeline(input, filter, output)

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()
	if err := pipe.Run(ctx); err != context.DeadlineExceeded {
		log.Printf("ERR: %v\n", err)
	}
	for _, node := range pipe.Nodes() {
		stat := node.Stats()
		log.Printf("Node:%s(%s)=Send: %d, Recv: %d\n", node.Name(), node.ID(), stat.Sent, stat.Received)
	}
	selina.Graph(pipe, os.Stdout)
}
