package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/licaonfee/selina"
)

var _ selina.Worker = (*SQLWriter)(nil)

type SQLWriterOptions struct {
	Driver  string
	ConnStr string
	Table   string
	Builder QueryBuilder
}
type SQLWriter struct {
	opts SQLWriterOptions
}

func (s *SQLWriter) Process(ctx context.Context, input <-chan []byte, output chan<- []byte) error {
	defer close(output)
	conn, err := sql.Open(s.opts.Driver, s.opts.ConnStr)
	if err != nil {
		return err
	}
	if s.opts.Table == "" {
		return fmt.Errorf("invalid table name \"%s\"", s.opts.Table)
	}
	for {
		select {
		case data, ok := <-input:
			if !ok {
				return nil
			}
			cols, values, err := deserialize(data)
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

func NewSQLWriter(opts SQLWriterOptions) *SQLWriter {
	return &SQLWriter{opts: opts}
}

func deserialize(data []byte) (cols []string, values []interface{}, err error) {
	obj := make(map[string]interface{})
	if err := json.Unmarshal(data, &obj); err != nil {
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
