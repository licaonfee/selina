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
	UseCRLF bool
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
	var headerWriten bool
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-args.Input:
			if !ok {
				return nil
			}
			data := make(map[string]interface{})
			if err := json.Unmarshal(msg, &data); err != nil {
				return err
			}
			getHeader(&e.opts.Header, data)
			if !headerWriten {
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

func getHeader(header *[]string, sample map[string]interface{}) {
	if len(*header) > 0 {
		return
	}
	*header = make([]string, 0, len(sample))
	for k := range sample {
		*header = append(*header, k)
	}
	sort.Strings(*header)
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
			if err != nil {
				if err != io.EOF {
					return err
				}
				continue
			}
			res := make(map[string]interface{})
			for i := 0; i < len(row); i++ {
				res[d.opts.Header[i]] = row[i]
			}
			b, err := json.Marshal(res)
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
