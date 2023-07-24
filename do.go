package gen

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen/field"
	"gorm.io/gen/helper"
)

// ResultInfo query/execute info
type ResultInfo struct {
	RowsAffected int64
	Error        error
}

//var _ Dao[T] = new(DO)

// DO (data object): implement basic query methods
// the structure embedded with a *gorm.DB, and has a element item "alias" will be used when used as a sub query
type DO[T any] struct {
	*DOConfig
	db        *gorm.DB
	alias     string // for subquery
	modelType reflect.Type
	tableName string

	backfillData interface{}
}

func (d DO[T]) getInstance(db *gorm.DB) *DO[T] {
	d.db = db
	return &d
}

type doOptions func(*gorm.DB) *gorm.DB

var (
	// Debug use DB in debug mode
	Debug doOptions = func(db *gorm.DB) *gorm.DB { return db.Debug() }
)

// UseDB specify a db connection(*gorm.DB)
func (d *DO[T]) UseDB(db *gorm.DB, opts ...DOOption[T]) {
	db = db.Session(&gorm.Session{Context: context.Background()})
	d.db = db
	config := &DOConfig{}
	for _, opt := range opts {
		if opt != nil {
			if applyErr := opt.Apply(config); applyErr != nil {
				panic(applyErr)
			}
		}
	}
	d.DOConfig = config
}

// ReplaceDB replace db connection
func (d *DO[T]) ReplaceDB(db *gorm.DB) {
	d.db = db.Session(&gorm.Session{})
}

// ReplaceConnPool replace db connection pool
func (d *DO[T]) ReplaceConnPool(pool gorm.ConnPool) {
	d.db = d.db.Session(&gorm.Session{Initialized: true}).Session(&gorm.Session{})
	d.db.Statement.ConnPool = pool
}

// UseModel specify a data model structure as a source for table name
func (d *DO[T]) UseModel() {
	var model = d.indirect(new(T))
	err := d.db.Statement.Parse(model)
	if err != nil {
		panic(fmt.Errorf("Cannot parse model: %+v\n%w", model, err))
	}
	d.tableName = d.db.Statement.Schema.Table
}

func (d *DO[T]) indirect(value interface{}) reflect.Type {
	mt := reflect.TypeOf(value)
	if mt.Kind() == reflect.Ptr {
		mt = mt.Elem()
	}
	return mt
}

// UseTable specify table name
func (d *DO[T]) UseTable(tableName string) {
	d.db = d.db.Table(tableName).Session(new(gorm.Session))
	//d.db.Statement.Schema.Table=tableName
	d.tableName = tableName
}

// TableName return table name
func (d DO[T]) TableName() string {
	return d.tableName
}

// Returning backfill data
func (d DO[T]) Returning(value interface{}, columns ...string) Dao[T] {
	d.backfillData = value

	var targetCulumns []clause.Column
	for _, column := range columns {
		targetCulumns = append(targetCulumns, clause.Column{Name: column})
	}
	d.db = d.db.Clauses(clause.Returning{Columns: targetCulumns})
	return &d
}

// Session replace db with new session
func (d *DO[T]) Session(config *gorm.Session) Dao[T] { return d.getInstance(d.db.Session(config)) }

// UnderlyingDB return the underlying database connection
func (d *DO[T]) UnderlyingDB() *gorm.DB { return d.underlyingDB() }

// Quote return qutoed data
func (d *DO[T]) Quote(raw string) string { return d.db.Statement.Quote(raw) }

// Build implement the interface of claues.Expression
// only call WHERE clause's Build
func (d *DO[T]) Build(builder clause.Builder) {
	for _, e := range d.buildCondition() {
		e.Build(builder)
	}
}

func (d *DO[T]) buildCondition() []clause.Expression {
	return d.db.Statement.BuildCondition(d.db)
}

// underlyingDO return self
func (d *DO[T]) underlyingDO() *DO[T]  { return d }
func (d *DO[T]) UnderlyingDO2() *DO[T] { return d }

// underlyingDB return self.db
func (d *DO[T]) underlyingDB() *gorm.DB  { return d.db }
func (d *DO[T]) UnderlyingDB2() *gorm.DB { return d.db }

func (d *DO[T]) withError(err error) *DO[T] {
	if err == nil {
		return d
	}

	newDB := d.db.Session(new(gorm.Session))
	_ = newDB.AddError(err)
	return d.getInstance(newDB)
}

