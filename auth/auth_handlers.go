package auth

import (
	"net/http"
	"context"
	"errors"
	"regexp"
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

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var reqBody SignUpReqBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 400,
				"message": "Error parsing JSON",
			}
		})
		return
	}

	if len(reqBody.Username) >= 100 || !regexp.MustCompile(`^(?!.*\p{Lu})[\p{L}\p{M}\p{N}._-]+$`).MatchString(reqBody.Username) {
		render.Status(r, 400)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"code": "USERNAME_INVALID",
				"statusCode": 400,
				"message": "Usernames must be less than 100 characters & can only have letters/numbers (any alphabet), underscores, dots, or dashes",
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
			}
		})
		return
	}

	if isUsernameTaken {
		render.Status(400)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"code": "USERNAME_TAKEN",
				"statusCode": 400,
				"message": "Username taken/already being used"
			}
		})
		return
	}

	var newUser model.AuthedUser
	err = pgxscan.Get(
		context.Background(),
		dbpool.Pool,
		&isUsernameTaken,
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
			}
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
		newUser.ID
	)
	if err != nil {
		log.Error().Err(err).Msg("Database err while adding session in SignUpHandler")
		render.Status(r, 500)
		render.JSON(w, r, map[string]interface{}{
			"error": map[string]interface{}{
				"statusCode": 500,
				"message": "Database error while adding session in SignUpHandler",
			}
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
