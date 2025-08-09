package rest

import "github.com/jackc/pgx/v5/pgxpool"

type RESTHandler struct {
	DB *pgxpool.Pool
}

