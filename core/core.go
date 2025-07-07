/*
Package core provides Go HTTP client for interacting with Core API a.k.a. NordVPN API
*/
package core

import (
	"fmt"
	"io"
	"net/http"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/google/uuid"
)

const (
	// linuxPlatformID defines the linux platform ID on the Notification Centre
	linuxPlatformID = 500
)

type CredentialsAPI interface {
	NotificationCredentials(appUserID string) (NotificationCredentialsResponse, error)
	NotificationCredentialsRevoke(appUserID string, purgeSession bool) (NotificationCredentialsRevokeResponse, error)
	ServiceCredentials(token string) (*CredentialsResponse, error)
	TokenRenew(renewalToken string, idempotencyKey uuid.UUID) (*TokenRenewResponse, error)
	Services() (ServicesResponse, error)
	CurrentUser() (*CurrentUserResponse, error)
	DeleteToken() error
	TrustedPassToken() (*TrustedPassTokenResponse, error)
	MultifactorAuthStatus() (*MultifactorAuthStatusResponse, error)
	Logout() error
}

type InsightsAPI interface {
	Insights() (*Insights, error)
}

type ServersAPI interface {
	Servers() (Servers, http.Header, error)
	RecommendedServers(filter ServersFilter, longitude, latitude float64) (Servers, http.Header, error)
	Server(id int64) (*Server, error)
	ServersCountries() (Countries, http.Header, error)
}

type CombinedAPI interface {
	InsightsAPI
	Base() string
	Plans() (*Plans, error)
	CreateUser(email, password string) (*UserCreateResponse, error)
}

// SubscriptionAPI is responsible for fetching the subscription data of the user
type SubscriptionAPI interface {
	// Orders returns a list of orders done by the user
	Orders() ([]Order, error)
	// Payments returns a list of payments done by the user
	Payments() ([]PaymentResponse, error)
}

type ErrMaxBytesLimit struct {
	Limit int64
}

func (err *ErrMaxBytesLimit) Error() string {
	return fmt.Sprintf("input exceeded the max limit of %d bytes", err.Limit)
}

// MaxBytesReadAll is a wrapper around io.ReadAll that limits the number of bytes read from the reader.
//
// If the reader exceeds the maxBytesLimit, the function returns an error.
func MaxBytesReadAll(r io.Reader) ([]byte, error) {
	limitedReader := &io.LimitedReader{
		R: r,
		N: internal.MaxBytesLimit,
	}
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}
	// check whether the io.ReadAll() stopped because of EOF coming from io.Reader or because of the limit
	//
	// two cases can happen here:
	// limit reached       - limitedReader.N <= 0
	// io.Reader is empty  - limitedReader.N > 0
	if limitedReader.N <= 0 {
		return nil, &ErrMaxBytesLimit{Limit: internal.MaxBytesLimit}
	}

	return data, nil
}
