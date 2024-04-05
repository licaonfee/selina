package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers/regex"
	"github.com/licaonfee/selina/workers/workers"
)

const sample = `this is a sample text
this is second line
#lines started with # will be skipped
this line pass`

func main() {
	rd := strings.NewReader(sample)
	input := selina.NewNode("Read", workers.NewReader(workers.TextReaderOptions{Reader: rd}))
	//https://regex101.com/r/7ZS3Uw/1
	filter := selina.NewNode("Filter", regex.NewFilter(regex.FilterOptions{Pattern: "^[^#].+"}))
	output := selina.NewNode("Write", workers.NewTextWriter(workers.TextWriterOptions{Writer: os.Stdout}))
	pipe := selina.LinealPipeline(input, filter, output)
	if err := pipe.Run(context.Background()); err != nil {
		fmt.Printf("ERR: %v\n", err)
	}
	for name, stat := range pipe.Stats() {
		fmt.Printf("Node:%s=%v\n", name, stat)
	}
}
