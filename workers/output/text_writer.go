package output

import (
	"bufio"
	"context"
	"io"
	"os"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*TextWriter)(nil)

type TextWriterOptions struct {
	Filename string
}

type TextWriter struct {
	opts TextWriterOptions
}

func (t *TextWriter) Process(ctx context.Context, in <-chan []byte, out chan<- []byte) (err error) {
	defer close(out)
	var f io.WriteCloser
	f, err = os.Create(t.opts.Filename)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)
	defer func() {
		if errFlush := w.Flush(); errFlush != nil {
			err = errFlush
			return
		}
		if errClose := f.Close(); errClose != nil {
			err = errClose
		}

	}()
	newLine := []byte("\n")
	for {
		select {
		case msg, ok := <-in:
			if !ok {
				return nil
			}
			_, err = w.Write(msg)
			if err != nil {
				return
			}
			_, err = w.Write(newLine)
			if err != nil {
				return
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}

}

func NewTextWriter(opts TextWriterOptions) *TextWriter {
	w := &TextWriter{opts: opts}
	return w
}
