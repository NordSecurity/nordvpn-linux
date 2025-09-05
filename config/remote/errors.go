package remote

import "fmt"

// DownloadErrorKind defines types of download errors for remote config.
type DownloadErrorKind int

const (
	DownloadErrorFileDownload DownloadErrorKind = iota
	DownloadErrorFileRename
	DownloadErrorHashIntegrity
	DownloadErrorHashParsing
	DownloadErrorIncludeFile
	DownloadErrorIntegrity
	DownloadErrorLocalFS
	DownloadErrorOther
	DownloadErrorParsing
	DownloadErrorRemoteFileNotFound
	DownloadErrorRemoteHashNotFound
	DownloadErrorWriteHash
	DownloadErrorWriteJson
)

func (e DownloadErrorKind) String() string {
	switch e {
	case DownloadErrorFileDownload:
		return "file_download_error"
	case DownloadErrorFileRename:
		return "file_rename_error"
	case DownloadErrorHashIntegrity:
		return "hash_integrity_error"
	case DownloadErrorHashParsing:
		return "hash_parsing_error"
	case DownloadErrorIncludeFile:
		return "include_file_error"
	case DownloadErrorIntegrity:
		return "integrity_error"
	case DownloadErrorLocalFS:
		return "local_fs_error"
	case DownloadErrorOther:
		return "other_error"
	case DownloadErrorParsing:
		return "parsing_error"
	case DownloadErrorRemoteFileNotFound:
		return "remote_file_not_found"
	case DownloadErrorRemoteHashNotFound:
		return "remote_hash_not_found"
	case DownloadErrorWriteHash:
		return "write_hash_error"
	case DownloadErrorWriteJson:
		return "write_json_error"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}

// LoadErrorKind defines types of load errors for remote config.
type LoadErrorKind int

const (
	LoadErrorFileNotFound LoadErrorKind = iota
	LoadErrorParsing
	LoadErrorFieldValidation
	LoadErrorMainHashJsonParsing
	LoadErrorIntegrity
	LoadErrorValidation
	LoadErrorMainJsonValidationFailure
	LoadErrorIncludeFile
	LoadErrorOther
)

func (e LoadErrorKind) String() string {
	switch e {
	case LoadErrorFileNotFound:
		return "file_not_found_error"
	case LoadErrorParsing:
		return "parsing_error"
	case LoadErrorFieldValidation:
		return "config_field_validation_error"
	case LoadErrorMainHashJsonParsing:
		return "main_hash_json_parsing_error"
	case LoadErrorIntegrity:
		return "integrity_error"
	case LoadErrorValidation:
		return "validation_error"
	case LoadErrorMainJsonValidationFailure:
		return "main_json_validation_failure"
	case LoadErrorIncludeFile:
		return "include_file_error"
	case LoadErrorOther:
		return "other_error"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}

// AnalyticsError is a generic error type for analytics-related errors.
type AnalyticsError[T fmt.Stringer] struct {
	Kind  T
	Cause error
}

func (e *AnalyticsError[T]) Error() string {
	return fmt.Sprintf("%s: %v", e.Kind, e.Cause)
}

func (e *AnalyticsError[T]) Unwrap() error {
	return e.Cause
}

// DownloadError holds information about a specific failure related to downloading remote configuration.
type DownloadError struct {
	AnalyticsError[DownloadErrorKind]
}

// LoadError holds information about a specific failure related to loading remote configuration.
type LoadError struct {
	AnalyticsError[LoadErrorKind]
}

func NewDownloadError(kind DownloadErrorKind, err error) *DownloadError {
	return &DownloadError{
		AnalyticsError: AnalyticsError[DownloadErrorKind]{
			Kind:  kind,
			Cause: err,
		},
	}
}

func NewLoadError(kind LoadErrorKind, err error) *LoadError {
	return &LoadError{
		AnalyticsError: AnalyticsError[LoadErrorKind]{
			Kind:  kind,
			Cause: err,
		},
	}
}
