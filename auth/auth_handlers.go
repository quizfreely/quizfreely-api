package auth

import (
	"net/http"
	"context"
	"errors"
	"unicode"
	"regexp"
	"encoding/json"
	"quizfreely/api/dbpool"
	"quizfreely/api/graph/model"

	"github.com/go-chi/render"
	"github.com/jackc/pgx/v5"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/rs/zerolog/log"
)

type SignUpReqBody struct {
	Username string `json:"username"`
	NewPassword string `json:"password"`
}

/* usernames can have letters and numbers from any alphabet, but no uppercase
underscores, dots, or dashes,
and must be less than 100 characters */
var usernameRegex = regexp.MustCompile(`^[\p{L}\p{M}\p{N}._-]+$`)
/* keep regexp.MustCompile outside of functions/handlers,
cause it's resource-expensive, we only want it to run once */
func IsUsernameValid(s string) bool {
	if len(s) == 0 || len(s) >= 100 {
		return false;
	}
	for _, r := range s {
		if unicode.Is(unicode.Upper, r) {
			return false
		}
	}
	return usernameRegex.MatchString(s)
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var reqBody SignUpReqBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 400,
				"message": "Error parsing JSON",
			},
		})
		return
	}

	if !IsUsernameValid(reqBody.Username) {
		render.Status(r, 400)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"code": "USERNAME_INVALID",
				"statusCode": 400,
				"message": "Usernames must be less than 100 characters & can only have letters/numbers (any alphabet, but no uppercase), underscores, dots, or dashes",
			},
		})
		return
	}

	var isUsernameTaken bool = false
	err = pgxscan.Get(
		context.Background(),
		dbpool.Pool,
		&isUsernameTaken,
		`SELECT EXISTS (
	SELECT 1 FROM auth.users
	WHERE username = $1 )`,
		reqBody.Username,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Error().Err(err).Msg("Database err while checking if username is taken in SignUpHandler")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 400,
				"message": "Database error while checking if username is taken in SignUpHandler",
			},
		})
		return
	}

	if isUsernameTaken {
		render.Status(r, 400)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"code": "USERNAME_TAKEN",
				"statusCode": 400,
				"message": "Username taken/already being used",
			},
		})
		return
	}

	var newUser model.AuthedUser
	err = pgxscan.Get(
		context.Background(),
		dbpool.Pool,
		&newUser,
		`INSERT INTO auth.users (username, encrypted_password, display_name, auth_type)
VALUES ($1, crypt($2, gen_salt('bf')), $1, 'USERNAME_PASSWORD')
RETURNING id, username, display_name, auth_type`,
		reqBody.Username,
		reqBody.NewPassword,
	)
	if err != nil {
		log.Error().Err(err).Msg("Database err while creating account in SignUpHandler")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 500,
				"message": "Database error while creating account in SignUpHandler",
			},
		})
		return
	}

	var newToken string
	err = pgxscan.Get(
		context.Background(),
		dbpool.Pool,
		&newToken,
		`INSERT INTO auth.sessions (user_id)
VALUES ($1) RETURNING token`,
		newUser.ID,
	)
	if err != nil {
		log.Error().Err(err).Msg("Database err while adding session in SignUpHandler")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 500,
				"message": "Database error while adding session in SignUpHandler",
			},
		})
		return
	}

	cookie := http.Cookie{
		Name: "auth",
		Value: newToken,
		Path: "/",
		/* 10 days * 24 hours per day * 60 mins per hour * 60s per min
		= 864000 seconds = 10 days */
		MaxAge: 864000,
		Secure: true,
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
	Token string `db:"token"`
	ID *string `db:"id"`
	Username         *string   `db:"username"`
	DisplayName      *string   `db:"display_name"`
	AuthType         *model.AuthType `db:"auth_type"`
	OauthGoogleEmail *string   `db:"oauth_google_email"`
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	var reqBody SignInReqBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 400,
				"message": "Error parsing JSON",
			},
		})
		return
	}
	if len(reqBody.Username) == 0 || len(reqBody.Username) >= 100 {
		render.Status(r, 400)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"code": "INCORRECT_USERNAME",
				"statusCode": 400,
				"message": "Invalid/wrong username",
			},
		})
		return
	}

	var tokenAndAuthedUser TokenAndAuthedUser
	err = pgxscan.Get(
		context.Background(),
		dbpool.Pool,
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
			context.Background(),
			dbpool.Pool,
			&usernameExists,
			`SELECT EXISTS (
	SELECT 1 FROM auth.users
	WHERE username = $1 )`,
			reqBody.Username,
		)
		/* `select exists` always returns 1 row (true or false, but not pgx.ErrNoRows) */
		if err2 != nil {
			log.Error().Err(err2).Msg("Database err while double-checking username in SignInHandler")
			render.Status(r, 500)
			render.JSON(w, r, map[string]interface{}{
				"error": map[string]interface{}{
					"statusCode": 500,
					"message": "Database error while double-checking username while signing in",
				},
			})
			return
		}

		/* `select exists` always returns 1 row (true or false, but not pgx.ErrNoRows)
		so we check if usernameExists is true or false */
		if (usernameExists) {
			render.Status(r, 400)
			render.JSON(w, r, map[string]interface{}{
				"error": map[string]interface{}{
					"code": "INCORRECT_PASSWORD",
					"statusCode": 400,
					"message": "Incorrect password",
				},
			})
			return
		} else {
			render.Status(r, 400)
			render.JSON(w, r, map[string]interface{}{
				"error": map[string]interface{}{
					"code": "INCORRECT_USERNAME",
					"statusCode": 400,
					"message": "Incorrect username",
				},
			})
			return
		}
	} else if err != nil {
		log.Error().Err(err).Msg("Database err in SignInHandler")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 500,
				"message": "Database error while signing in",
			},
		})
		return
	}

	cookie := http.Cookie{
		Name: "auth",
		Value: tokenAndAuthedUser.Token,
		Path: "/",
		/* 10 days * 24 hours per day * 60 mins per hour * 60s per min
		= 864000 seconds = 10 days */
		MaxAge: 864000,
		Secure: true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	render.JSON(w, r, map[string]interface{}{
		"error": false,
		"data": map[string]interface{}{
			"user": map[string]interface{}{
				"id": tokenAndAuthedUser.ID,
				"username": tokenAndAuthedUser.Username,
				"display_name": tokenAndAuthedUser.DisplayName,
				"auth_type": tokenAndAuthedUser.AuthType,
				"oauth_google_email": tokenAndAuthedUser.OauthGoogleEmail,
			},
		},
	})
}
