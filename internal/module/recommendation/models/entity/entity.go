package entity

import (
	"database/sql"
	"time"
)

type Venues struct {
	ID             int          `db:"id" json:"id"`
	Name           string       `db:"name" json:"name"`
	IsSoldOut      bool         `db:"is_sold_out" json:"is_sold_out"`
	IsFirstSoldOut bool         `db:"is_first_sold_out" json:"is_first_sold_out"`
	CreatedAt      time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt      sql.NullTime `db:"updated_at" json:"updated_at"`
	DeletedAt      sql.NullTime `db:"deleted_at" json:"deleted_at"`
}
