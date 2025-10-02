package auth

type UserProfile struct {
	// Standard OIDC claims
	Aud string  `json:"aud"` // Audience (client ID)
	Exp float64 `json:"exp"` // Expiration time (Unix timestamp)
	Iat float64 `json:"iat"` // Issued at (Unix timestamp)
	Iss string  `json:"iss"` // Issuer
	Sub string  `json:"sub"` // Subject (user ID)
	Sid string  `json:"sid"` // Session ID

	// Auth0 profile claims
	Name      string `json:"name"`       // Full name or email
	Nickname  string `json:"nickname"`   // Username/nickname
	Picture   string `json:"picture"`    // Profile picture URL
	UpdatedAt string `json:"updated_at"` // Last profile update
}