// BeCond implements Condition
func (d *DO[T]) BeCond() interface{} { return d.buildCondition() }

// CondError implements Condition
func (d *DO[T]) CondError() error { return nil }

// Debug return a DO with db in debug mode
func (d *DO[T]) Debug() Dao[T] { return d.getInstance(d.db.Debug()) }

// WithContext return a DO with db with context
func (d *DO[T]) WithContext(ctx context.Context) Dao[T] { return d.getInstance(d.db.WithContext(ctx)) }

// Clauses specify Clauses
func (d *DO[T]) Clauses(conds ...clause.Expression) Dao[T] {
	if err := checkConds(conds); err != nil {
		newDB := d.db.Session(new(gorm.Session))
		_ = newDB.AddError(err)
		return d.getInstance(newDB)
	}
	return d.getInstance(d.db.Clauses(conds...))
}

// As alias cannot be heired, As must used on tail
func (d DO[T]) As(alias string) Dao[T] {
	d.alias = alias
	d.db = d.db.Table(fmt.Sprintf("%s AS %s", d.Quote(d.TableName()), d.Quote(alias)))
	return &d
}

// Alias return alias name
func (d *DO[T]) Alias() string { return d.alias }

// Columns return columns for Subquery
func (*DO[T]) Columns(cols ...field.Expr) Columns { return cols }

// ======================== chainable api ========================

// Not ...
func (d *DO[T]) Not(conds ...Condition) Dao[T] {
	exprs, err := condToExpression(conds)
	if err != nil {
		return d.withError(err)
	}
	if len(exprs) == 0 {
		return d
	}
	return d.getInstance(d.db.Clauses(clause.Where{Exprs: []clause.Expression{clause.Not(exprs...)}}))
}

// Or ...
func (d *DO[T]) Or(conds ...Condition) Dao[T] {
	exprs, err := condToExpression(conds)
	if err != nil {
		return d.withError(err)
	}
	if len(exprs) == 0 {
		return d
	}
	return d.getInstance(d.db.Clauses(clause.Where{Exprs: []clause.Expression{clause.Or(clause.And(exprs...))}}))
}

// Select ...
func (d *DO[T]) Select(columns ...field.Expr) Dao[T] {
	if len(columns) == 0 {
		return d.getInstance(d.db.Clauses(clause.Select{}))
	}
	query, args := buildExpr4Select(d.db.Statement, columns...)
	return d.getInstance(d.db.Select(query, args...))
}

// Where ...
func (d *DO[T]) Where(conds ...Condition) Dao[T] {
	exprs, err := condToExpression(conds)
	if err != nil {
		return d.withError(err)
	}
	if len(exprs) == 0 {
		return d
	}
	return d.getInstance(d.db.Clauses(clause.Where{Exprs: exprs}))
}

// Order ...
func (d *DO[T]) Order(columns ...field.Expr) Dao[T] {
	// lazy build Columns
	// if c, ok := d.db.Statement.Clauses[clause.OrderBy{}.Name()]; ok {
	// 	if order, ok := c.Expression.(clause.OrderBy); ok {
	// 		if expr, ok := order.Expression.(clause.CommaExpression); ok {
	// 			expr.Exprs = append(expr.Exprs, toExpression(columns)...)
	// 			return d.newInstance(d.db.Clauses(clause.OrderBy{Expression: expr}))
	// 		}
	// 	}
	// }
	// return d.newInstance(d.db.Clauses(clause.OrderBy{Expression: clause.CommaExpression{Exprs: toExpression(columns)}}))
	if len(columns) == 0 {
		return d
	}
	return d.getInstance(d.db.Order(d.toOrderValue(columns...)))
}

func (d *DO[T]) toOrderValue(columns ...field.Expr) string {
	// eager build Columns
	stmt := &gorm.Statement{DB: d.db.Statement.DB, Table: d.db.Statement.Table, Schema: d.db.Statement.Schema}

	for i, c := range columns {
		if i != 0 {
			stmt.WriteByte(',')
		}
		c.Build(stmt)
	}

	return stmt.SQL.String()
}

// Distinct ...
func (d *DO[T]) Distinct(columns ...field.Expr) Dao[T] {
	return d.getInstance(d.db.Distinct(toInterfaceSlice(toColExprFullName(d.db.Statement, columns...))...))
}

