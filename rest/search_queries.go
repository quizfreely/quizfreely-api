package rest

import (
	"net/http"
	"strconv"
	"quizfreely/api/graph/model"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

func (rh *RESTHandler) GetSearchQueries(w http.ResponseWriter, r *http.Request) {
	searchParams := r.URL.Query() /* query string */
	q := searchParams.Get("q")
	limit := searchParams.Get("limit")
	offset := searchParams.Get("offset")

	l := 5
	if limit != "" {
		lInt, err := strconv.Atoi(limit)
		if err == nil && lInt > 0 && lInt < 5 {
			l = lInt
		}
	}

	o := 0
	if offset != "" {
		oInt, err := strconv.Atoi(offset)
		if err == nil && oInt > 0 {
			o = oInt
		}
	}

	var searchQueries []*model.SearchQuery
	sql := `
		SELECT
			id,
			query,
			to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SS.MSTZH:TZM') as updated_at,
			(SELECT display_name FROM auth.users WHERE id = search_queries.user_id) AS user_display_name,
			(SELECT COUNT(*) FROM public.studysets WHERE document @@ websearch_to_tsquery('english', search_queries.query)) AS studyset_count
		FROM public.search_queries
		WHERE query ILIKE $1
		ORDER BY (SELECT COUNT(*) FROM public.studysets WHERE document @@ websearch_to_tsquery('english', search_queries.query)) DESC
		LIMIT $2 OFFSET $3
	`
	err := pgxscan.Select(r.Context(), rh.DB, &searchQueries, sql, "%"+q+"%", l, o)
	if err != nil {
		log.Error().Err(err).Msg("Database err in GetSearchQueries")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 400,
				"message":    "Database error in GetSearchQueries",
			},
		})
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"error": false,
		"data": map[string]interface{}{
			"queries": searchQueries,
		},
	})
}
