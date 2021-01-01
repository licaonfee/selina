package csv

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"sort"
	"strconv"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Encoder)(nil)

type EncoderOptions struct {
	//Header acts as a filter, if a field is not in header is skipped
	Header []string
	//Comma default ,
	Comma rune
	//UseCRLF use \r\n instead of \n
	UseCRLF    bool
	Handler    selina.ErrorHandler
	ReadFormat selina.Unmarshaler
}

type Encoder struct {
	opts EncoderOptions
}

func (e *Encoder) Process(ctx context.Context, args selina.ProcessArgs) error {
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
	errHandler := selina.DefaultErrorHanler
	if e.opts.Handler != nil {
		errHandler = e.opts.Handler
	}

	var headerWriten bool
	rf := json.Unmarshal
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
			err := rf(msg, &data)
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

func NewEncoder(opts EncoderOptions) *Encoder {
	return &Encoder{opts: opts}
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

func sendData(ctx context.Context, row []string, w *csv.Writer, buff *bytes.Buffer, output chan<- []byte) error {
	buff.Reset()
	if err := w.Write(row); err != nil {
		return err
	}
	w.Flush()
	b := make([]byte, buff.Len())
	copy(b, buff.Bytes())
	if err := selina.SendContext(ctx, b, output); err != nil {
		return err
	}
	return nil
}

var _ selina.Worker = (*Decoder)(nil)

type DecoderOptions struct {
	Header  []string
	Comma   rune
	Comment rune
	Handler selina.ErrorHandler
	Codec   selina.Marshaler
}

type Decoder struct {
	opts DecoderOptions
}

func (d *Decoder) Process(ctx context.Context, args selina.ProcessArgs) error {
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
	errHandler := selina.DefaultErrorHanler
	if d.opts.Handler != nil {
		errHandler = d.opts.Handler
	}
	codec := json.Marshal
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
			_, _ = buff.Write(msg)
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
				return err
			}
			if err := selina.SendContext(ctx, b, args.Output); err != nil {
				return err
			}
		}
	}
}

func NewDecoder(opts DecoderOptions) *Decoder {
	return &Decoder{opts: opts}
}
