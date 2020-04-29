package sql

import (
	"strings"
)

type QueryBuilder interface {
	Insert(table string, cols []string) string
}

type DefaultQueryBuilder struct{}

func (d *DefaultQueryBuilder) Insert(table string, cols []string) string {
	query := "INSERT INTO `" + table + "`(" + strings.Join(cols, ",") + ")" +
		"VALUES(" + strings.Repeat("?,", len(cols))[:len(cols)*2-1] + ");"
	return query
}
