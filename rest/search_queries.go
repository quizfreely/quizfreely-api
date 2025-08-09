package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"unicode"
	"strconv"
	"quizfreely/api/graph/model"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

	return searchQueries, nil


	var isUsernameTaken bool = false
	err = pgxscan.Get(
		r.Context(),
		ah.DB,
		&isUsernameTaken,
		`SELECT EXISTS (
	SELECT 1 FROM auth.users
	WHERE username = $1 )`,
		reqBody.Username,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Error().Err(err).Msg("Database err while checking if username is taken in SignUp")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 400,
				"message":    "Database error while checking if username is taken in SignUp",
			},
		})
		return
	}

	if isUsernameTaken {
		render.Status(r, 400)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"code":       "USERNAME_TAKEN",
				"statusCode": 400,
				"message":    "Username taken/already being used",
			},
		})
		return
	}

	var newUser model.AuthedUser
	err = pgxscan.Get(
		r.Context(),
		ah.DB,
		&newUser,
		`INSERT INTO auth.users (username, encrypted_password, display_name, auth_type)
VALUES ($1, crypt($2, gen_salt('bf')), $1, 'USERNAME_PASSWORD')
RETURNING id, username, display_name, auth_type`,
		reqBody.Username,
		reqBody.NewPassword,
	)
	if err != nil {
		log.Error().Err(err).Msg("Database err while creating account in SignUp")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 500,
				"message":    "Database error while creating account in SignUp",
			},
		})
		return
	}

	var newToken string
	err = pgxscan.Get(
		r.Context(),
		ah.DB,
		&newToken,
		`INSERT INTO auth.sessions (user_id)
VALUES ($1) RETURNING token`,
		newUser.ID,
	)
	if err != nil {
		log.Error().Err(err).Msg("Database err while adding session in SignUp")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 500,
				"message":    "Database error while adding session in SignUp",
			},
		})
		return
	}

	cookie := http.Cookie{
		Name:  "auth",
		Value: newToken,
		Path:  "/",
		/* 10 days * 24 hours per day * 60 mins per hour * 60s per min
		= 864000 seconds = 10 days */
		MaxAge:   864000,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	render.JSON(w, r, map[string]interface{}{
		"error": false,
		"data": map[string]interface{}{
			"user": newUser,
		},
	})
}

type SignInReqBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenAndAuthedUser struct {
	Token            string          `db:"token"`
	ID               *string         `db:"id"`
	Username         *string         `db:"username"`
	DisplayName      *string         `db:"display_name"`
	AuthType         *model.AuthType `db:"auth_type"`
	OauthGoogleEmail *string         `db:"oauth_google_email"`
}

func (ah *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var reqBody SignInReqBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 400,
				"message":    "Error parsing JSON",
			},
		})
		return
	}
	if len(reqBody.Username) == 0 || len(reqBody.Username) >= 100 {
		render.Status(r, 400)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"code":       "INCORRECT_USERNAME",
				"statusCode": 400,
				"message":    "Invalid/wrong username",
			},
		})
		return
	}

	var tokenAndAuthedUser TokenAndAuthedUser
	err = pgxscan.Get(
		r.Context(),
		ah.DB,
		&tokenAndAuthedUser,
		`WITH u AS (
	SELECT id, username, display_name, auth_type, oauth_google_email
	FROM auth.users
	WHERE username = $1 AND
		encrypted_password = crypt($2, encrypted_password)
), s AS (
	INSERT INTO auth.sessions (user_id)
	SELECT id FROM u
	RETURNING token
) SELECT s.token, u.id, u.username, u.display_name,
	u.auth_type, u.oauth_google_email
FROM s, u`,
		reqBody.Username,
		reqBody.Password,
	)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		var usernameExists bool = false
		err2 := pgxscan.Get(
			r.Context(),
			ah.DB,
			&usernameExists,
			`SELECT EXISTS (
	SELECT 1 FROM auth.users
	WHERE username = $1 )`,
			reqBody.Username,
		)
		/* `select exists` always returns 1 row (true or false, but not pgx.ErrNoRows) */
		if err2 != nil {
			log.Error().Err(err2).Msg("Database err while double-checking username in SignIn")
			render.Status(r, 500)
			render.JSON(w, r, map[string]interface{}{
				"error": map[string]interface{}{
					"statusCode": 500,
					"message":    "Database error while double-checking username while signing in",
				},
			})
			return
		}

		/* `select exists` always returns 1 row (true or false, but not pgx.ErrNoRows)
		so we check if usernameExists is true or false */
		if usernameExists {
			render.Status(r, 400)
			render.JSON(w, r, map[string]interface{}{
				"error": map[string]interface{}{
					"code":       "INCORRECT_PASSWORD",
					"statusCode": 400,
					"message":    "Incorrect password",
				},
			})
			return
		} else {
			render.Status(r, 400)
			render.JSON(w, r, map[string]interface{}{
				"error": map[string]interface{}{
					"code":       "INCORRECT_USERNAME",
					"statusCode": 400,
					"message":    "Incorrect username",
				},
			})
			return
		}
	} else if err != nil {
		log.Error().Err(err).Msg("Database err in SignIn")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 500,
				"message":    "Database error while signing in",
			},
		})
		return
	}

	cookie := http.Cookie{
		Name:  "auth",
		Value: tokenAndAuthedUser.Token,
		Path:  "/",
		/* 10 days * 24 hours per day * 60 mins per hour * 60s per min
		= 864000 seconds = 10 days */
		MaxAge:   864000,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	render.JSON(w, r, map[string]interface{}{
		"error": false,
		"data": map[string]interface{}{
			"user": map[string]interface{}{
				"id":                 tokenAndAuthedUser.ID,
				"username":           tokenAndAuthedUser.Username,
				"display_name":       tokenAndAuthedUser.DisplayName,
				"auth_type":          tokenAndAuthedUser.AuthType,
				"oauth_google_email": tokenAndAuthedUser.OauthGoogleEmail,
			},
		},
	})
}

