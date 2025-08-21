package loader

// import vikstrous/dataloadgen with your other imports
import (
	"context"
	"net/http"
	"time"
	"quizfreely/api/auth"
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

func (dr *dataReader) getTermsByIDs(ctx context.Context, ids []string) ([]*model.Term, []error) {
	var terms []*model.Term

	err := pgxscan.Select(
		ctx,
		dr.db,
		&terms,
		`SELECT t.id, t.studyset_id, t.term, t.def, t.sort_order,
	to_char(t.created_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as created_at,
	to_char(t.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as updated_at
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(id, og_order)
LEFT JOIN terms t
	ON t.id = input.id
ORDER BY input.og_order`,
		ids,
	)
	if err != nil {
		return nil, []error{err}
	}

	return terms, nil
}

func (dr *dataReader) getTermsByStudysetIDs(ctx context.Context, studysetIDs []string) ([][]*model.Term, []error) {
	var terms []*model.Term

	err := pgxscan.Select(
		ctx,
		dr.db,
		&terms,
		`SELECT t.id, t.studyset_id, t.term, t.def, t.sort_order,
	to_char(t.created_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as created_at,
	to_char(t.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as updated_at
FROM terms t
WHERE t.studyset_id = ANY($1::uuid[])
ORDER BY t.studyset_id, t.sort_order`,
		studysetIDs,
	)
	if err != nil {
		return nil, []error{err}
	}

    // Group terms by studyset_id
    grouped := make(map[string][]*model.Term)
    for _, t := range terms {
		if t.StudysetID != nil {
        	grouped[*t.StudysetID] = append(grouped[*t.StudysetID], t)
		}
    }

    // Reassemble in the same order as studysetIDs
    orderedTerms := make([][]*model.Term, len(studysetIDs))
    for i, id := range studysetIDs {
        orderedTerms[i] = grouped[id]
    }

	return orderedTerms, nil
}

func (dr *dataReader) getTermsCountByStudysetIDs(ctx context.Context, studysetIDs []string) ([]*int32, []error) {
    type countResult struct {
        StudysetID string `db:"studyset_id"`
        Count      int32  `db:"term_count"`
    }

    var results []countResult

    err := pgxscan.Select(
        ctx,
        dr.db,
        &results,
        `SELECT studyset_id, COUNT(*) AS term_count
         FROM terms
         WHERE studyset_id = ANY($1::uuid[])
         GROUP BY studyset_id`,
        studysetIDs,
    )
    if err != nil {
        return nil, []error{err}
    }

    // Map studysetID -> count for quick lookup
    countsMap := make(map[string]int32, len(results))
    for _, r := range results {
        countsMap[r.StudysetID] = r.Count
    }

    // Assemble slice in the same order as studysetIDs
    orderedCounts := make([]*int32, len(studysetIDs))
    for i, id := range studysetIDs {
        if c, ok := countsMap[id]; ok {
            orderedCounts[i] = &c
        } else {
            zero := int32(0)
            orderedCounts[i] = &zero
        }
    }

    return orderedCounts, nil
}

func (dr *dataReader) getTermsProgress(ctx context.Context, termIDs []string) ([]*model.TermProgress, []error) {
	authedUser := auth.AuthedUserContext(ctx)

	var termsProgress []*model.TermProgress

	err := pgxscan.Select(
		ctx,
		dr.db,
		&termsProgress,
		`SELECT tp.id,
	to_char(tp.term_first_reviewed_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as term_first_reviewed_at,
	to_char(tp.term_last_reviewed_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as term_last_reviewed_at,
	tp.term_review_count,
	to_char(tp.def_first_reviewed_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as def_first_reviewed_at,
	to_char(tp.def_last_reviewed_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as def_last_reviewed_at,
	tp.def_review_count,
	tp.term_leitner_system_box, tp.def_leitner_system_box,
	tp.term_correct_count, tp.term_incorrect_count,
	tp.def_correct_count, tp.def_incorrect_count
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(term_id, og_order)
LEFT JOIN term_progress tp
	ON tp.term_id = input.term_id
	AND tp.user_id = $2
ORDER BY input.og_order`,
		termIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}

	return termsProgress, nil
}

func (dr *dataReader) getTermsTopConfusionPairs(ctx context.Context, termIDs []string) ([][]*model.TermConfusionPair, []error) {
	authedUser := auth.AuthedUserContext(ctx)

	var confusionPairs []*model.TermConfusionPair

	err := pgxscan.Select(
		ctx,
		dr.db,
		&confusionPairs,
		`SELECT tcp.id,
	tcp.term_id,
    tcp.confused_term_id,
    tcp.answered_with,
    tcp.confused_count,
	to_char(tcp.last_confused_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as last_confused_at
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(term_id, og_order)
LEFT JOIN term_confusion_pairs tcp
	ON tcp.term_id = input.term_id
	AND tcp.user_id = $2
ORDER BY input.og_order ASC, tcp.confused_count DESC`,
		termIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}

    grouped := make(map[string][]*model.TermConfusionPair)
    for _, c := range confusionPairs {
        if c.TermID != nil {
			grouped[*c.TermID] = append(grouped[*c.TermID], c)
		}
    }

    orderedConfusionPairs := make([][]*model.TermConfusionPair, len(termIDs))
    for i, id := range termIDs {
        orderedConfusionPairs[i] = grouped[id]
    }

	return orderedConfusionPairs, nil
}

func (dr *dataReader) getPracticeTestsByStudysetIDs(ctx context.Context, studysetIDs []string) ([][]*model.PracticeTest, []error) {
	authedUser := auth.AuthedUserContext(ctx)

	var practiceTests []*model.PracticeTest

	err := pgxscan.Select(
		ctx,
		dr.db,
		&practiceTests,
		`SELECT pt.id,
	to_char(pt.timestamp, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as timestamp,
	pt.studyset_id,
    pt.questions_correct,
    pt.questions_total,
    pt.questions jsonb
FROM unnest($1::uuid[]) WITH ORDINALITY AS input(studyset_id, og_order)
LEFT JOIN practice_tests pt
	ON pt.studyset_id = input.studyset_id
	AND pt.user_id = $2
ORDER BY input.og_order ASC, pt.timestamp DESC`,
		studysetIDs,
		authedUser.ID,
	)
	if err != nil {
		return nil, []error{err}
	}

    grouped := make(map[string][]*model.PracticeTest)
    for _, pt := range practiceTests {
        if pt.StudysetID != nil {
			grouped[*pt.StudysetID] = append(grouped[*pt.StudysetID], pt)
		}
    }

    orderedPracticeTests := make([][]*model.PracticeTest, len(practiceTests))
    for i, id := range studysetIDs {
        orderedConfusionPairs[i] = grouped[id]
    }

	return orderedPracticeTests, nil
}

// Loaders wrap your data loaders to inject via middleware
type Loaders struct {
	UserLoader *dataloadgen.Loader[string, *model.User]
	TermByIDLoader *dataloadgen.Loader[string, *model.Term]
	TermByStudysetIDLoader *dataloadgen.Loader[string, []*model.Term]
	TermsCountByStudysetIDLoader *dataloadgen.Loader[string, *int32]
	TermProgressLoader *dataloadgen.Loader[string, *model.TermProgress]
	TermTopConfusionPairsLoader *dataloadgen.Loader[string, []*model.TermConfusionPair]
	PracticeTestByStudysetIDLoader *dataloadgen.Loader[string, []*model.PracticeTest]
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(db *pgxpool.Pool) *Loaders {
	// define the data loader
	dr := &dataReader{db: db}
	return &Loaders{
		UserLoader: dataloadgen.NewLoader(dr.getUsers, dataloadgen.WithWait(time.Millisecond)),
		TermByIDLoader: dataloadgen.NewLoader(dr.getTermsByIDs, dataloadgen.WithWait(time.Millisecond)),
		TermByStudysetIDLoader: dataloadgen.NewLoader(dr.getTermsByStudysetIDs, dataloadgen.WithWait(time.Millisecond)),
		TermsCountByStudysetIDLoader: dataloadgen.NewLoader(dr.getTermsCountByStudysetIDs, dataloadgen.WithWait(time.Millisecond)),
		TermProgressLoader: dataloadgen.NewLoader(dr.getTermsProgress, dataloadgen.WithWait(time.Millisecond)),
		TermTopConfusionPairsLoader: dataloadgen.NewLoader(dr.getTermsTopConfusionPairs, dataloadgen.WithWait(time.Millisecond)),
		PracticeTestByStudysetIDLoader: dataloadgen.NewLoader(dr.getPracticeTestsByStudysetIDs, dataloadgen.WithWait(time.Millisecond)),
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

func GetTermByID(ctx context.Context, id string) (*model.Term, error) {
	loaders := For(ctx)
	return loaders.TermByIDLoader.Load(ctx, id)
}

func GetTermsByIDs(ctx context.Context, ids []string) ([]*model.Term, error) {
	loaders := For(ctx)
	return loaders.TermByIDLoader.LoadAll(ctx, ids)
}

// GetTermsByStudysetID returns a single studyset's terms efficiently
func GetTermsByStudysetID(ctx context.Context, studysetID string) ([]*model.Term, error) {
	loaders := For(ctx)
	return loaders.TermByStudysetIDLoader.Load(ctx, studysetID)
}

// GetTermsByStudysetIDs returns many studysets' terms efficiently
func GetTermsByStudysetIDs(ctx context.Context, studysetIDs []string) ([][]*model.Term, error) {
	loaders := For(ctx)
	return loaders.TermByStudysetIDLoader.LoadAll(ctx, studysetIDs)
}

// GetTermsCountByStudysetID returns a single studyset's terms count efficiently
func GetTermsCountByStudysetID(ctx context.Context, studysetID string) (*int32, error) {
	loaders := For(ctx)
	return loaders.TermsCountByStudysetIDLoader.Load(ctx, studysetID)
}

// GetTermsCountByStudysetIDs returns many studysets' terms counts efficiently
func GetTermsCountByStudysetIDs(ctx context.Context, studysetIDs []string) ([]*int32, error) {
	loaders := For(ctx)
	return loaders.TermsCountByStudysetIDLoader.LoadAll(ctx, studysetIDs)
}

// GetTermProgress returns a single term's progress record by term id efficiently
func GetTermProgress(ctx context.Context, termID string) (*model.TermProgress, error) {
	loaders := For(ctx)
	return loaders.TermProgressLoader.Load(ctx, termID)
}

// GetTermsProgress returns many terms' progress records by term ids efficiently
func GetTermsProgress(ctx context.Context, termIDs []string) ([]*model.TermProgress, error) {
	loaders := For(ctx)
	return loaders.TermProgressLoader.LoadAll(ctx, termIDs)
}

// GetTermTopConfusionPairs returns a single term's confusion pairs
func GetTermTopConfusionPairs(ctx context.Context, termID string) ([]*model.TermConfusionPair, error) {
	loaders := For(ctx)
	return loaders.TermTopConfusionPairsLoader.Load(ctx, termID)
}

// GetTermsTopConfusionPairs returns many terms' confusion pairs
func GetTermsTopConfusionPairs(ctx context.Context, termIDs []string) ([][]*model.TermConfusionPair, error) {
	loaders := For(ctx)
	return loaders.TermTopConfusionPairsLoader.LoadAll(ctx, termIDs)
}

func getPracticeTestsByStudysetID(ctx context.Context, studysetID string) ([]*model.PracticeTest, error) {
	loaders := For(ctx)
	return loaders.TermTopConfusionPairsLoader.Load(ctx, studysetID)
}

func getPracticeTestsByStudysetIDs(ctx context.Context, studysetIDs []string) ([][]*model.PracticeTest, error) {
	loaders := For(ctx)
	return loaders.TermTopConfusionPairsLoader.LoadAll(ctx, studysetIDs)
}
