package remote

type Analytics interface {
	NotifyDownload(string, string, error)
	NotifyLocalUse(string, string, error)
	NotifyJsonParse(string, string, error)
	NotifyPartialRollout(string, string, int, bool)
}
