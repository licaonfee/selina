package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/licaonfee/magiccol"
	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Reader)(nil)

//ReaderOptions provide parameters to create a Reader
type ReaderOptions struct {
	//Driver which driver should be used
	//this require that users import required driver
	Driver string
	//ConnStr connection string relative to Driver
	ConnStr string
	//Query which SQL select will be executed into database
	Query string
	//Mapper allow to configure type Scan, default magiccol.DefaultMapper
	Mapper *magiccol.Mapper
	//WriteFormat default is json.Marshal
	WriteFormat selina.Marshaler
}

//Check if a combination of options is valid
//this not guarantees that worker will not fail
func (o ReaderOptions) Check() error {
	var driverOK bool
	for _, d := range sql.Drivers() {
		if d == o.Driver {
			driverOK = true
		}
	}
	if !driverOK {
		return fmt.Errorf("missing driver '%s'", o.Driver)
	}
	//validate if a query is valid or not require
	//implements any posible dialect so we just check
	//only if there is a query
	if o.Query == "" {
		return fmt.Errorf("empty query")
	}
	return nil
}

//Reader a Worker that execute a given Query and export data via output channel
type Reader struct {
	opts ReaderOptions
}

//Process implements Worker interface
func (s *Reader) Process(ctx context.Context, args selina.ProcessArgs) (err error) {
	defer close(args.Output)
	if err := s.opts.Check(); err != nil {
		return err
	}
	db, err := sql.Open(s.opts.Driver, s.opts.ConnStr)
	if err != nil {
		return err
	}

	var input <-chan []byte
	if args.Input != nil {
		input = args.Input
	} else {
		in := make(chan []byte, 1)
		in <- nil
		close(in)
		input = in
	}
	codec := json.Marshal
	if s.opts.WriteFormat != nil {
		codec = s.opts.WriteFormat
	}
	for {
		select {
		case _, ok := <-input:
			if !ok {
				return nil
			}
			rows, err := db.QueryContext(ctx, s.opts.Query)
			if err != nil {
				return err
			}
			if err := s.serializeRows(ctx, codec, rows, args.Output); err != nil {
				return err
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Reader) serializeRows(ctx context.Context, codec selina.Marshaler, rows *sql.Rows, out chan<- []byte) error {
	defer rows.Close()
	obj := make(map[string]interface{})
	m := s.opts.Mapper
	if m == nil {
		m = magiccol.DefaultMapper()
	}
	sc, err := magiccol.NewScanner(magiccol.Options{Rows: rows, Mapper: m})
	if err != nil {
		return err
	}
	for sc.Scan() {
		sc.SetMap(obj)
		msg, err := codec(obj)
		if err != nil {
			return err
		}
		if err := selina.SendContext(ctx, msg, out); err != nil {
			return err
		}
	}
	if sc.Err() != nil {
		return sc.Err()
	}
	return nil
}

//NewReader create a new Reader with given options
func NewReader(opts ReaderOptions) *Reader {
	return &Reader{opts: opts}
}
