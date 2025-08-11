package graph

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	"regexp"

	"github.com/jackc/pgx/v5/pgxpool"
)

var validTitleRegex = regexp.MustCompile(`[\p{L}\p{M}\p{N}]`)

type Resolver struct {
	DB *pgxpool.Pool
}
