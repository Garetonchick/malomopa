package db

import (
	"fmt"
	"strings"
)

type cqlInsertQuery struct {
	tableName    string
	queryColumns []string
}

func fullTableName(keyspace, table string) string {
	return fmt.Sprintf("%s.%s", keyspace, table)
}

func newInsert(keyspace, table string) *cqlInsertQuery {
	return &cqlInsertQuery{
		tableName: fullTableName(keyspace, table),
	}
}

func (iq *cqlInsertQuery) columns(queryColumn ...string) *cqlInsertQuery {
	iq.queryColumns = append(iq.queryColumns, queryColumn...)
	return iq
}

func (iq *cqlInsertQuery) build() string {
	columnNames := fmt.Sprintf("(%s)", strings.Join(iq.queryColumns, ", "))

	questions := make([]string, len(iq.queryColumns))
	for i := 0; i < len(iq.queryColumns); i++ {
		questions[i] = "?"
	}
	columnValues := fmt.Sprintf("(%s)", strings.Join(questions, ", "))

	return fmt.Sprintf("INSERT INTO %s %s VALUES %s", iq.tableName, columnNames, columnValues)
}

type cqlUpdateQuery struct {
	tableName string
	setStmts  []string
	whereStmt string
	ifStmt    string
}

func newUpdate(keyspace, table string) *cqlUpdateQuery {
	return &cqlUpdateQuery{
		tableName: fullTableName(keyspace, table),
	}
}

func (uq *cqlUpdateQuery) set(stmt ...string) *cqlUpdateQuery {
	uq.setStmts = append(uq.setStmts, stmt...)
	return uq
}

func (uq *cqlUpdateQuery) where(whereStmt string) *cqlUpdateQuery {
	uq.whereStmt = whereStmt
	return uq
}

func (uq *cqlUpdateQuery) casIf(ifStmt string) *cqlUpdateQuery {
	uq.ifStmt = ifStmt
	return uq
}

func (uq *cqlUpdateQuery) build() string {
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s ",
		uq.tableName,
		strings.Join(uq.setStmts, ", "),
		uq.whereStmt,
	)

	if uq.ifStmt != "" {
		query += fmt.Sprintf("IF %s", uq.ifStmt)
	}

	return query
}

type cqlSelectQuery struct {
	tableName    string
	queryColumns []string
	whereStmt    string
	limitStmt    int64
}

func newSelect() *cqlSelectQuery {
	return &cqlSelectQuery{}
}

func (sq *cqlSelectQuery) from(keyspace, table string) *cqlSelectQuery {
	sq.tableName = fullTableName(keyspace, table)
	return sq
}

func (sq *cqlSelectQuery) columns(queryColumn ...string) *cqlSelectQuery {
	sq.queryColumns = append(sq.queryColumns, queryColumn...)
	return sq
}

func (sq *cqlSelectQuery) where(whereStmt string) *cqlSelectQuery {
	sq.whereStmt = whereStmt
	return sq
}

func (sq *cqlSelectQuery) limit(limitStmt int64) *cqlSelectQuery {
	sq.limitStmt = limitStmt
	return sq
}

func (sq *cqlSelectQuery) build() string {
	query := fmt.Sprintf(
		"SELECT %s FROM %s ",
		strings.Join(sq.queryColumns, ", "),
		sq.tableName,
	)

	if sq.whereStmt != "" {
		query += fmt.Sprintf("WHERE %s ", sq.whereStmt)
	}

	if sq.limitStmt != 0 {
		query += fmt.Sprintf("LIMIT %d ALLOW FILTERING", sq.limitStmt)
	}

	return query
}
