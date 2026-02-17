package order

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	domain "github.com/loloneme/pulse-flow/internal/domain/order"
	"github.com/loloneme/pulse-flow/internal/infrastructure/persistence"
)

const (
	alias     = "o"
	tableName = "orders"
	pkField   = "id"
)

type Repository = persistence.Repository[uuid.UUID, domain.Order, Order]

func NewRepository(db *sqlx.DB) *Repository {
	cols := persistence.NewColumns(writableColumns, readableColumns, alias, pkField)
	return persistence.NewRepository[uuid.UUID, domain.Order, Order](db, tableName, cols)
}
