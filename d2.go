package gen

import (
	"strings"

	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func toColExprFullName(stmt *gorm.Statement, columns ...field.Expr) []string {
	return buildColExpr(stmt, columns, field.WithAll)
}

func getColumnName(columns ...field.Expr) (result []string) {
	for _, c := range columns {
		result = append(result, c.ColumnName().String())
	}
	return result
}

func buildColExpr(stmt *gorm.Statement, cols []field.Expr, opts ...field.BuildOpt) []string {
	results := make([]string, len(cols))
	for i, c := range cols {
		switch c.RawExpr().(type) {
		case clause.Column:
			results[i] = c.BuildColumn(stmt, opts...).String()
		case clause.Expression:
			sql, args := c.BuildWithArgs(stmt)
			results[i] = stmt.Dialector.Explain(sql.String(), args...)
		}
	}
	return results
}

func buildExpr4Select(stmt *gorm.Statement, exprs ...field.Expr) (query string, args []interface{}) {
	if len(exprs) == 0 {
		return "", nil
	}

	var queryItems []string
	for _, e := range exprs {
		sql, vars := e.BuildWithArgs(stmt)
		queryItems = append(queryItems, sql.String())
		args = append(args, vars...)
	}
	if len(args) == 0 {
		return queryItems[0], toInterfaceSlice(queryItems[1:])
	}
	return strings.Join(queryItems, ","), args
}

func toExpression(exprs ...field.Expr) []clause.Expression {
	result := make([]clause.Expression, len(exprs))
	for i, e := range exprs {
		result[i] = singleExpr(e)
	}
	return result
}

func toExpressionInterface(exprs ...field.Expr) []interface{} {
	result := make([]interface{}, len(exprs))
	for i, e := range exprs {
		result[i] = singleExpr(e)
	}
	return result
}

func singleExpr(e field.Expr) clause.Expression {
	switch v := e.RawExpr().(type) {
	case clause.Expression:
		return v
	case clause.Column:
		return clause.NamedExpr{SQL: "?", Vars: []interface{}{v}}
	default:
		return clause.Expr{}
	}
}

func toInterfaceSlice(value interface{}) []interface{} {
	switch v := value.(type) {
	case string:
		return []interface{}{v}
	case []string:
		res := make([]interface{}, len(v))
		for i, item := range v {
			res[i] = item
		}
		return res
	case []clause.Column:
		res := make([]interface{}, len(v))
		for i, item := range v {
			res[i] = item
		}
		return res
	default:
		return nil
	}
}

// ======================== New Table ========================

// Table return a new table produced by subquery,
// the return value has to be used as root node
//
//	Table(u.Select(u.ID, u.Name).Where(u.Age.Gt(18))).Select()
//
// the above usage is equivalent to SQL statement:
//
//	SELECT * FROM (SELECT `id`, `name` FROM `users_info` WHERE `age` > ?)"
func Table[T any](subQueries ...SubQuery[T]) Dao[T] {
	if len(subQueries) == 0 {
		return &DO[T]{}
	}
	tablePlaceholder := make([]string, len(subQueries))
	tableExprs := make([]interface{}, len(subQueries))
	for i, query := range subQueries {
		tablePlaceholder[i] = "(?)"

		do := query.UnderlyingDO()
		// ignore alias, or will misuse with sub query alias
		tableExprs[i] = do.db.Table(do.TableName())
		if do.alias != "" {
			tablePlaceholder[i] += " AS " + do.Quote(do.alias)
		}
	}

	return &DO[T]{
		db: subQueries[0].UnderlyingDO().db.Session(&gorm.Session{NewDB: true}).Table(strings.Join(tablePlaceholder, ", "), tableExprs...),
	}
}

// ======================== sub query method ========================

// Columns columns array
type Columns []field.Expr

// Set assign value by subquery
func (cs Columns) Set(query SubQuery[T]) field.AssignExpr {
	return field.AssignSubQuery(cs, query.UnderlyingDB())
}

// In accept query or value
func (cs Columns) In(queryOrValue Condition) field.Expr {
	if len(cs) == 0 {
		return field.EmptyExpr()
	}

	switch query := queryOrValue.(type) {
	case field.Value:
		return field.ContainsValue(cs, query)
	case SubQuery[T]:
		return field.ContainsSubQuery(cs, query.UnderlyingDB())
	default:
		return field.EmptyExpr()
	}
}

// NotIn ...
func (cs Columns) NotIn(queryOrValue Condition) field.Expr {
	return field.Not(cs.In(queryOrValue))
}

// Eq ...
func (cs Columns) Eq(query SubQuery[T]) field.Expr {
	if len(cs) == 0 {
		return field.EmptyExpr()
	}
	return field.CompareSubQuery(field.EqOp, cs[0], query.UnderlyingDB())
}

// Neq ...
func (cs Columns) Neq(query SubQuery[T]) field.Expr {
	if len(cs) == 0 {
		return field.EmptyExpr()
	}
	return field.CompareSubQuery(field.NeqOp, cs[0], query.UnderlyingDB())
}

// Gt ...
func (cs Columns) Gt(query SubQuery[T]) field.Expr {
	if len(cs) == 0 {
		return field.EmptyExpr()
	}
	return field.CompareSubQuery(field.GtOp, cs[0], query.UnderlyingDB())
}

// Gte ...
func (cs Columns) Gte(query SubQuery[T]) field.Expr {
	if len(cs) == 0 {
		return field.EmptyExpr()
	}
	return field.CompareSubQuery(field.GteOp, cs[0], query.UnderlyingDB())
}

// Lt ...
func (cs Columns) Lt(query SubQuery[T]) field.Expr {
	if len(cs) == 0 {
		return field.EmptyExpr()
	}
	return field.CompareSubQuery(field.LtOp, cs[0], query.UnderlyingDB())
}

// Lte ...
func (cs Columns) Lte(query SubQuery[T]) field.Expr {
	if len(cs) == 0 {
		return field.EmptyExpr()
	}
	return field.CompareSubQuery(field.LteOp, cs[0], query.UnderlyingDB())
}
