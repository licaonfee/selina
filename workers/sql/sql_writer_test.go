package sql_test

import (
	"testing"

	dbsql "database/sql"

	"golang.org/x/net/context"

	"github.com/licaonfee/selina"

	"github.com/licaonfee/selina/workers"
	"github.com/licaonfee/selina/workers/sql"
)

func TestWriterProcess(t *testing.T) {
	//All tests uses DefaultQueryBuilder
	tests := []struct {
		name        string
		opts        sql.WriterOptions
		in          []string
		expectCount int
		countQuery  string
		wantErr     bool
	}{
		{
			name:        "Success one column",
			opts:        sql.WriterOptions{Driver: ramsqlDriver, ConnStr: "w_success_one_column", Table: "members"},
			in:          []string{`{"name":"Nina"}`, `{"name":"Arani"}`},
			countQuery:  "SELECT COUNT(*) FROM members where name='Nina' or name='Arani';",
			expectCount: 2,
			wantErr:     false,
		},
		{
			name:        "Success three columns",
			opts:        sql.WriterOptions{Driver: ramsqlDriver, ConnStr: "w_success_three_columns", Table: "members"},
			in:          []string{`{"id":10,"name":"Alice", "mood":"sad"}`, `{"id":11,"name":"Sarah","mood":"happy"}`},
			countQuery:  "SELECT COUNT(*) FROM members where id=10 or id='11';",
			expectCount: 2,
			wantErr:     false,
		},
		{
			name:        "Success variant columns",
			opts:        sql.WriterOptions{Driver: ramsqlDriver, ConnStr: "w_success_variant_columns", Table: "members"},
			in:          []string{`{"id":15,"name":"Jess"}`, `{"name":"Alex","mood":"happy"}`},
			countQuery:  "SELECT COUNT(*) FROM members where name='Jess' or name='Alex';",
			expectCount: 2,
			wantErr:     false,
		},
		{
			name:    "Invalid connection",
			opts:    sql.WriterOptions{Driver: "", ConnStr: "w_invalid_connection", Table: "members"},
			in:      []string{},
			wantErr: true,
		},
		{
			name:    "Empty table",
			opts:    sql.WriterOptions{Driver: ramsqlDriver, ConnStr: "w_empty_table", Table: ""},
			in:      []string{`{"name":"Valery"}`},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDB(tt.opts.ConnStr)
			s := sql.NewWriter(tt.opts)
			input := selina.SliceAsChannel(tt.in, true)
			output := make(chan []byte)
			if err := s.Process(context.Background(), input, output); (err != nil) != tt.wantErr {
				t.Fatalf("Process() err = %v", err)
				return
			}
			if tt.countQuery == "" {
				return
			}
			conn, _ := dbsql.Open(ramsqlDriver, tt.opts.ConnStr)
			res, err := conn.Query(tt.countQuery)
			if err != nil {
				t.Fatalf("Process() err = %v", err)
				return
			}
			defer res.Close()
			if res.Next() {
				var count int
				_ = res.Scan(&count)
				if count != tt.expectCount {
					t.Fatalf("Process() count = %d , expect %d", count, tt.expectCount)
				}
			}

		})
	}
}

func TestWriterProcessCloseInput(t *testing.T) {
	const dbname = "writer_close_input"
	setupDB(dbname)
	s := sql.NewWriter(sql.WriterOptions{
		Driver:  ramsqlDriver,
		ConnStr: dbname,
		Table:   "members",
	})
	if err := workers.ATProcessCloseInput(s); err != nil {
		t.Fatal(err)
	}
}

func TestWriterProcessCloseOutput(t *testing.T) {
	const dbname = "writer_close_output"
	setupDB(dbname)
	s := sql.NewWriter(sql.WriterOptions{
		Driver:  ramsqlDriver,
		ConnStr: dbname,
		Table:   "members",
	})
	if err := workers.ATProcessCloseOutput(s); err != nil {
		t.Fatal(err)
	}
}

func TestWriterProcessCancel(t *testing.T) {
	const dbname = "writer_cancel"
	setupDB(dbname)
	s := sql.NewWriter(sql.WriterOptions{
		Driver:  ramsqlDriver,
		ConnStr: dbname,
		Table:   "members",
	})
	if err := workers.ATProcessCancel(s); err != nil {
		t.Fatal(err)
	}
}
