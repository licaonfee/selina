package sql_test

import (
	"testing"

	"github.com/licaonfee/selina/workers/sql"
)

var storeString string

func BenchmarkDefaultBuilder_Insert(b *testing.B) {
	builder := sql.DefaultQueryBuilder{}
	cols := []string{"column1", "column2", "column3", "column4"}
	for i := 0; i < b.N; i++ {
		storeString = builder.Insert("mytable", cols)
	}
}
