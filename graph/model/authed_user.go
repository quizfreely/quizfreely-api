package model

type AuthedUser struct {
	ID               *string   `json:"id,omitempty" db:"id"`
	Username         *string   `json:"username,omitempty" db:"username"`
	DisplayName      *string   `json:"displayName,omitempty" db:"display_name"`
	AuthType         *AuthType `json:"authType,omitempty" db:"auth_type"`
	OauthGoogleEmail *string   `json:"oauthGoogleEmail,omitempty" db:"oauth_google_email"`
}