type DeleteAccountRequest struct {
	DeleteAllMyStudysets bool   `json:"deleteAllMyStudysets"`
	ConfirmPassword      string `json:"confirmPassword"`
}

func (ah *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	var req DeleteAccountRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 400,
				"message":    "Error parsing JSON",
			},
		})
		return
	}

	authedUser := AuthedUserContext(r.Context())
	if authedUser == nil {
		render.Status(r, 401)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"code":       "NOT_AUTHED",
				"statusCode": 401,
				"message":    "You are not signed in, so you cannot delete your account",
			},
		})
		return
	}

	tx, err := ah.DB.Begin(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("Database err while starting transaction in DeleteAccount")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 500,
				"message":    "Database error while deleting account",
			},
		})
		return
	}
	defer tx.Rollback(r.Context())

	// Delete studysets based on user preference
	if req.DeleteAllMyStudysets {
		_, err = tx.Exec(r.Context(), "delete from public.studysets where user_id = $1", authedUser.ID)
	} else {
		_, err = tx.Exec(r.Context(), "delete from public.studysets where user_id = $1 and private = true", authedUser.ID)
	}
	if err != nil {
		log.Error().Err(err).Msg("Database err while deleting studysets in DeleteAccount")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 500,
				"message":    "Database error while deleting account",
			},
		})
		return
	}

	// Set auth context before deleting user
	_, err = tx.Exec(r.Context(), "select set_config('qzfr_api.scope', 'auth', true)")
	if err != nil {
		log.Error().Err(err).Msg("Database err while setting auth context in DeleteAccount")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 500,
				"message":    "Database error while deleting account",
			},
		})
		return
	}

	// Delete user based on auth type
	if authedUser.AuthType != nil && *authedUser.AuthType == model.AuthTypeOauthGoogle {
		_, err = tx.Exec(r.Context(), "delete from auth.users where id = $1", authedUser.ID)
		if err != nil {
			log.Error().Err(err).Msg("Database err while deleting user in DeleteAccount")
			render.Status(r, 500)
			render.JSON(w, r, map[string]interface{}{
				"error": map[string]interface{}{
					"statusCode": 500,
					"message":    "Database error while deleting account",
				},
			})
			return
		}
	} else {
		if req.ConfirmPassword == "" {
			render.Status(r, 403)
			render.JSON(w, r, map[string]interface{}{
				"error": map[string]interface{}{
					"code":       "INCORRECT_PASSWORD",
					"statusCode": 403,
					"message":    "Password confirmation is required",
				},
			})
			return
		}

		var deleted bool
		err = tx.QueryRow(
			r.Context(),
			"delete from auth.users where id = $1 and encrypted_password = crypt($2, encrypted_password) RETURNING true",
			authedUser.ID, req.ConfirmPassword,
		).Scan(&deleted)

		if err != nil || !deleted {
			log.Error().Err(err).Msg("Database err while deleting user with password in DeleteAccount")
			render.Status(r, 403)
			render.JSON(w, r, map[string]interface{}{
				"error": map[string]interface{}{
					"code":       "INCORRECT_PASSWORD",
					"statusCode": 403,
					"message":    "Wrong password confirmation while trying to delete account",
				},
			})
			return
		}
	}

	err = tx.Commit(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("Database err while committing transaction in DeleteAccount")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 500,
				"message":    "Database error while deleting account",
			},
		})
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"error": false,
		"data": map[string]interface{}{
			"authed": false,
		},
	})
}
