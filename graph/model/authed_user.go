package model

type AuthedUser struct {
	ID               *string   `json:"id,omitempty" db:"id"`
	Username         *string   `json:"username,omitempty" db:"username"`
	DisplayName      *string   `json:"display_name,omitempty" db:"display_name,omitempty`
	AuthType         *AuthType `json:"auth_type,omitempty" db:"auth_type,omitempty`
	OauthGoogleEmail *string   `json:"oauth_google_email,omitempty" db:"oauth_google_email,omitempty`
}
