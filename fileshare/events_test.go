package fileshare

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventToStringWithWorkingMarshaller(t *testing.T) {
	event := Event{
		Kind: EventKindRequestQueued{
			Peer:       "12.12.12.12",
			TransferId: "c13c619c-c70b-49b8-9396-72de88155c43",
			Files: []QueuedFile{
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
	event := Event{
		Kind: EventKindRequestQueued{
			Peer:       "12.12.12.12",
			TransferId: "c13c619c-c70b-49b8-9396-72de88155c43",
			Files: []QueuedFile{
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

	expected := "fileshare.EventKindRequestQueued"

	assert.Equal(t, expected, eventToString(event, brokenMarshaler{}))
}

type brokenMarshaler struct{}

func (bm brokenMarshaler) Marshal(v any) ([]byte, error) {
	return nil, fmt.Errorf("broken marshaler")
}
