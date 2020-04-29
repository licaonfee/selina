package sql

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*Reader)(nil)

type ReaderOptions struct {
	Driver  string
	ConnStr string
	Query   string
}

type Reader struct {
	opts ReaderOptions
}

func (s *Reader) Process(ctx context.Context, in <-chan []byte, out chan<- []byte) (err error) {
	defer close(out)
	db, err := sql.Open(s.opts.Driver, s.opts.ConnStr)
	if err != nil {
		return err
	}
	for {
		select {
		case _, ok := <-in:
			if !ok {
				return nil
			}
		default:
			rows, err := db.QueryContext(ctx, s.opts.Query)
			if err != nil {
				return err
			}
			if err := serializeRows(ctx, rows, out); err != nil {
				return err
			}
			return nil
		}
	}
}

func serializeRows(ctx context.Context, rows *sql.Rows, out chan<- []byte) error {
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	obj := make(map[string]interface{})
	values := make([]interface{}, len(cols))
	pointers := make([]interface{}, len(cols))
	for i, k := range cols {
		obj[k] = values[i]
		pointers[i] = &values[i]
	}
	for rows.Next() {
		if err := rows.Scan(pointers...); err != nil {
			return err
		}
		for i, v := range values {
			obj[cols[i]] = v
		}
		msg, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		if err := sendMessage(ctx, out, msg); err != nil {
			return err
		}
	}
	return nil
}

func sendMessage(ctx context.Context, out chan<- []byte, msg []byte) error {
	select {
	case out <- msg:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func NewSQLReader(opts ReaderOptions) *Reader {
	return &Reader{opts: opts}
}