// Omit ...
func (d *DO[T]) Omit(columns ...field.Expr) Dao[T] {
	if len(columns) == 0 {
		return d
	}
	return d.getInstance(d.db.Omit(getColumnName(columns...)...))
}

// Group ...
func (d *DO[T]) Group(columns ...field.Expr) Dao[T] {
	if len(columns) == 0 {
		return d
	}

	stmt := &gorm.Statement{DB: d.db.Statement.DB, Table: d.db.Statement.Table, Schema: d.db.Statement.Schema}

	for i, c := range columns {
		if i != 0 {
			stmt.WriteByte(',')
		}
		c.Build(stmt)
	}

	return d.getInstance(d.db.Group(stmt.SQL.String()))
}

// Having ...
func (d *DO[T]) Having(conds ...Condition) Dao[T] {
	exprs, err := condToExpression(conds)
	if err != nil {
		return d.withError(err)
	}
	if len(exprs) == 0 {
		return d
	}
	return d.getInstance(d.db.Clauses(clause.GroupBy{Having: exprs}))
}

// Limit ...
func (d *DO[T]) Limit(limit int) Dao[T] {
	return d.getInstance(d.db.Limit(limit))
}

// Offset ...
func (d *DO[T]) Offset(offset int) Dao[T] {
	return d.getInstance(d.db.Offset(offset))
}

// Scopes ...
func (d *DO[T]) Scopes(funcs ...func(Dao[T]) Dao[T]) Dao[T] {
	fcs := make([]func(*gorm.DB) *gorm.DB, len(funcs))
	for i, f := range funcs {
		sf := f
		fcs[i] = func(tx *gorm.DB) *gorm.DB { return sf(d.getInstance(tx)).(*DO[T]).db }
	}
	return d.getInstance(d.db.Scopes(fcs...))
}

// Unscoped ...
func (d *DO[T]) Unscoped() Dao[T] {
	return d.getInstance(d.db.Unscoped())
}

// Join ...
func (d *DO[T]) Join(table schema.Tabler, conds ...field.Expr) Dao[T] {
	return d.join(table, clause.InnerJoin, conds)
}

// LeftJoin ...
func (d *DO[T]) LeftJoin(table schema.Tabler, conds ...field.Expr) Dao[T] {
	return d.join(table, clause.LeftJoin, conds)
}

// RightJoin ...
func (d *DO[T]) RightJoin(table schema.Tabler, conds ...field.Expr) Dao[T] {
	return d.join(table, clause.RightJoin, conds)
}

func (d *DO[T]) join(table schema.Tabler, joinType clause.JoinType, conds []field.Expr) Dao[T] {
	if len(conds) == 0 {
		return d.withError(ErrEmptyCondition)
	}

	join := clause.Join{
		Type:  joinType,
		Table: clause.Table{Name: table.TableName()},
		ON:    clause.Where{Exprs: toExpression(conds...)},
	}
	if do, ok := table.(Dao[T]); ok {
		join.Expression = helper.NewJoinTblExpr(join, Table[T](do).underlyingDB().Statement.TableExpr)
	}
	if al, ok := table.(interface{ Alias() string }); ok {
		join.Table.Alias = al.Alias()
	}

	from := getFromClause(d.db)
	from.Joins = append(from.Joins, join)
	return d.getInstance(d.db.Clauses(from))
}

// Attrs ...
func (d *DO[T]) Attrs(attrs ...field.AssignExpr) Dao[T] {
	if len(attrs) == 0 {
		return d
	}
	return d.getInstance(d.db.Attrs(d.attrsValue(attrs)...))
}

// Assign ...
func (d *DO[T]) Assign(attrs ...field.AssignExpr) Dao[T] {
	if len(attrs) == 0 {
		return d
	}
	return d.getInstance(d.db.Assign(d.attrsValue(attrs)...))
}

func (d *DO[T]) attrsValue(attrs []field.AssignExpr) []interface{} {
	values := make([]interface{}, 0, len(attrs))
	for _, attr := range attrs {
		if expr, ok := attr.AssignExpr().(clause.Eq); ok {
			values = append(values, expr)
		}
	}
	return values
}

