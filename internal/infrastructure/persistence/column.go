package persistence

import (
	"fmt"
	"strings"
)

type Columns struct {
	writable   []string
	readable   []string
	alias      string
	primaryKey string
}

func NewColumns(writable []string, readable []string, alias string, primaryKey string) *Columns {
	return &Columns{
		writable:   writable,
		readable:   readable,
		alias:      alias,
		primaryKey: primaryKey,
	}
}

func (c *Columns) ForSelect(rawFields []string) []string {
	if len(rawFields) == 0 {
		return c.readable
	}

	fields := make([]string, 0, len(rawFields))
	for _, raw := range rawFields {
		if raw == "*" {
			fields = append(fields, c.readable...)
			break
		} else if raw != c.alias+"."+c.primaryKey {
			fields = append(fields, c.alias+"."+raw)
		} else {
			for _, field := range c.readable {
				fields = append(fields, c.alias+"."+field)
			}
		}
	}

	return fields
}

func (c *Columns) ForInsert() []string {
	return c.writable
}

func (c *Columns) GetOnConflictStatement() string {
	onConflict := make([]string, 0, len(c.writable))
	needChangeUpdatedAt := false

	for _, column := range c.writable {
		if !needChangeUpdatedAt && column == "updated_at" {
			needChangeUpdatedAt = true
			continue
		}

		onConflict = append(onConflict, fmt.Sprintf("%s=EXCLUDED.%s", column, column))
	}

	if needChangeUpdatedAt {
		onConflict = append(onConflict, "updated_at=now()")
	}

	return strings.Join(onConflict, ", ")
}
