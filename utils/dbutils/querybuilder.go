package dbutils

import (
	"strings"
	"sync"
)

type QueryBuilder struct {
	mu       sync.Mutex
	args     []any
	criteria []string
	groupBy  string
	orderBy  []string
	query    string
	limit    string
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

func (qb *QueryBuilder) reset() {
	qb.args = nil
	qb.criteria = nil
	qb.groupBy = ""
	qb.orderBy = nil
	qb.query = ""
	qb.limit = ""
}

func (qb *QueryBuilder) Query(query string) *QueryBuilder {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	qb.reset()
	qb.query = query
	return qb
}

func (qb *QueryBuilder) Where(query string, value any) *QueryBuilder {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	qb.criteria = append(qb.criteria, query)

	if value != nil {
		qb.args = append(qb.args, value)
	}

	return qb
}

func (qb *QueryBuilder) GroupBy(condition string) *QueryBuilder {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	qb.groupBy = "GROUP BY " + condition

	return qb
}

func (qb *QueryBuilder) OrderBy(column string, order string) *QueryBuilder {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	qb.orderBy = append(qb.orderBy, column+" "+order)

	return qb
}

func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	qb.limit = "LIMIT ?"

	qb.args = append(qb.args, limit)

	return qb
}

func (qb *QueryBuilder) GetQuery() (string, []any) {
	qb.mu.Lock()
	defer qb.mu.Unlock()

	query := qb.query

	if len(qb.criteria) > 0 {
		query = query + " WHERE " + strings.Join(qb.criteria, " AND ")
	}

	if qb.groupBy != "" {
		query = query + " " + qb.groupBy
	}

	if len(qb.orderBy) > 0 {
		query = query + " ORDER BY " + strings.Join(qb.orderBy, ", ")
	}

	if qb.limit != "" {
		query = query + " " + qb.limit
	}

	return query, qb.args
}
