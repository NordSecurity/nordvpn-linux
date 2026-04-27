package core

type DevicesRequest struct {
	HardwareIdentifier string `json:"hardware_identifier"`
	PublicKey          string `json:"public_key"`
	Os                 string `json:"os"`
	Type               string `json:"type"`
	Name               string `json:"name"`
}

type UpdateDeviceRequest struct {
	PublicKey string `json:"public_key"`
	Name      string `json:"name"`
}

type DevicesResponse struct {
	UUID               string `json:"uuid"`
	HardwareIdentifier string `json:"hardware_identifier"`
	PublicKey          string `json:"public_key"`
	OS                 string `json:"os"`
	Type               string `json:"type"`
	Name               string `json:"name"`
}
