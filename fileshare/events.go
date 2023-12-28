package fileshare

import (
	"encoding/json"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
)

type eventType string

const (
	requestReceived  eventType = "RequestReceived"
	requestQueued    eventType = "RequestQueued"
	transferStarted  eventType = "TransferStarted"
	transferProgress eventType = "TransferProgress"
	transferFinished eventType = "TransferFinished"
)

type finishReason string

const (
	transferFailed   finishReason = "TransferFailed"
	transferCanceled finishReason = "TransferCanceled"
	fileDownloaded   finishReason = "FileDownloaded"
	fileUploaded     finishReason = "FileUploaded"
	fileCanceled     finishReason = "FileCanceled"
	fileFailed       finishReason = "FileFailed"
	fileRejected     finishReason = "FileRejected"
)

type genericEvent struct {
	Type eventType
	Data json.RawMessage // This includes more specific events
}

type requestReceivedEvent struct {
	TransferID string `json:"transfer"`
	Peer       string
	Files      []*pb.File
}

type transferProgressEvent struct {
	TransferID  string `json:"transfer"`
	FileID      string `json:"file"`
	Transferred uint64 `json:"transfered"` // nolint:misspell // We receive this json from the library
}

type transferFinishedEvent struct {
	TransferID string                  `json:"transfer"`
	Reason     finishReason            `json:"reason"`
	Data       transferFinshedSubEvent `json:"data"` // sub-event data
}

type transferFinshedSubEvent struct {
	File   string    `json:"file"`
	Status pb.Status `json:"status"`
	ByPeer bool      `json:"by_peer"`
}
