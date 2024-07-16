package fileshare

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

type EventCallback interface {
	OnEvent(event Event)
}

type Event struct {
	Kind      EventKind
	Timestamp int64
}

type EventKind interface{}

type EventKindRequestReceived struct {
	Peer       string
	TransferId string
	Files      []ReceivedFile
}

type ReceivedFile struct {
	Id   string
	Path string
	Size uint64
}

type EventKindRequestQueued struct {
	Peer       string
	TransferId string
	Files      []QueuedFile
}

type QueuedFile struct {
	BaseDir *string
	Id      string
	Path    string
	Size    uint64
}

type EventKindFileStarted struct {
	TransferId  string
	FileId      string
	Transferred uint64
}

type EventKindFileProgress struct {
	TransferId  string
	FileId      string
	Transferred uint64
}

type EventKindTransferFailed struct {
	Status     Status
	TransferId string
}

type Status struct {
	OsErrorCode *int32
	Status      StatusCode
}

type EventKindTransferFinalized struct {
	TransferId string
	ByPeer     bool
}

type EventKindFileDownloaded struct {
	TransferId string
	FileId     string
	FinalPath  string
}

type EventKindFileUploaded struct {
	TransferId string
	FileId     string
}

type EventKindFileRejected struct {
	TransferId string
	FileId     string
	ByPeer     bool
}

type EventKindFileFailed struct {
	TransferId string
	FileId     string
	Status     Status
}

type EventKindUnknown struct{}

var defaultMarshaler = jsonMarshaler{}

func EventToString(event Event) string {
	return eventToString(event, defaultMarshaler)
}

func eventToString(event Event, m marshaler) string {
	json, err := m.Marshal(event)
	if err != nil {
		log.Printf(internal.WarningPrefix+" failed to marshall event: %T, returning just its type\n", event.Kind)
		return fmt.Sprintf("%T", event.Kind)
	}
	return string(json)
}

type marshaler interface {
	Marshal(v any) ([]byte, error)
}

type jsonMarshaler struct{}

func (jm jsonMarshaler) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
