package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Writer)(nil)

//WriterOptions provide parameters to create a Writer
type WriterOptions struct {
	//Driver which driver should be used
	//this require that users import required driver
	Driver string
	//ConnStr connection string relative to Driver
	ConnStr string
	//Table in which table data will be inserted
	Table string
	//Builder (optional) customize SQL generation
	Builder QueryBuilder
	Codec   selina.Unmarshaler
}

//Check if a combination of options is valid
//this not guarantees that worker will not fail
func (o WriterOptions) Check() error {
	var driverOk bool
	for _, d := range sql.Drivers() {
		if d == o.Driver {
			driverOk = true
		}
	}
	if !driverOk {
		return fmt.Errorf("missing driver '%s'", o.Driver)
	}
	if o.Table == "" {
		return fmt.Errorf("invalid table name \"%s\"", o.Table)
	}
	return nil
}

//Writer a Worker that insert data into database
type Writer struct {
	opts WriterOptions
}

//Process implements Worker interface
func (s *Writer) Process(ctx context.Context, args selina.ProcessArgs) error {
	defer close(args.Output)
	if err := s.opts.Check(); err != nil {
		return err
	}
	conn, err := sql.Open(s.opts.Driver, s.opts.ConnStr)
	if err != nil {
		return err
	}
	codec := json.Unmarshal
	if s.opts.Codec != nil {
		codec = s.opts.Codec
	}
	for {
		select {
		case data, ok := <-args.Input:
			if !ok {
				return nil
			}
			cols, values, err := deserialize(codec, data)
			if err != nil {
				return err
			}
			if s.opts.Builder == nil {
				s.opts.Builder = &DefaultQueryBuilder{}
			}

			query := s.opts.Builder.Insert(s.opts.Table, cols)
			_, err = conn.ExecContext(ctx, query, values...)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

//NewWriter create a new Writer with given options
func NewWriter(opts WriterOptions) *Writer {
	return &Writer{opts: opts}
}

func deserialize(codec selina.Unmarshaler, data []byte) (cols []string, values []interface{}, err error) {
	obj := make(map[string]interface{})
	if err := codec(data, &obj); err != nil {
		return nil, nil, err
	}
	cols = make([]string, 0, len(obj))
	values = make([]interface{}, 0, len(obj))
	for name, val := range obj {
		cols = append(cols, name)
		values = append(values, val)
	}
	return cols, values, nil
}
