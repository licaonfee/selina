package sql

import (
	"strings"
)

//QueryBuilder define methods to build SQL queries for database operations
type QueryBuilder interface {
	//Insert create a new insert statement with given table and colums
	Insert(table string, cols []string) string
}

//DefaultQueryBuilder when no Builder is provided to Worker , this implementation is used
type DefaultQueryBuilder struct{}

//Insert implements QueryBuilder interface
//nolint gosec
func (d *DefaultQueryBuilder) Insert(table string, cols []string) string {
	query := "INSERT INTO `" + table + "`(" + strings.Join(cols, ",") + ")" +
		"VALUES(" + strings.Repeat("?,", len(cols))[:len(cols)*2-1] + ");"
	return query
}
