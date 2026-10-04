package pgxotel

import (
	"github.com/jackc/pgx/v5"
	"github.com/pgx-contrib/pgxotel"
)

func NewQueryTracer(name string) pgx.QueryTracer {
	return &pgxotel.QueryTracer{Name: name}
}
