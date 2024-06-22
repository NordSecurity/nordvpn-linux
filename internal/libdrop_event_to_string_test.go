package internal

import (
	"fmt"
	"testing"

	norddrop "github.com/NordSecurity/libdrop-go/v7"
	"github.com/stretchr/testify/assert"
)

func TestEventToStringWithWorkingMarshaller(t *testing.T) {
	event := norddrop.Event{
		Kind: norddrop.EventKindRequestQueued{
			Peer:       "12.12.12.12",
			TransferId: "c13c619c-c70b-49b8-9396-72de88155c43",
			Files: []norddrop.QueuedFile{
				{
					Id:   "file1ID",
					Path: "testfile-small",
					Size: 100,
				},
				{
					Id:   "file2ID",
					Path: "testfile-big",
					Size: 1000,
				},
				{
					Id:   "file3ID",
					Path: "file3.txt",
					Size: 1000,
				},
			},
		},
	}

	expected := `{
    "Timestamp": 0,
    "Kind": {
      "Peer": "12.12.12.12",
      "TransferId": "c13c619c-c70b-49b8-9396-72de88155c43",
      "Files": [
        {
          "Id": "file1ID",
          "Path": "testfile-small",
          "Size": 100,
          "BaseDir": null
        },
        {
          "Id": "file2ID",
          "Path": "testfile-big",
          "Size": 1000,
          "BaseDir": null
        },
        {
          "Id": "file3ID",
          "Path": "file3.txt",
          "Size": 1000,
          "BaseDir": null
        }
      ]
    }
  }`

	assert.JSONEq(t, expected, eventToString(event, jsonMarshaler{}))
}

func TestEventToStringWithBrokenMarshaler(t *testing.T) {
	event := norddrop.Event{
		Kind: norddrop.EventKindRequestQueued{
			Peer:       "12.12.12.12",
			TransferId: "c13c619c-c70b-49b8-9396-72de88155c43",
			Files: []norddrop.QueuedFile{
				{
					Id:   "file1ID",
					Path: "testfile-small",
					Size: 100,
				},
				{
					Id:   "file2ID",
					Path: "testfile-big",
					Size: 1000,
				},
				{
					Id:   "file3ID",
					Path: "file3.txt",
					Size: 1000,
				},
			},
		},
	}

	expected := "norddrop.EventKindRequestQueued"

	assert.Equal(t, expected, eventToString(event, brokenMarshaler{}))
}

type brokenMarshaler struct{}

func (bm brokenMarshaler) Marshal(v any) ([]byte, error) {
	return nil, fmt.Errorf("broken marshaler")
}