// Joins ...
func (d *DO[T]) Joins(field field.RelationField) Dao[T] {
	var args []interface{}

	if conds := field.GetConds(); len(conds) > 0 {
		var exprs []clause.Expression
		for _, oe := range toExpression(conds...) {
			switch e := oe.(type) {
			case clause.Eq:
				if c, ok := e.Column.(clause.Column); ok {
					c.Table = field.Name()
					e.Column = c
				}
				exprs = append(exprs, e)
			case clause.Neq:
				if c, ok := e.Column.(clause.Column); ok {
					c.Table = field.Name()
					e.Column = c
				}
				exprs = append(exprs, e)
			case clause.Gt:
				if c, ok := e.Column.(clause.Column); ok {
					c.Table = field.Name()
					e.Column = c
				}
				exprs = append(exprs, e)
			case clause.Gte:
				if c, ok := e.Column.(clause.Column); ok {
					c.Table = field.Name()
					e.Column = c
				}
				exprs = append(exprs, e)
			case clause.Lt:
				if c, ok := e.Column.(clause.Column); ok {
					c.Table = field.Name()
					e.Column = c
				}
				exprs = append(exprs, e)
			case clause.Lte:
				if c, ok := e.Column.(clause.Column); ok {
					c.Table = field.Name()
					e.Column = c
				}
				exprs = append(exprs, e)
			case clause.Like:
				if c, ok := e.Column.(clause.Column); ok {
					c.Table = field.Name()
					e.Column = c
				}
				exprs = append(exprs, e)
			}
		}

		args = append(args, d.db.Clauses(clause.Where{
			Exprs: exprs,
		}))
	}
	if columns := field.GetSelects(); len(columns) > 0 {
		colNames := make([]string, len(columns))
		for i, c := range columns {
			colNames[i] = string(c.ColumnName())
		}
		args = append(args, func(db *gorm.DB) *gorm.DB {
			return db.Select(colNames)
		})
	}
	if columns := field.GetOrderCol(); len(columns) > 0 {
		var os []string
		for _, oe := range columns {
			switch e := oe.RawExpr().(type) {
			case clause.Expr:
				vs := []interface{}{}
				for _, v := range e.Vars {
					if c, ok := v.(clause.Column); ok {
						vs = append(vs, clause.Column{
							Table: field.Name(),
							Name:  c.Name,
							Alias: c.Alias,
							Raw:   c.Raw,
						})
					}
				}
				e.Vars = vs
				newStmt := &gorm.Statement{DB: d.db.Statement.DB, Table: d.db.Statement.Table, Schema: d.db.Statement.Schema}
				e.Build(newStmt)
				os = append(os, newStmt.SQL.String())
			}
		}
		args = append(args, d.db.Order(strings.Join(os, ",")))
	}
	if clauses := field.GetClauses(); len(clauses) > 0 {
		args = append(args, func(db *gorm.DB) *gorm.DB {
			return db.Clauses(clauses...)
		})
	}
	if funcs := field.GetScopes(); len(funcs) > 0 {
		for _, f := range funcs {
			args = append(args, (func(*gorm.DB) *gorm.DB)(f))
		}
	}
	if offset, limit := field.GetPage(); offset|limit != 0 {
		args = append(args, func(db *gorm.DB) *gorm.DB {
			return db.Offset(offset).Limit(limit)
		})
	}

	return d.getInstance(d.db.Joins(field.Path(), args...))
}

// Preload ...
func (d *DO[T]) Preload(field field.RelationField) Dao[T] {
	var args []interface{}
	if conds := field.GetConds(); len(conds) > 0 {
		args = append(args, toExpressionInterface(conds...)...)
	}
	if columns := field.GetSelects(); len(columns) > 0 {
		colNames := make([]string, len(columns))
		for i, c := range columns {
			colNames[i] = string(c.ColumnName())
		}
		args = append(args, func(db *gorm.DB) *gorm.DB {
			return db.Select(colNames)
		})
	}
	if columns := field.GetOrderCol(); len(columns) > 0 {
		args = append(args, func(db *gorm.DB) *gorm.DB {
			return db.Order(d.toOrderValue(columns...))
		})
	}
	if clauses := field.GetClauses(); len(clauses) > 0 {
		args = append(args, func(db *gorm.DB) *gorm.DB {
			return db.Clauses(clauses...)
		})
	}
	if funcs := field.GetScopes(); len(funcs) > 0 {
		for _, f := range funcs {
			args = append(args, (func(*gorm.DB) *gorm.DB)(f))
		}
	}
	if offset, limit := field.GetPage(); offset|limit != 0 {
		args = append(args, func(db *gorm.DB) *gorm.DB {
			return db.Offset(offset).Limit(limit)
		})
	}
	return d.getInstance(d.db.Preload(field.Path(), args...))
}

