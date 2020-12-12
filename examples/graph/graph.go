package main

import (
	"io/ioutil"
	"os"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers/csv"
	"github.com/licaonfee/selina/workers/custom"
	"github.com/licaonfee/selina/workers/regex"
	"github.com/licaonfee/selina/workers/sql"
	"github.com/licaonfee/selina/workers/text"
)

// This pipeline does nothing and will panic if you start it
// just exists to renderize a fancy graph
func main() {
	input := selina.NewNode("Read", text.NewReader(text.ReaderOptions{}))
	filter := selina.NewNode("CSV", csv.NewEncoder(csv.EncoderOptions{}))
	filter2 := selina.NewNode("Filter", regex.NewFilter(regex.FilterOptions{}))
	custom := selina.NewNode("Custom", custom.NewFunction(custom.FunctionOptions{}))
	output := selina.NewNode("Write", text.NewWriter(text.WriterOptions{Writer: ioutil.Discard, SkipNewLine: true}))
	output2 := selina.NewNode("Store", sql.NewWriter(sql.WriterOptions{}))
	input.Chain(filter).Chain(output)
	input.Chain(filter2).Chain(custom).Chain(output2)

	pipe := selina.FreePipeline(input, filter, filter2, output, custom, output2)

	selina.Graph(pipe, os.Stdout)
}
