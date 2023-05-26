package fileshare

import "os/user"

type OsInfo interface {
	CurrentUser() (*user.User, error)
	GetGroupIds(*user.User) ([]string, error)
}

type StdOsInfo struct{}

func (StdOsInfo) CurrentUser() (*user.User, error) {
	return user.Current()
}

func (StdOsInfo) GetGroupIds(user *user.User) ([]string, error) {
	return user.GroupIds()
}
