package config

import "github.com/google/uuid"

type TokenData struct {
	Token                  string     `json:"token,omitempty"`
	TokenExpiry            string     `json:"token_expiry,omitempty"`
	RenewToken             string     `json:"renew_token,omitempty"`
	IsOAuth                bool       `json:"is_oauth,omitempty"`
	TrustedPassToken       string     `json:"trusted_pass_token,omitempty"`
	TrustedPassOwnerID     string     `json:"trusted_pass_owner_id,omitempty"`
	TrustedPassTokenExpiry string     `json:"trusted_pass_token_expiry,omitempty"`
	ServiceExpiry          string     `json:"service_expiry,omitempty"`
	NordLynxPrivateKey     string     `json:"nordlynx_private_key"`
	OpenVPNUsername        string     `json:"openvpn_username"`
	OpenVPNPassword        string     `json:"openvpn_password"`
	NCData                 NCData     `json:"nc_data,omitempty"`
	IdempotencyKey         *uuid.UUID `json:"idempotency_key,omitempty"`
}
