package sql_test

import (
	"context"
	dbsql "database/sql"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/licaonfee/selina"

	"github.com/licaonfee/selina/workers"

	"github.com/licaonfee/selina/workers/sql"
	_ "github.com/proullon/ramsql/driver"
)

const sampleDB = `CREATE TABLE members(id INT, name STRING, mood STRING);
INSERT INTO members(id, name, mood) values (0, "selina", "cheerful");
INSERT INTO members(id, name, mood) values (1, "lizbeth", "liverish");`

const ramsqlDriver = "ramsql"

func setupDB(name string) {
	db, err := dbsql.Open(ramsqlDriver, name)

	if err != nil {
		panic(err)
	}
	for _, query := range strings.Split(sampleDB, "\n") {
		_, err := db.Exec(query)
		if err != nil {
			panic(err)
		}
	}
}

//ramsql return all data as slice of bytes ,
// and not implemets driver.RowsColumnTypeScanType
// so all values are scanned to interface{} and strings
// are stored as base64
func decode(data []string) []string {
	res := make([]string, len(data))
	for i, row := range data {
		val := make(map[string]string)
		if err := json.Unmarshal([]byte(row), &val); err != nil {
			panic(err)
		}
		for k, v := range val {
			nv, _ := base64.StdEncoding.DecodeString(v)
			val[k] = string(nv)
		}
		newdata, err := json.Marshal(val)
		if err != nil {
			panic(err)
		}
		res[i] = string(newdata)
	}
	return res
}
func TestSQLReader_Process(t *testing.T) {
	tests := []struct {
		name    string
		opts    sql.ReaderOptions
		want    []string
		wantErr bool
	}{
		{
			name:    "Unregistered Driver",
			opts:    sql.ReaderOptions{Driver: "unknow", ConnStr: "unknow", Query: ""},
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "Empty Query",
			opts:    sql.ReaderOptions{Driver: ramsqlDriver, ConnStr: "empty_query", Query: ""},
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "Invalid Query",
			opts:    sql.ReaderOptions{Driver: ramsqlDriver, ConnStr: "invalid_query", Query: "SE;"},
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "Success",
			opts:    sql.ReaderOptions{Driver: ramsqlDriver, ConnStr: "success", Query: "SELECT name FROM members;"},
			want:    []string{`{"name":"selina"}`, `{"name":"lizbeth"}`},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDB(tt.opts.ConnStr)
			s := sql.NewReader(tt.opts)
			output := make(chan []byte, len(tt.want)+1)
			args := selina.ProcessArgs{Input: nil, Output: output}
			err := s.Process(context.Background(), args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Process() err = %v , wantErr=%v", err, tt.wantErr)
			}
			got := decode(selina.ChannelAsSlice(output))
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Process() got = %#v , want = %#v", got, tt.want)
			}
		})
	}
}
func TestSQLReader_Process_close_input(t *testing.T) {
	const dbname = "reader_close_input"
	setupDB(dbname)
	s := sql.NewReader(sql.ReaderOptions{
		Driver:  ramsqlDriver,
		ConnStr: dbname,
		Query:   "SELECT name FROM members;",
	})
	if err := workers.ATProcessCloseInput(s); err != nil {
		t.Fatal(err)
	}
}

func TestSQLReader_Process_close_output(t *testing.T) {
	const dbname = "reader_close_output"
	setupDB(dbname)
	s := sql.NewReader(sql.ReaderOptions{
		Driver:  ramsqlDriver,
		ConnStr: dbname,
		Query:   "SELECT name FROM members;",
	})
	if err := workers.ATProcessCloseOutput(s); err != nil {
		t.Fatal(err)
	}
}

func TestSQLReader_Process_cancel(t *testing.T) {
	const dbname = "reader_cancel"
	setupDB(dbname)
	s := sql.NewReader(sql.ReaderOptions{
		Driver:  ramsqlDriver,
		ConnStr: dbname,
		Query:   "SELECT name FROM members;",
	})
	if err := workers.ATProcessCancel(s); err != nil {
		t.Fatal(err)
	}
}
