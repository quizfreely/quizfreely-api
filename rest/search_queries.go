package rest

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"quizfreely/api/graph/model"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

// async function searchQueries(query, limit, offset) {
//     try {
//         /*
//             replace whitespace (tabs, spaces, etc and multiple) with a space
//             whitespace characters next to eachother will be replaced with a single space
//         */
//         let spaceRegex = /\s+/gu;
//         let inputQuery = query.replaceAll(spaceRegex, " ");
//         /* after that, replace and sign ("&") with "and" */
//         inputQuery = inputQuery.replaceAll("&", "and");
//         /*
//             after "sanitizing" spaces, remove special characters
//             this will keep letters (any alphabet) accent marks, numbers (any alphabet), underscore, period/dot, and dashes
//         */
//         let rmRegex = /[^\p{L}\p{M}\p{N} _.-]/gu;
//         inputQuery = inputQuery.replaceAll(rmRegex, "");
//         let result;
//         if (inputQuery.length <= 3) {
//             result = await pool.query(
//                 "select query, subject from public.search_queries " +
//                 "where query ilike $1 limit $2 offset $3",
//                 [
//                     /* percent sign (%) to match querys that start with inputQuery */
//                     (inputQuery + "%"),
//                     limit,
//                     offset
//                 ]
//             )
//         } else {
//             result = await pool.query(
//                 "select query, subject from public.search_queries " +
//                 "where similarity(query, $1) > 0.15 " +
//                 "order by similarity(query, $1) desc limit $2 offset $3",
//                 [
//                     inputQuery,
//                     limit,
//                     offset
//                 ]
//             )
//         }
//         return {
//             data: result.rows
//         }
//     } catch (error) {
//         return {
//             error: error
//         }
//     }
// }

var spaceRegex = regexp.MustCompile(`\s+`)
var removeRegex = regexp.MustCompile(`[^\pL\pM\pN _.\-]`)

func (rh *RESTHandler) GetSearchQueries(w http.ResponseWriter, r *http.Request) {
	searchParams := r.URL.Query() /* query string */
	rawQuery := searchParams.Get("q")
	limit := searchParams.Get("limit")
	offset := searchParams.Get("offset")

	l := 5
	if limit != "" {
		lInt, err := strconv.Atoi(limit)
		if err == nil && lInt > 0 {
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

	// Sanitize input query
	inputQuery := spaceRegex.ReplaceAllString(rawQuery, " ")
	inputQuery = strings.ReplaceAll(inputQuery, "&", "and")
	inputQuery = removeRegex.ReplaceAllString(inputQuery, "")
	log.Info().Msg(inputQuery)

	var searchQueries []*model.SearchQuery
	var err error

	if len(inputQuery) <= 3 {
		err = pgxscan.Select(r.Context(), rh.DB, &searchQueries,
			"SELECT query, subject FROM public.search_queries WHERE query ILIKE $1 ORDER BY query LIMIT $2 OFFSET $3",
			inputQuery+"%", l, o)
	} else {
		err = pgxscan.Select(r.Context(), rh.DB, &searchQueries,
			"SELECT query, subject FROM public.search_queries WHERE similarity(query, $1) > 0.15 ORDER BY similarity(query, $1) DESC LIMIT $2 OFFSET $3",
			inputQuery, l, o)
	}

	if err != nil {
		log.Error().Err(err).Msg("Database error in GetSearchQueries")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 500,
				"message":    "Database error in GetSearchQueries",
			},
		})
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"error": nil,
		"data": map[string]interface{}{
			"queries": searchQueries,
		},
	})
}
