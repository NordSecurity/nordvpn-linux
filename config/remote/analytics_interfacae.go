package remote

// Analytics defines an interface for reporting various analytics events related to downloads,
// local usage, JSON parsing, and partial rollouts. Implementations of this interface are
// responsible for handling the notification logic for each event type.
type Analytics interface {
	NotifyDownload(string, string)
	NotifyDownloadFailure(string, string, DownloadError)
	NotifyLocalUse(string, string, error)
	NotifyJsonParse(string, string, error)
	NotifyPartialRollout(string, string, int, bool)
}
