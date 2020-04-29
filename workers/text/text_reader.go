package text

import (
	"bufio"
	"context"
	"errors"
	"io"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Reader)(nil)

type ReaderOptions struct {
	Reader    io.Reader
	AutoClose bool
}

type Reader struct {
	opts ReaderOptions
}

func (t *Reader) cleanup() error {
	if t.opts.Reader == nil {
		return nil
	}
	if c, ok := t.opts.Reader.(io.Closer); t.opts.AutoClose && ok {
		return c.Close()
	}
	return nil
}

var ErrNilReader = errors.New("nil io.Reader provided to TextReader")

func (t *Reader) Process(ctx context.Context, input <-chan []byte, out chan<- []byte) (err error) {
	defer func() {
		close(out)
		cerr := t.cleanup()
		if err == nil { //if an error occurred not override it
			err = cerr
		}
	}()
	if t.opts.Reader == nil {
		return ErrNilReader
	}
	sc := bufio.NewScanner(t.opts.Reader)
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

func NewTextReader(opts ReaderOptions) *Reader {
	t := Reader{opts: opts}
	return &t
}
