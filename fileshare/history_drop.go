//go:build drop

package fileshare

func FileshareHistoryImplementation(storagePath string) Storage {
	return NewJsonFile(storagePath)
}
