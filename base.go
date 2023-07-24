package gen

import (
	"context"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type BaseDo[T any] interface {
	SubQuery[T]
	Debug() BaseDo[T]
	WithContext(ctx context.Context) BaseDo[T]
	WithResult(fc func(tx Dao[T])) ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() BaseDo[T]
	WriteDB() BaseDo[T]
	As(alias string) Dao[T]
	Session(config *gorm.Session) BaseDo[T]
	Columns(cols ...field.Expr) Columns
	Clauses(conds ...clause.Expression) BaseDo[T]
	Not(conds ...Condition) BaseDo[T]
	Or(conds ...Condition) BaseDo[T]
	Select(conds ...field.Expr) BaseDo[T]
	Where(conds ...Condition) BaseDo[T]
	WhereStruct(get field.GetField, data any) BaseDo[T]
	Order(conds ...field.Expr) BaseDo[T]
	Distinct(cols ...field.Expr) BaseDo[T]
	Omit(cols ...field.Expr) BaseDo[T]
	Join(table schema.Tabler, on ...field.Expr) BaseDo[T]
	LeftJoin(table schema.Tabler, on ...field.Expr) BaseDo[T]
	RightJoin(table schema.Tabler, on ...field.Expr) BaseDo[T]
	Group(cols ...field.Expr) BaseDo[T]
	Having(conds ...Condition) BaseDo[T]
	Limit(limit int) BaseDo[T]
	Offset(offset int) BaseDo[T]
	Count() (count int64, err error)
	Scopes(funcs ...func(Dao[T]) Dao[T]) BaseDo[T]
	Unscoped() BaseDo[T]
	Create(values ...any) error
	CreateInBatches(values []*T, batchSize int) error
	Save(values ...*T) error
	First() (*T, error)
	Take() (*T, error)
	Last() (*T, error)
	Find() ([]*T, error)
	FindInBatch(batchSize int, fc func(tx Dao[T], batch int) error) (results []*T, err error)
	FindInBatches(result *[]*T, batchSize int, fc func(tx Dao[T], batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*T) (info ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info ResultInfo, err error)
	Updates(value interface{}) (info ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info ResultInfo, err error)
	UpdateColumns(value interface{}) (info ResultInfo, err error)
	UpdateFrom(q SubQuery[T]) Dao[T]
	Attrs(attrs ...field.AssignExpr) BaseDo[T]
	Assign(attrs ...field.AssignExpr) BaseDo[T]
	Joins(fields ...field.RelationField) BaseDo[T]
	Preload(fields ...field.RelationField) BaseDo[T]
	FirstOrInit() (*T, error)
	FirstOrCreate() (*T, error)
	FindByPage(offset int, limit int) (result []*T, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) BaseDo[T]
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

//func id() BaseDo[any] {
//	data := DO[any]{}
//	return &data
//}
