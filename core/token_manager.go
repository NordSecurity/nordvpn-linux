package core

// TokenManager manages access tokens, including renewal and storage.
type TokenManager interface {
	// Token returns currently used token
	Token() (string, error)

	// Renew renews the token based on its current validity.
	Renew() error

	// Store manually stores a new token
	Store(token string) error

	// Invalidate clears current token data directly or by delegating based on provided error reason
	Invalidate(err error) error
}
