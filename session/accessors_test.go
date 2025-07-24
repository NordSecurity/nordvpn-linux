package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type SimpleNothingProvider struct{}

type SimpleSessionTokenProvider struct {
	GetTokenFunc func() string
}

func (s *SimpleSessionTokenProvider) GetToken() string {
	return s.GetTokenFunc()
}

func Test_GetToken_IncompatibleInterfaces(t *testing.T) {
	res, err := GetToken(0)
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)

	res, err = GetToken("some text")
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)

	res, err = GetToken(&SimpleNothingProvider{})
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)

	res, err = GetToken(func() string { return "token" })
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)
}

func Test_GetToken_CompatibleInterface(t *testing.T) {
	expectedToken := "valid-token"
	prov := &SimpleSessionTokenProvider{GetTokenFunc: func() string { return expectedToken }}

	res, err := GetToken(prov)
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, res, "Must match")
}

type SimpleSessionRenewalProvider struct {
	GetRenewalTokenFunc func() string
}

func (s *SimpleSessionRenewalProvider) GetRenewalToken() string {
	return s.GetRenewalTokenFunc()
}

func Test_GetRenewalToken_IncompatibleInterfaces(t *testing.T) {
	res, err := GetRenewalToken(0)
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)

	res, err = GetRenewalToken("some text")
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)

	res, err = GetRenewalToken(&SimpleNothingProvider{})
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)

	res, err = GetRenewalToken(func() string { return "token" })
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)
}

func Test_GetRenewalToken_CompatibleInterface(t *testing.T) {
	expectedToken := "valid-token"
	prov := &SimpleSessionRenewalProvider{GetRenewalTokenFunc: func() string { return expectedToken }}

	res, err := GetRenewalToken(prov)
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, res, "Must match")
}

type SimpleSessionOwnerIDProvider struct {
	GetOwnerIDFunc func() string
}

func (s *SimpleSessionOwnerIDProvider) GetOwnerID() string {
	return s.GetOwnerIDFunc()
}

func Test_GetOwnerID_IncompatibleInterfaces(t *testing.T) {
	res, err := GetOwnerID(0)
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)

	res, err = GetOwnerID("some text")
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)

	res, err = GetOwnerID(&SimpleNothingProvider{})
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)

	res, err = GetOwnerID(func() string { return "token" })
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)
}

func Test_GetOwnerID_CompatibleInterface(t *testing.T) {
	expectedOwnerID := "valid-owned-id"
	prov := &SimpleSessionOwnerIDProvider{GetOwnerIDFunc: func() string { return expectedOwnerID }}

	res, err := GetOwnerID(prov)
	assert.NoError(t, err)
	assert.Equal(t, expectedOwnerID, res, "Must match")
}

type SimpleSessionExpiryProvider struct {
	GetExpiryFunc func() time.Time
}

func (s *SimpleSessionExpiryProvider) GetExpiry() time.Time {
	return s.GetExpiryFunc()
}

func Test_GetExpiryToken_IncompatibleInterfaces(t *testing.T) {
	res, err := GetExpiry(0)
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)

	res, err = GetExpiry("some text")
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)

	res, err = GetExpiry(&SimpleNothingProvider{})
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)

	res, err = GetExpiry(func() string { return "token" })
	assert.Empty(t, res, "Must be empty")
	assert.Error(t, err)
}

func Test_GetExpiryToken_CompatibleInterface(t *testing.T) {
	expectedExpiration := time.Now().Add(1234)
	prov := &SimpleSessionExpiryProvider{GetExpiryFunc: func() time.Time { return expectedExpiration }}

	res, err := GetExpiry(prov)
	assert.NoError(t, err)
	assert.Equal(t, expectedExpiration, res, "Must match")
}
