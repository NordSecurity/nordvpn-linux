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

type DedicatedServers []DedicatedServer

type DedicatedServerStatus string

const (
	DedicatedServerStatusRunning = "running"
)

type DedicatedServer struct {
	UUID     string                `json:"uuid"`
	Name     string                `json:"name"`
	Status   DedicatedServerStatus `json:"status"`
	IP       string                `json:"ip"`
	Location `json:"location"`
}

type ConnectRequest struct {
	DeviceUUID      string `json:"device_uuid"`
	DevicePublicKey string `json:"device_public_key"`
}

type ConnectResponse struct {
	ServerPublicKey string `json:"server_public_key"`
	ServerEndpoint  string `json:"server_endpoint"`
}
