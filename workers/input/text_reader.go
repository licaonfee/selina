package input

import (
	"bufio"
	"context"
	"io"
	"os"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*TextReader)(nil)

type TextReaderOptions struct {
	Filename string
}

type TextReader struct {
	opts TextReaderOptions
}

func (t *TextReader) Process(ctx context.Context, input <-chan []byte, out chan<- []byte) (err error) {
	var f io.ReadCloser
	defer close(out)
	f, err = os.Open(t.opts.Filename)
	if err != nil {
		return err
	}
	defer func() {
		if errClose := f.Close(); errClose != nil {
			err = errClose
		}
	}()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		select {
		case _, ok := <-input:
			if !ok {
				return nil
			}
		case out <- []byte(sc.Text()):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func NewTextReader(opts TextReaderOptions) *TextReader {
	t := TextReader{opts: opts}
	return &t
}