// UpdateFrom specify update sub query
func (d *DO[T]) UpdateFrom(q SubQuery[T]) Dao[T] {
	var tableName strings.Builder
	d.db.Statement.QuoteTo(&tableName, d.TableName())
	if d.alias != "" {
		tableName.WriteString(" AS ")
		d.db.Statement.QuoteTo(&tableName, d.alias)
	}

	tableName.WriteByte(',')
	if _, ok := q.underlyingDB().Statement.Clauses["SELECT"]; ok || len(q.underlyingDB().Statement.Selects) > 0 {
		tableName.WriteString("(" + q.underlyingDB().ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Table(q.underlyingDO().TableName()).Find(nil) }) + ")")
	} else {
		d.db.Statement.QuoteTo(&tableName, q.underlyingDO().TableName())
	}
	if alias := q.underlyingDO().alias; alias != "" {
		tableName.WriteString(" AS ")
		d.db.Statement.QuoteTo(&tableName, alias)
	}

	return d.getInstance(d.db.Clauses(clause.Update{Table: clause.Table{Name: tableName.String(), Raw: true}}))
}

func getFromClause(db *gorm.DB) *clause.From {
	if db == nil || db.Statement == nil {
		return &clause.From{}
	}
	c, ok := db.Statement.Clauses[clause.From{}.Name()]
	if !ok || c.Expression == nil {
		return &clause.From{}
	}
	from, ok := c.Expression.(clause.From)
	if !ok {
		return &clause.From{}
	}
	return &from
}

// ======================== finisher api ========================

// Create ...
func (d *DO[T]) Create(value ...*T) error {
	return d.db.Create(value).Error
}

// CreateInBatches ...
func (d *DO[T]) CreateInBatches(value []*T, batchSize int) error {
	return d.db.CreateInBatches(value, batchSize).Error
}

// Save ...
func (d *DO[T]) Save(value ...*T) error {
	return d.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(value).Error
}

// First ...
func (d *DO[T]) First() (result *T, err error) {
	return d.singleQuery(d.db.First)
}

// Take ...
func (d *DO[T]) Take() (result *T, err error) {
	return d.singleQuery(d.db.Take)
}

// Last ...
func (d *DO[T]) Last() (result *T, err error) {
	return d.singleQuery(d.db.Last)
}

