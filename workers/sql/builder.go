package sql

import (
	"strings"
)

type QueryBuilder interface {
	Insert(table string, cols []string) string
}

type DefaultQueryBuilder struct{}

func (d *DefaultQueryBuilder) Insert(table string, cols []string) string {
	colfield := strings.Join(cols, ",")
	query := "INSERT INTO " + table + "(" + colfield + ")" +
		"VALUES(" + strings.Repeat("?,", len(cols))[:len(cols)*2-1] + ");"
	return query
}
