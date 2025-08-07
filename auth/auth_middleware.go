package auth

import (
	"net/http"
	"context"
	"quizfreely/api/dbpool"

	"github.com/google/uuid"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/rs/zerolog/log"
)

var authedUserCtxKey = &contextKey{"authedUser"}
type contextKey = struct {
	name string
}

type AuthedUser struct {
	ID               *string `db:"id"`
	Username         *string `db:"username"`
	DisplayName      *string `db:"display_name"`
	AuthType         *AuthType `db:"auth_type"`
	OauthGoogleEmail *string `db:"oauth_google_email"`
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		var token string

		cookie, err := r.Cookie("auth")
		if err == nil && cookie != nil {
			/* if auth cookie exists, use it as token */
			token = cookie.Value
		}

		if token == "" {
			/* if token is still empty, check Authorization header */
			header := r.Header.Get("Authorization")
			if header != "" {
				/* split `Bearer abc123def456` by space,
				then check if "Bearer" and actual token both exist */
				headerParts := strings.SplitN(header, " ", 2)
				if len(headerParts) == 2 && strings.EqualFold(headerParts[0], "Bearer ") {
					token = headerParts[1]
				} else {
					/* only send an error if the header exists but is wrong,
					if it doesn't exist, they're not logged in, which is fine, no err */
					http.Error(w, "Authorization header exists, but it's invalid", 400)
					return
				}
			}
		}

		if token != "" {
			/* if token is NOT empty, use it to get their account
			(doesn't matter if it's from the `auth` cookie or `Authorization` header) */
			var authedUser *AuthedUser
			err = pgxscan.Get(
				context.Background(),
				dbpool.Pool,
				&authedUser,
				`SELECT u.id, u.username, u.display_name, u.auth_type, u.oauth_google_email
FROM auth.sessions s
JOIN auth.users u ON s.user_id = u.id
WHERE s.token = $1 AND s.expire_at > now()`,
				token
			)
			if err != nil {
				ctx := context.WithValue(r.Context(), authedUserCtxKey, authedUser)
				r = r.WithContext(ctx)
			} else {
				log.Error().Err(err).Msg("Database error in AuthMiddleware")
			}
		}

		next.ServeHTTP(w, r)
	})
}
