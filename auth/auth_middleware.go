package auth

import (
	"context"
	"errors"
	"net/http"
	"quizfreely/api/graph/model"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

var authedUserCtxKey = &contextKey{"authedUser"}

type contextKey struct {
	name string
}

func (ah *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
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
				if len(headerParts) == 2 && strings.EqualFold(headerParts[0], "Bearer") {
					token = headerParts[1]
				} else {
					/* only send an error if the header exists but is wrong,
					if it doesn't exist, they're not logged in, which is fine, no err */
					http.Error(w, "Authorization header exists, but it's invalid", 400)
					return
				}
			}
		}
		/* this AuthMiddleware should only send an error if the structure of the req
		is broken. stuff like invalid/expired tokens should NOT cause an error response
		because this middleware should only populate authedUser context stuff
		each handler controls any error responses if not logged in based on auth context,
		because some handlers allow not-logged-in users while others might not */

		if token != "" {
			/* if token is NOT empty,
			AFTER checking the cookie & header with the 2 if-statements above,
			use it to get their account */
			authedUser := &model.AuthedUser{}
			err = pgxscan.Get(
				context.Background(),
				ah.DB,
				authedUser,
				`SELECT u.id, u.username, u.display_name, u.auth_type, u.oauth_google_email
FROM auth.sessions s
JOIN auth.users u ON s.user_id = u.id
WHERE s.token = $1 AND s.expire_at > now()`,
				token,
			)
			if err == nil {
				ctx := context.WithValue(r.Context(), authedUserCtxKey, authedUser)
				r = r.WithContext(ctx)
			} else {
				/* if err is pgx.ErrNoRows, that means the token is invalid or expired,
				don't throw an error for that, just continue as a not-logged-in user,
				expired tokens shouldn't cause an error when querying public data for example */
				if !errors.Is(err, pgx.ErrNoRows) {
					log.Error().Err(err).Msg("Database error in AuthMiddleware")
					/* unlike pgx.ErrNoRows, which is probably a client error
					like an invalid or expired token,
					other errors should be logged because it might be the api's fault */
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func ForContext(ctx context.Context) *model.AuthedUser {
	raw, _ := ctx.Value(authedUserCtxKey).(*model.AuthedUser)
	return raw
}
