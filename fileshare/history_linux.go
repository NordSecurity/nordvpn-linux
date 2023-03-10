//go:build !drop

package fileshare

func FileshareHistoryImplementation() Storage {
	return MockStorage{}
}
