package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers/regex"
	"github.com/licaonfee/selina/workers/text"
)

const sample = `this is a sample text
this is second line
#lines started with # will be skipped
this line pass`

func main() {
	rd := strings.NewReader(sample)
	input := selina.NewNode("Read", text.NewReader(text.ReaderOptions{Reader: rd}))
	//https://regex101.com/r/7ZS3Uw/1
	filter := selina.NewNode("Filter", regex.NewFilter(regex.FilterOptions{Pattern: "^[^#].+"}))
	output := selina.NewNode("Write", text.NewWriter(text.WriterOptions{Writer: os.Stdout}))
	pipe := selina.NewSimplePipeline(input, filter, output)
	if err := pipe.Run(context.Background()); err != nil {
		fmt.Printf("ERR: %v\n", err)
	}
	for name, stat := range pipe.Stats() {
		fmt.Printf("Node:%s=%#v\n", name, stat)
	}
}
