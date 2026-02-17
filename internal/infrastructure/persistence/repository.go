package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var st = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type id = uuid.UUID

type model any

type dto[ID id, M model] interface {
	Values() []any
	GetID() ID
	ToModel() M
	FromModel(m M) any
}

type Repository[ID id, M model, T dto[ID, M]] struct {
	db         *sqlx.DB
	tableName  string
	columns    *Columns
	trxManager *trmsqlx.CtxGetter
}

func NewRepository[ID id, M model, T dto[ID, M]](db *sqlx.DB, tableName string, columns *Columns) *Repository[ID, M, T] {
	return &Repository[ID, M, T]{
		db:         db,
		tableName:  tableName,
		columns:    columns,
		trxManager: trmsqlx.DefaultCtxGetter,
	}
}

func (r *Repository[ID, M, T]) GetByID(ctx context.Context, id ID) (M, error) {
	var model M
	var entity T

	query, args, err := st.Select("*").
		From(r.tableName).
		Where(sq.Eq{r.columns.primaryKey: id}).
		ToSql()

	if err != nil {
		return model, fmt.Errorf("build sql query: %w", err)
	}

	err = r.trxManager.DefaultTrOrDB(ctx, r.db).
		GetContext(ctx, &entity, query, args...)

	if errors.Is(err, sql.ErrNoRows) {
		return model, errors.New("not found")
	}

	if err != nil {
		return model, fmt.Errorf("get entity: %w", err)
	}

	model = entity.ToModel()

	return model, nil
}

func (r *Repository[ID, M, T]) Save(ctx context.Context, model M) error {
	return r.SaveMany(ctx, []M{model})
}

func (r *Repository[ID, M, T]) SaveMany(ctx context.Context, models []M) error {
	if len(models) == 0 {
		return nil
	}

	entities := make([]T, len(models))
	var entity T
	for i, model := range models {
		temp := entity.FromModel(model)
		entity, ok := temp.(T)
		if !ok {
			return errors.New("type assertion to dto failed")
		}
		entities[i] = entity
	}

	tr := r.trxManager.DefaultTrOrDB(ctx, r.db)
	updateQuery, updateArgs, err := r.getUpdateQuery(entities)
	if err != nil {
		return err
	}

	_, err = tr.ExecContext(ctx, updateQuery, updateArgs...)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository[ID, M, T]) getUpdateQuery(entities []T) (string, []interface{}, error) {
	builder := st.Insert(r.tableName).
		Columns(append([]string{r.columns.primaryKey}, r.columns.ForInsert()...)...)

	for _, e := range entities {
		values, err := r.getValuesForEntity(e)
		if err != nil {
			return "", nil, err
		}

		builder = builder.Values(values...)
	}

	suffix := fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s", r.columns.primaryKey, r.columns.GetOnConflictStatement())
	return builder.Suffix(suffix).ToSql()
}

func (r *Repository[ID, M, T]) getValuesForEntity(entity T) ([]interface{}, error) {
	values := entity.Values()

	entityID := entity.GetID()
	isDef, err := r.isDefaultID(entityID)
	if err != nil {
		return nil, err
	}

	if isDef {
		return append([]interface{}{sq.Expr("DEFAULT")}, values...), nil
	}
	return append([]interface{}{entityID}, values...), nil
}

func (r *Repository[ID, M, T]) isDefaultID(id any) (bool, error) {
	switch value := id.(type) {
	case int64:
		return value == 0, nil
	case uuid.UUID:
		return value == uuid.Nil, nil
	default:
		return false, errors.New("unknown ID type")
	}
}
