package fileshare

import (
	"fmt"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLibdropTransferToInternalTransfer(t *testing.T) {
	var timestamp = timestamppb.New(time.Now().Truncate(time.Millisecond))
	var timestampUnix = timestamp.AsTime().UnixMilli()

	// Outgoing transfer request with 3 files
	// Defined as a function so that every test would get a clean newly initialized struct
	getBasicLibdropTransfer := func() LibdropTransfer {
		return LibdropTransfer{
			ID:        "transfer_id",
			Peer:      "1.2.3.4",
			CreatedAt: timestampUnix,
			Direction: "outgoing",
			States:    []LibdropTransferState{},
			Files: []LibdropFile{
				{
					ID:           "file1_id",
					TransferID:   "transfer_id",
					BasePath:     "/tmp",
					RelativePath: "dir/file1",
					TotalSize:    10,
					CreatedAt:    uint64(timestampUnix),
					States:       []LibdropFileState{},
				},
				{
					ID:           "file2_id",
					TransferID:   "transfer_id",
					BasePath:     "/tmp",
					RelativePath: "dir/file2",
					TotalSize:    100,
					CreatedAt:    uint64(timestampUnix),
					States:       []LibdropFileState{},
				},
				{
					ID:           "file3_id",
					TransferID:   "transfer_id",
					BasePath:     "/tmp",
					RelativePath: "dir/file3",
					TotalSize:    1000,
					CreatedAt:    uint64(timestampUnix),
					States:       []LibdropFileState{},
				},
			},
		}
	}

	// Result of converting getBasicLibdropTransfer()
	getBasicTransfer := func() *pb.Transfer {
		return &pb.Transfer{
			Id:               "transfer_id",
			Direction:        pb.Direction_OUTGOING,
			Peer:             "1.2.3.4",
			Status:           pb.Status_REQUESTED,
			Created:          timestamp,
			Path:             "/tmp/dir",
			TotalSize:        1110,
			TotalTransferred: 0,
			Files: []*pb.File{
				{
					Id:          "file1_id",
					Path:        "dir/file1",
					FullPath:    "/tmp/dir/file1",
					Size:        10,
					Transferred: 0,
					Status:      pb.Status_REQUESTED,
				},
				{
					Id:          "file2_id",
					Path:        "dir/file2",
					FullPath:    "/tmp/dir/file2",
					Size:        100,
					Transferred: 0,
					Status:      pb.Status_REQUESTED,
				},
				{
					Id:          "file3_id",
					Path:        "dir/file3",
					FullPath:    "/tmp/dir/file3",
					Size:        1000,
					Transferred: 0,
					Status:      pb.Status_REQUESTED,
				},
			},
		}
	}

	tests := []struct {
		name string
		in   func(*LibdropTransfer)
		out  func(*pb.Transfer)
	}{
		{
			name: "outgoing transfer request",
			in:   func(in *LibdropTransfer) {},
			out:  func(out *pb.Transfer) {},
		},
		{
			name: "incoming transfer request",
			in: func(in *LibdropTransfer) {
				in.Direction = "incoming"
				in.Files[0].BasePath = ""
				in.Files[1].BasePath = ""
				in.Files[2].BasePath = ""
			},
			out: func(out *pb.Transfer) {
				out.Direction = pb.Direction_INCOMING
				// Path is unknown until request is accepted
				out.Path = ""
				// Only a relative path is shown for files
				out.Files[0].FullPath = "dir/file1"
				out.Files[1].FullPath = "dir/file2"
				out.Files[2].FullPath = "dir/file3"
			},
		},
		{
			name: "incoming transfer in progress with one cancelled file",
			in: func(in *LibdropTransfer) {
				in.Direction = "incoming"
				in.Files[0].BasePath = ""
				in.Files[0].States = []LibdropFileState{
					{State: "pending", BasePath: "/tmp"},
					{State: "started"},
				}
				in.Files[1].BasePath = ""
				in.Files[1].States = []LibdropFileState{
					{State: "pending", BasePath: "/tmp"},
					{State: "started"},
				}
				in.Files[2].BasePath = ""
				in.Files[2].States = []LibdropFileState{
					{State: "rejected"},
				}
			},
			out: func(out *pb.Transfer) {
				out.Direction = pb.Direction_INCOMING
				out.Path = "/tmp"
				out.Status = pb.Status_ONGOING
				out.TotalSize = out.Files[0].Size + out.Files[1].Size
				out.Files[0].FullPath = "/tmp/dir/file1"
				out.Files[0].Status = pb.Status_ONGOING
				out.Files[1].FullPath = "/tmp/dir/file2"
				out.Files[1].Status = pb.Status_ONGOING
				out.Files[2].FullPath = "dir/file3"
				out.Files[2].Status = pb.Status_CANCELED
			},
		},
		{
			name: "outgoing one file transfer with error",
			in: func(in *LibdropTransfer) {
				in.Files = []LibdropFile{in.Files[0]}
				in.Files[0].BasePath = "/tmp/dir"
				in.Files[0].RelativePath = "file1"
				in.Files[0].States = []LibdropFileState{
					{State: "started"},
					{State: "failed", StatusCode: 33},
				}
			},
			out: func(out *pb.Transfer) {
				out.Status = pb.Status_FINISHED_WITH_ERRORS
				out.TotalSize = 0 // Only one file and it errored out
				out.Path = "/tmp/dir/file1"
				out.Files = []*pb.File{out.Files[0]}
				out.Files[0].Path = "file1"
				out.Files[0].Status = pb.Status_FILE_CHECKSUM_MISMATCH
			},
		},
		{
			name: "finalized outgoing transfer with one rejected file",
			in: func(in *LibdropTransfer) {
				in.States = []LibdropTransferState{
					{State: "cancel", ByPeer: true},
				}
				in.Files[0].States = []LibdropFileState{
					{State: "started"},
					{State: "completed"},
				}
				in.Files[1].States = []LibdropFileState{
					{State: "started"},
					{State: "completed"},
				}
				in.Files[2].States = []LibdropFileState{
					{State: "rejected"},
				}
			},
			out: func(out *pb.Transfer) {
				out.Status = pb.Status_SUCCESS
				out.TotalSize = out.Files[0].Size + out.Files[1].Size
				out.TotalTransferred = out.Files[0].Size + out.Files[1].Size
				out.Files[0].Status = pb.Status_SUCCESS
				out.Files[0].Transferred = out.Files[0].Size
				out.Files[1].Status = pb.Status_SUCCESS
				out.Files[1].Transferred = out.Files[1].Size
				out.Files[2].Status = pb.Status_CANCELED
			},
		},
		{
			name: "cancelled outgoing transfer with one rejected file",
			in: func(in *LibdropTransfer) {
				in.States = []LibdropTransferState{
					{State: "cancel", ByPeer: true},
				}
				in.Files[0].States = []LibdropFileState{
					{State: "started"},
				}
				in.Files[1].States = []LibdropFileState{
					{State: "started"},
				}
				in.Files[2].States = []LibdropFileState{
					{State: "rejected"},
				}
			},
			out: func(out *pb.Transfer) {
				out.Status = pb.Status_CANCELED_BY_PEER
				out.TotalSize = out.Files[0].Size + out.Files[1].Size
				out.Files[0].Status = pb.Status_CANCELED
				out.Files[1].Status = pb.Status_CANCELED
				out.Files[2].Status = pb.Status_CANCELED
			},
		},
		{
			name: "incoming transfer in progress with some finished files",
			in: func(in *LibdropTransfer) {
				in.Direction = "incoming"
				in.Files[0].BasePath = ""
				in.Files[1].BasePath = ""
				in.Files[2].BasePath = ""
				in.Files[0].States = []LibdropFileState{
					{State: "pending", BasePath: "/tmp"},
					{State: "started"},
					{State: "paused"},
					{State: "started", BytesReceived: 5},
				}
				in.Files[1].States = []LibdropFileState{
					{State: "pending", BasePath: "/tmp"},
					{State: "started"},
					{State: "paused"},
					{State: "started", BytesReceived: 5},
					{State: "completed", FinalPath: "/tmp/dir/file2_(1)"},
				}
				in.Files[2].States = []LibdropFileState{}
			},
			out: func(out *pb.Transfer) {
				out.Direction = pb.Direction_INCOMING
				out.Path = "/tmp"
				out.Status = pb.Status_ONGOING
				out.TotalTransferred = 5 + out.Files[1].Size
				out.Files[0].FullPath = "/tmp/dir/file1"
				out.Files[0].Status = pb.Status_ONGOING
				out.Files[0].Transferred = 5
				out.Files[1].FullPath = "/tmp/dir/file2_(1)"
				out.Files[1].Status = pb.Status_SUCCESS
				out.Files[1].Transferred = out.Files[1].Size
				out.Files[2].FullPath = "dir/file3"
				out.Files[2].Status = pb.Status_REQUESTED
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in := getBasicLibdropTransfer()
			expected := getBasicTransfer()
			test.in(&in)
			test.out(expected)
			result := LibdropTransferToInternalTransfer(in)

			assert.True(t, proto.Equal(expected, result), fmt.Sprintf("\nExpect: %+v\n\nActual: %+v",
				expected, result))
		})
	}
}
