package loader

// import vikstrous/dataloadgen with your other imports
import (
	"context"
	"net/http"
	"time"
	"quizfreely/api/graph/model"

	"github.com/vikstrous/dataloadgen"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/georgysavva/scany/v2/pgxscan"
)

type ctxKey string

const (
	loadersKey = ctxKey("dataloaders")
)

type dataReader struct {
	db *pgxpool.Pool
}

// getUsers implements a batch function that can retrieve many users by ID,
// for use in a dataloader
func (dr *dataReader) getUsers(ctx context.Context, userIDs []string) ([]*model.User, []error) {
	var users []*model.User

	err := pgxscan.Select(
		ctx,
		dr.db,
		&users,
		`SELECT id, username, display_name
FROM auth.users
		WHERE id = ANY($1::uuid[])`,
		userIDs,
	)
	if err != nil {
		return nil, []error{err}
	}

	return users, nil
}

// Loaders wrap your data loaders to inject via middleware
type Loaders struct {
	UserLoader *dataloadgen.Loader[string, *model.User]
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(db *pgxpool.Pool) *Loaders {
	// define the data loader
	dr := &dataReader{db: db}
	return &Loaders{
		UserLoader: dataloadgen.NewLoader(dr.getUsers, dataloadgen.WithWait(time.Millisecond)),
	}
}

// Middleware injects data loaders into the context
func Middleware(db *pgxpool.Pool, next http.Handler) http.Handler {
	// return a middleware that injects the loader to the request context
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loader := NewLoaders(db)
		r = r.WithContext(context.WithValue(r.Context(), loadersKey, loader))
		next.ServeHTTP(w, r)
	})
}

// For returns the dataloader for a given context
func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

// GetUser returns single user by id efficiently
func GetUser(ctx context.Context, userID string) (*model.User, error) {
	loaders := For(ctx)
	return loaders.UserLoader.Load(ctx, userID)
}

// GetUsers returns many users by ids efficiently
func GetUsers(ctx context.Context, userIDs []string) ([]*model.User, error) {
	loaders := For(ctx)
	return loaders.UserLoader.LoadAll(ctx, userIDs)
}
