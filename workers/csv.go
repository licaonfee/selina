package workers

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*CSVEncoder)(nil)

// CSVEncoderOptions configure csv encoding
type CSVEncoderOptions struct {
	// Header acts as a filter, if a field is not in header is skipped
	Header []string
	// Comma default ,
	Comma rune
	// UseCRLF use \r\n instead of \n
	UseCRLF    bool
	Handler    selina.ErrorHandler
	ReadFormat selina.Unmarshaler
}

// CSVEncoder transform messages into csv text
type CSVEncoder struct {
	opts CSVEncoderOptions
}

// Process implements selina.Worker interface
func (e *CSVEncoder) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	if args.Input == nil {
		return selina.ErrNilUpstream
	}
	buff := &bytes.Buffer{}
	w := csv.NewWriter(buff)
	if e.opts.Comma != rune(0) {
		w.Comma = e.opts.Comma
	}
	w.UseCRLF = e.opts.UseCRLF
	errHandler := selina.DefaultErrorHandler
	if e.opts.Handler != nil {
		errHandler = e.opts.Handler
	}

	var headerWriten bool
	rf := selina.DefaultUnmarshaler
	if e.opts.ReadFormat != nil {
		rf = e.opts.ReadFormat
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-args.Input:
			if !ok {
				return nil
			}
			data := make(map[string]interface{})
			err := rf(msg.Bytes(), &data)
			selina.FreeBuffer(msg)
			switch {
			case err == nil:
			case errHandler(err):
				continue
			default:
				return err
			}
			if !headerWriten {
				if len(e.opts.Header) == 0 {
					e.opts.Header = getHeader(data)
				}
				if err := sendData(ctx, e.opts.Header, w, buff, args.Output); err != nil {
					return err
				}
				headerWriten = true
			}
			res := getRow(e.opts.Header, data)
			if err := sendData(ctx, res, w, buff, args.Output); err != nil {
				return err
			}
		}
	}
}

// NewCSVEncoder returns a new Encoder with given options
func NewCSVEncoder(opts CSVEncoderOptions) *CSVEncoder {
	return &CSVEncoder{opts: opts}
}

func getHeader(sample map[string]interface{}) []string {
	header := make([]string, 0, len(sample))
	for k := range sample {
		header = append(header, k)
	}
	sort.Strings(header)
	return header
}

func getRow(header []string, data map[string]interface{}) []string {
	res := make([]string, len(header))
	for i := 0; i < len(header); i++ {
		value, ok := data[header[i]]
		if !ok {
			continue
		}
		switch v := value.(type) {
		case float64:
			res[i] = strconv.FormatFloat(v, 'f', -1, 64)
		case string:
			res[i] = v
		}
	}
	return res
}

func sendData(ctx context.Context, row []string, w *csv.Writer, buff *bytes.Buffer, output chan<- *bytes.Buffer) error {
	buff.Reset()
	if err := w.Write(row); err != nil {
		return err
	}
	w.Flush()
	b := selina.GetBuffer()
	_, _ = io.Copy(b, buff)
	if err := selina.SendContext(ctx, b, output); err != nil {
		return err
	}
	return nil
}

var _ selina.Worker = (*CSVDecoder)(nil)

// CSVDecoderOptions configure csv read format
type CSVDecoderOptions struct {
	Header  []string
	Comma   rune
	Comment rune
	Handler selina.ErrorHandler
	Codec   selina.Marshaler
}

// CSVDecoder parse csv lines into key value pairs
type CSVDecoder struct {
	opts CSVDecoderOptions
}

// Process implements selina.Worker interface
func (d *CSVDecoder) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	if args.Input == nil {
		return selina.ErrNilUpstream
	}
	buff := &bytes.Buffer{}
	r := csv.NewReader(buff)
	if d.opts.Comma != rune(0) {
		r.Comma = d.opts.Comma
	}
	r.Comment = d.opts.Comment
	r.ReuseRecord = true
	errHandler := selina.DefaultErrorHandler
	if d.opts.Handler != nil {
		errHandler = d.opts.Handler
	}
	codec := selina.DefaultMarshaler
	if d.opts.Codec != nil {
		codec = d.opts.Codec
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-args.Input:
			if !ok {
				return nil
			}
			buff.Reset()
			_, _ = io.Copy(buff, msg)
			selina.FreeBuffer(msg)
			row, err := r.Read()
			switch {
			case err == nil:
			case err == io.EOF || errHandler(err):
				continue
			default:
				return err
			}
			res := make(map[string]interface{})
			for i := 0; i < len(row); i++ {
				res[d.opts.Header[i]] = row[i]
			}
			b, err := codec(res)
			if err != nil {
				return fmt.Errorf("encoding %w", err)
			}
			nb := selina.GetBuffer()
			nb.Write(b)
			if err := selina.SendContext(ctx, nb, args.Output); err != nil {
				return err
			}
		}
	}
}

// NewCSVDecoder return a new csv decoder with given options
func NewCSVDecoder(opts CSVDecoderOptions) *CSVDecoder {
	return &CSVDecoder{opts: opts}
}
