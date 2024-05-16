package service

type NorduserService interface {
	Enable(uid uint32, gid uint32, home string) error
	Stop(uid uint32, wait bool) error
	StopAll()
	Restart(uid uint32) error
}
