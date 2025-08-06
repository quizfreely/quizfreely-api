package dbpool

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Init (dbUrl string) {
	var err error
	Pool, err = pgxpool.New(
		context.Background(),
		dbUrl,
	)
	if err != nil {
		log.Fatal(
			"Error creating database connection pool: ",
			err,
		)
	}
}
