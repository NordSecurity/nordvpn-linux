package vpn

// LibConfigGetter is interface to acquire config for vpn implementation library
type LibConfigGetter interface {
	GetConfig() (string, error)
}