func (d *DO[T]) singleQuery(query func(dest interface{}, conds ...interface{}) *gorm.DB) (result *T, err error) {
	result = new(T)
	if err = query(result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

// Find ...
func (d *DO[T]) Find() (results []*T, err error) {
	return d.multiQuery(d.db.Find)
}

func (d *DO[T]) multiQuery(query func(dest interface{}, conds ...interface{}) *gorm.DB) (results []*T, err error) {
	results = make([]*T, 0)
	err = query(&results).Error
	return results, err
}

func (d *DO[T]) findToMap() (interface{}, error) {
	var results []map[string]interface{}
	err := d.db.Find(&results).Error
	return results, err
}

// FindInBatches ...
func (d *DO[T]) FindInBatches(dest interface{}, batchSize int, fc func(tx Dao[T], batch int) error) error {
	return d.db.FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error { return fc(d.getInstance(tx), batch) }).Error
}

// FirstOrInit ...
func (d *DO[T]) FirstOrInit() (result *T, err error) {
	return d.singleQuery(d.db.FirstOrInit)
}

// FirstOrCreate ...
func (d *DO[T]) FirstOrCreate() (result *T, err error) {
	return d.singleQuery(d.db.FirstOrCreate)
}

// Update ...
func (d *DO[T]) Update(column field.Expr, value interface{}) (info ResultInfo, err error) {
	tx := d.db.Model(d.newResultPointer())
	columnStr := column.BuildColumn(d.db.Statement, field.WithoutQuote).String()

	var result *gorm.DB
	switch value := value.(type) {
	case field.AssignExpr:
		result = tx.Update(columnStr, value.AssignExpr())
	case SubQuery[T]:
		result = tx.Update(columnStr, value.underlyingDB())
	default:
		result = tx.Update(columnStr, value)
	}
	return ResultInfo{RowsAffected: result.RowsAffected, Error: result.Error}, result.Error
}

// UpdateSimple ...
func (d *DO[T]) UpdateSimple(columns ...field.AssignExpr) (info ResultInfo, err error) {
	if len(columns) == 0 {
		return
	}

	result := d.db.Model(d.newResultPointer()).Clauses(d.assignSet(columns)).Omit("*").Updates(map[string]interface{}{})
	return ResultInfo{RowsAffected: result.RowsAffected, Error: result.Error}, result.Error
}

// Updates ...
func (d *DO[T]) Updates(value interface{}) (info ResultInfo, err error) {
	var rawTyp, valTyp reflect.Type

	rawTyp = reflect.TypeOf(value)
	if rawTyp.Kind() == reflect.Ptr {
		valTyp = rawTyp.Elem()
	} else {
		valTyp = rawTyp
	}

	tx := d.db
	if d.backfillData != nil {
		tx = tx.Model(d.backfillData)
	}
	switch {
	case valTyp != d.modelType: // different type with model
		if d.backfillData == nil {
			tx = tx.Model(d.newResultPointer())
		}
	case rawTyp.Kind() == reflect.Ptr: // ignore ptr value
	default: // for fixing "reflect.Value.Addr of unaddressable value" panic
		ptr := reflect.New(d.modelType)
		ptr.Elem().Set(reflect.ValueOf(value))
		value = ptr.Interface()
	}
	result := tx.Updates(value)
	return ResultInfo{RowsAffected: result.RowsAffected, Error: result.Error}, result.Error
}

// UpdateColumn ...
func (d *DO[T]) UpdateColumn(column field.Expr, value interface{}) (info ResultInfo, err error) {
	tx := d.db.Model(d.newResultPointer())
	columnStr := column.BuildColumn(d.db.Statement, field.WithoutQuote).String()

	var result *gorm.DB
	switch value := value.(type) {
	case field.Expr:
		result = tx.UpdateColumn(columnStr, value.RawExpr())
	case SubQuery[T]:
		result = d.db.UpdateColumn(columnStr, value.underlyingDB())
	default:
		result = d.db.UpdateColumn(columnStr, value)
	}
	return ResultInfo{RowsAffected: result.RowsAffected, Error: result.Error}, result.Error
}

// UpdateColumnSimple ...
func (d *DO[T]) UpdateColumnSimple(columns ...field.AssignExpr) (info ResultInfo, err error) {
	if len(columns) == 0 {
		return
	}

	result := d.db.Model(d.newResultPointer()).Clauses(d.assignSet(columns)).Omit("*").UpdateColumns(map[string]interface{}{})
	return ResultInfo{RowsAffected: result.RowsAffected, Error: result.Error}, result.Error
}

// UpdateColumns ...
func (d *DO[T]) UpdateColumns(value interface{}) (info ResultInfo, err error) {
	result := d.db.Model(d.newResultPointer()).UpdateColumns(value)
	return ResultInfo{RowsAffected: result.RowsAffected, Error: result.Error}, result.Error
}

// assignSet fetch all set
func (d *DO[T]) assignSet(exprs []field.AssignExpr) (set clause.Set) {
	for _, expr := range exprs {
		column := clause.Column{Table: d.alias, Name: string(expr.ColumnName())}
		switch e := expr.AssignExpr().(type) {
		case clause.Expr:
			set = append(set, clause.Assignment{Column: column, Value: e})
		case clause.Eq:
			set = append(set, clause.Assignment{Column: column, Value: e.Value})
		case clause.Set:
			set = append(set, e...)
		}
	}

	stmt := d.db.Session(&gorm.Session{}).Statement
	stmt.Dest = map[string]interface{}{}
	return append(set, callbacks.ConvertToAssignments(stmt)...)
}

// Delete ...
func (d *DO[T]) Delete(models ...*T) (info ResultInfo, err error) {
	var result *gorm.DB
	if len(models) == 0 || reflect.ValueOf(models[0]).Len() == 0 {
		result = d.db.Model(d.newResultPointer()).Delete(reflect.New(d.modelType).Interface())
	} else {
		targets := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(d.modelType)), 0, len(models))
		value := reflect.ValueOf(models[0])
		for i := 0; i < value.Len(); i++ {
			targets = reflect.Append(targets, value.Index(i))
		}
		result = d.db.Delete(targets.Interface())
	}
	return ResultInfo{RowsAffected: result.RowsAffected, Error: result.Error}, result.Error
}

