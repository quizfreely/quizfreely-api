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
		`SELECT u.id, u.username, u.display_name
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(id, og_order)
LEFT JOIN auth.users u ON u.id = input.id
ORDER BY input.og_order`,
		userIDs,
	)
	if err != nil {
		return nil, []error{err}
	}

	return users, nil
}

func (dr *dataReader) getTermsProgress(ctx context.Context, termAndUserIDs [][]string) ([]*model.TermProgress, []error) {
	var termsProgress []*model.TermProgress

	err := pgxscan.Select(
		ctx,
		dr.db,
		&termsProgress,
		`SELECT tp.id, tp.term_first_reviewed_at, tp.term_last_reviewed_at,
tp.term_review_count, tp.def_first_reviewed_at, tp.def_last_reviewed_at,
tp.def_review_count, tp.term_leitner_system_box, tp.def_leitner_system_box
FROM unnest($1::uuid[2][]) WITH ORDINALITY AS input(ids, og_order)
LEFT JOIN term_progress tp
	ON tp.term_id = input.ids[1]
	AND tp.user_id = input.ids[2]
ORDER BY input.og_order`,
		termAndUserIDs,
	)
	if err != nil {
		return nil, []error{err}
	}

	return termsProgress, nil
}

// Loaders wrap your data loaders to inject via middleware
type Loaders struct {
	UserLoader *dataloadgen.Loader[string, *model.User]
	TermProgressLoader *dataloadgen.Loader[[]string, *model.TermProgress]
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(db *pgxpool.Pool) *Loaders {
	// define the data loader
	dr := &dataReader{db: db}
	return &Loaders{
		UserLoader: dataloadgen.NewLoader(dr.getUsers, dataloadgen.WithWait(time.Millisecond)),
		TermProgressLoader: dataloadgen.NewLoader(dr.getTermsProgress, dataloadgen.WithWait(time.Millisecond)),
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

// GetUser returns a single term's progress record by term id and user id efficiently
func GetTermProgress(ctx context.Context, termAndUserID []string) (*model.TermProgress, error) {
	loaders := For(ctx)
	return loaders.TermProgressLoader.Load(ctx, termAndUserID)
}

// GetUsers returns many terms' progress records by term ids and user ids efficiently
func GetTermsProgress(ctx context.Context, termAndUserIDs [][]string) ([]*model.TermProgress, error) {
	loaders := For(ctx)
	return loaders.TermProgressLoader.LoadAll(ctx, termAndUserIDs)
}
