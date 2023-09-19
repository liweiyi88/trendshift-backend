package dbutils

import (
	"strings"
)

type QueryBuilder struct {
	args     []any
	criteria []string
	groupBy  string
	orderBy  string
	query    string
	limit    string
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

func (qb *QueryBuilder) Query(query string) *QueryBuilder {
	qb.query = query
	return qb
}

func (qb *QueryBuilder) Where(query string, value any) *QueryBuilder {
	qb.criteria = append(qb.criteria, query)

	if value != nil {
		qb.args = append(qb.args, value)
	}

	return qb
}

func (qb *QueryBuilder) GroupBy(condition string) *QueryBuilder {
	qb.groupBy = "GROUP BY " + condition

	return qb
}

func (qb *QueryBuilder) OrderBy(condition string, order string) *QueryBuilder {
	qb.orderBy = "ORDER BY " + condition

	order = strings.ToLower(order)
	if order == "asc" {
		qb.orderBy = qb.orderBy + " ASC"
	} else {
		qb.orderBy = qb.orderBy + " DESC"
	}

	return qb
}

func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = "LIMIT ?"

	qb.args = append(qb.args, limit)

	return qb
}

func (qb *QueryBuilder) GetQuery() (string, []any) {
	query := qb.query

	if len(qb.criteria) > 0 {
		query = query + " WHERE " + strings.Join(qb.criteria, " AND ")
	}

	if qb.groupBy != "" {
		query = query + " " + qb.groupBy
	}

	if qb.orderBy != "" {
		query = query + " " + qb.orderBy
	}

	if qb.limit != "" {
		query = query + " " + qb.limit
	}

	return query, qb.args
}
