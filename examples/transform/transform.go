package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers/csv"
	"github.com/licaonfee/selina/workers/workers"
)

const sample = `{"name":"Jimmy", "age": 22, "pet":"cat"}
{"name":"Anne Mc", "age":35, "pet":"dog"}
{"name":"Andrew; Don", "age":"19.7", "pet":"fish"}`

func main() {
	rd := strings.NewReader(sample)
	input := selina.NewNode("Read", workers.NewReader(workers.TextReaderOptions{Reader: rd}))
	//Just print name and pet
	filter := selina.NewNode("CSV", csv.NewCSVEncoder(csv.EncoderOptions{Comma: ';', UseCRLF: false, Header: []string{"name", "pet"}}))
	output := selina.NewNode("Write", workers.NewTextWriter(workers.TextWriterOptions{Writer: os.Stdout, SkipNewLine: true}))
	pipe := selina.LinealPipeline(input, filter, output)
	if err := pipe.Run(context.Background()); err != nil {
		fmt.Printf("ERR: %v\n", err)
	}
	for name, stat := range pipe.Stats() {
		fmt.Printf("Node:%s=%v\n", name, stat)
	}
}
