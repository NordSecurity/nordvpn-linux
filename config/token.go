package config

type TokenData struct {
	Token              string `json:"token,omitempty"`
	TokenExpiry        string `json:"token_expiry,omitempty"`
	RenewToken         string `json:"renew_token,omitempty"`
	ServiceExpiry      string `json:"service_expiry,omitempty"`
	NordLynxPrivateKey string `json:"nordlynx_private_key"`
	OpenVPNUsername    string `json:"openvpn_username"`
	OpenVPNPassword    string `json:"openvpn_password"`
	NCData             NCData `json:"nc_data,omitempty"`
}
