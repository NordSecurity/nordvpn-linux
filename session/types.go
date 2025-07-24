package session

import "time"

type CredentialsBasedSession interface {
	GetUsername() string
	GetPassword() string
}

type SessionTokenProvider interface {
	GetToken() string
}

type SessionRenewalTokenProvider interface {
	GetRenewalToken() string
}

type SessionIDProvider interface {
	GetID() string
}

type SessionOwnerProvider interface {
	GetOwnerID() string
}

type SessionOwnerConsumer interface {
	SetOwnerID() string
}

type SessionExpiryProvider interface {
	GetExpiry() time.Time
}

type ExpirableSession interface {
	IsExpired() bool
}
