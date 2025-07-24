package remote

import "fmt"

const (
	debuggerEventBaseKey = "remote-config"

	// defaultMaxGroup represents the maximum value for a rollout group,
	// effectively making the value to be in range of 1-100 (inclusive) to reflect percentage-based groups.
	defaultMaxGroup  uint32 = 100
	logPrefix               = "[Remote Config]"
	messageNamespace        = "nordvpn-linux"
	rcFailure               = "failure"
	rcSuccess               = "success"
	subscope                = "linux-rc"
)

// EventType defines the type of remote config analytics event.
type EventType int

const (
	Download EventType = iota
	DownloadSuccess
	DownloadFailure
	LocalUse
	JSONParseSuccess
	JSONParseFailure
	PartialRollout
)

func (e EventType) String() string {
	switch e {
	case Download:
		return "rc_download"
	case DownloadSuccess:
		return "rc_download_success"
	case DownloadFailure:
		return "rc_download_failure"
	case LocalUse:
		return "rc_local_use"
	case JSONParseSuccess:
		return "rc_json_parse_success"
	case JSONParseFailure:
		return "rc_json_parse_failure"
	case PartialRollout:
		return "rc_partial_rollout"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}

// DownloadErrorKind defines types of download errors for remote config.
type DownloadErrorKind int

const (
	DownloadErrorRemoteHashNotFound DownloadErrorKind = iota
	DownloadErrorRemoteFileNotFound
	DownloadErrorIntegrity
	DownloadErrorFileDownload
	DownloadErrorNetwork
	DownloadErrorOther
)

func (e DownloadErrorKind) String() string {
	switch e {
	case DownloadErrorRemoteHashNotFound:
		return "remote_hash_not_found"
	case DownloadErrorRemoteFileNotFound:
		return "remote_file_not_found"
	case DownloadErrorIntegrity:
		return "integrity_error"
	case DownloadErrorFileDownload:
		return "file_download_error"
	case DownloadErrorNetwork:
		return "network_error"
	case DownloadErrorOther:
		return "other_error"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}

// FeatureName defines the name of a feature for remote config.
type FeatureName int

const (
	FeatureMain FeatureName = iota
	FeatureLibtelio
	FeatureMeshnet
)

func (f FeatureName) String() string {
	switch f {
	case FeatureMain:
		return "nordvpn"
	case FeatureLibtelio:
		return "libtelio"
	case FeatureMeshnet:
		return "meshnet"
	default:
		return fmt.Sprintf("%d", int(f))
	}
}
