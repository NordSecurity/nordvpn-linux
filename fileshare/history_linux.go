//go:build !drop

package fileshare

func FileshareHistoryImplementation(_ string) Storage {
	return MockStorage{}
}