// Count ...
func (d *DO[T]) Count() (count int64, err error) {
	return count, d.db.Session(&gorm.Session{}).Model(d.newResultPointer()).Count(&count).Error
}

// Row ...
func (d *DO[T]) Row() *sql.Row {
	return d.db.Model(d.newResultPointer()).Row()
}

// Rows ...
func (d *DO[T]) Rows() (*sql.Rows, error) {
	return d.db.Model(d.newResultPointer()).Rows()
}

// Scan ...
func (d *DO[T]) Scan(dest interface{}) error {
	return d.db.Model(d.newResultPointer()).Scan(dest).Error
}

// Pluck ...
func (d *DO[T]) Pluck(column field.Expr, dest interface{}) error {
	return d.db.Model(d.newResultPointer()).Pluck(column.ColumnName().String(), dest).Error
}

// ScanRows ...
func (d *DO[T]) ScanRows(rows *sql.Rows, dest interface{}) error {
	return d.db.Model(d.newResultPointer()).ScanRows(rows, dest)
}

// WithResult ...
func (d DO[T]) WithResult(fc func(tx Dao[T])) ResultInfo {
	d.db = d.db.Set("", "")
	fc(&d)
	return ResultInfo{RowsAffected: d.db.RowsAffected, Error: d.db.Error}
}

func (d *DO[T]) newResultPointer() interface{} {
	if d.backfillData != nil {
		return d.backfillData
	}
	if d.modelType == nil {
		return nil
	}
	return reflect.New(d.modelType).Interface()
}

func (d *DO[T]) newResultSlicePointer() interface{} {
	return reflect.New(reflect.SliceOf(reflect.PtrTo(d.modelType))).Interface()
}

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

		do := query.underlyingDO()
		// ignore alias, or will misuse with sub query alias
		tableExprs[i] = do.db.Table(do.TableName())
		if do.alias != "" {
			tablePlaceholder[i] += " AS " + do.Quote(do.alias)
		}
	}

	return &DO[T]{
		db: subQueries[0].underlyingDO().db.Session(&gorm.Session{NewDB: true}).Table(strings.Join(tablePlaceholder, ", "), tableExprs...),
	}
}

// ======================== sub query method ========================

// Columns columns array
type Columns []field.Expr

// Set assign value by subquery
func (cs Columns) Set(query SubQuery[T]) field.AssignExpr {
	return field.AssignSubQuery(cs, query.underlyingDB())
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
		return field.ContainsSubQuery(cs, query.underlyingDB())
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
	return field.CompareSubQuery(field.EqOp, cs[0], query.underlyingDB())
}

// Neq ...
func (cs Columns) Neq(query SubQuery[T]) field.Expr {
	if len(cs) == 0 {
		return field.EmptyExpr()
	}
	return field.CompareSubQuery(field.NeqOp, cs[0], query.underlyingDB())
}

// Gt ...
func (cs Columns) Gt(query SubQuery[T]) field.Expr {
	if len(cs) == 0 {
		return field.EmptyExpr()
	}
	return field.CompareSubQuery(field.GtOp, cs[0], query.underlyingDB())
}

// Gte ...
func (cs Columns) Gte(query SubQuery[T]) field.Expr {
	if len(cs) == 0 {
		return field.EmptyExpr()
	}
	return field.CompareSubQuery(field.GteOp, cs[0], query.underlyingDB())
}

// Lt ...
func (cs Columns) Lt(query SubQuery[T]) field.Expr {
	if len(cs) == 0 {
		return field.EmptyExpr()
	}
	return field.CompareSubQuery(field.LtOp, cs[0], query.underlyingDB())
}

// Lte ...
func (cs Columns) Lte(query SubQuery[T]) field.Expr {
	if len(cs) == 0 {
		return field.EmptyExpr()
	}
	return field.CompareSubQuery(field.LteOp, cs[0], query.underlyingDB())
}
