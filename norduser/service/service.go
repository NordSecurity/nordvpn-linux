package service

type Service interface {
	Enable(uid uint32, gid uint32, home string) error
	Stop(uid uint32, wait bool) error
	StopAll()
	Restart(uid uint32) error
}
