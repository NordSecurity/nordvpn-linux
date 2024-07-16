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

const (
	directionIncoming = "incoming"
	exampleFileDir1   = "/tmp/dir/file1"
	exampleFileDir3   = "dir/file3"
)

func TestLibdropTransferToInternalTransfer(t *testing.T) {
	timestamp := timestamppb.New(time.Now().Truncate(time.Millisecond))
	timestampUnix := timestamp.AsTime().UnixMilli()

	// Outgoing transfer request with 3 files
	// Defined as a function so that every test would get a clean newly initialized struct
	getBasicLibdropTransfer := func() LibdropTransfer {
		return LibdropTransfer{
			Id:        "transfer_id",
			CreatedAt: timestampUnix,
			Peer:      "1.2.3.4",
			States:    []TransferState{},
			Kind: TransferKindOutgoing{
				Paths: []OutgoingPath{
					{
						FileId:       "file1_id",
						RelativePath: "dir/file1",
						Bytes:        10,
						Source: OutgoingFileSourceBasePath{
							BasePath: "/tmp",
						},
						States: []OutgoingPathState{},
					},
					{
						FileId:       "file2_id",
						RelativePath: "dir/file2",
						Bytes:        100,
						Source: OutgoingFileSourceBasePath{
							BasePath: "/tmp",
						},
						States: []OutgoingPathState{},
					},
					{
						FileId:       "file3_id",
						RelativePath: exampleFileDir3,
						Bytes:        1000,
						Source: OutgoingFileSourceBasePath{
							BasePath: "/tmp",
						},
						States: []OutgoingPathState{},
					},
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
					FullPath:    exampleFileDir1,
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
					Path:        exampleFileDir3,
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
				in.Kind = TransferKindIncoming{
					Paths: []IncomingPath{
						{
							FileId:       "file1_id",
							RelativePath: "dir/file1",
							Bytes:        10,
							States:       []IncomingPathState{},
						},

						{
							FileId:       "file2_id",
							RelativePath: "dir/file2",
							Bytes:        100,
							States:       []IncomingPathState{},
						},
						{
							FileId:       "file3_id",
							RelativePath: exampleFileDir3,
							Bytes:        1000,
							States:       []IncomingPathState{},
						},
					},
				}
			},
			out: func(out *pb.Transfer) {
				out.Direction = pb.Direction_INCOMING
				// Path is unknown until request is accepted
				out.Path = ""
				// Only a relative path is shown for files
				out.Files[0].FullPath = "dir/file1"
				out.Files[1].FullPath = "dir/file2"
				out.Files[2].FullPath = exampleFileDir3
			},
		},
		{
			name: "incoming transfer in progress with one cancelled file",
			in: func(in *LibdropTransfer) {
				in.Kind = TransferKindIncoming{
					Paths: []IncomingPath{
						{
							FileId:       "file1_id",
							RelativePath: "dir/file1",
							Bytes:        10,
							States: []IncomingPathState{
								{
									Kind: IncomingPathStateKindPending{
										BaseDir: "/tmp",
									},
								},
								{
									Kind: IncomingPathStateKindStarted{},
								},
							},
						},
						{
							FileId:       "file2_id",
							RelativePath: "dir/file2",
							Bytes:        100,
							States: []IncomingPathState{
								{
									Kind: IncomingPathStateKindPending{
										BaseDir: "/tmp",
									},
								},
								{
									Kind: IncomingPathStateKindStarted{},
								},
							},
						},
						{
							FileId:       "file3_id",
							RelativePath: "dir/file3",
							Bytes:        1000,
							States: []IncomingPathState{
								{
									Kind: IncomingPathStateKindRejected{
										ByPeer: false,
									},
								},
							},
						},
					},
				}
			},
			out: func(out *pb.Transfer) {
				out.Direction = pb.Direction_INCOMING
				out.Path = "/tmp"
				out.Status = pb.Status_ONGOING
				out.TotalSize = out.Files[0].Size + out.Files[1].Size
				out.Files[0].FullPath = exampleFileDir1
				out.Files[0].Status = pb.Status_ONGOING
				out.Files[1].FullPath = "/tmp/dir/file2"
				out.Files[1].Status = pb.Status_ONGOING
				out.Files[2].FullPath = exampleFileDir3
				out.Files[2].Status = pb.Status_CANCELED
			},
		},
		{
			name: "outgoing one file transfer with error",
			in: func(in *LibdropTransfer) {
				in.Kind = TransferKindOutgoing{
					Paths: []OutgoingPath{
						{
							FileId:       "file1_id",
							RelativePath: "file1",
							Bytes:        10,
							Source: OutgoingFileSourceBasePath{
								BasePath: "/tmp/dir",
							},
							States: []OutgoingPathState{
								{
									Kind: OutgoingPathStateKindStarted{},
								},
								{
									Kind: OutgoingPathStateKindFailed{
										Status: 33,
									},
								},
							},
						},
					},
				}
			},
			out: func(out *pb.Transfer) {
				out.Status = pb.Status_FINISHED_WITH_ERRORS
				out.TotalSize = 0 // Only one file and it errored out
				out.Path = exampleFileDir1
				out.Files = []*pb.File{out.Files[0]}
				out.Files[0].Path = "file1"
				out.Files[0].Status = pb.Status_FILE_CHECKSUM_MISMATCH
			},
		},
		{
			name: "finalized outgoing transfer with one rejected file",
			in: func(in *LibdropTransfer) {
				in.States = []TransferState{
					{
						Kind: TransferStateKindCancel{
							ByPeer: true,
						},
					},
				}
				in.Kind = TransferKindOutgoing{
					Paths: []OutgoingPath{
						{
							FileId:       "file1_id",
							RelativePath: "dir/file1",
							Bytes:        10,
							Source: OutgoingFileSourceBasePath{
								BasePath: "/tmp",
							},
							States: []OutgoingPathState{
								{
									Kind: OutgoingPathStateKindStarted{
										BytesSent: 10,
									},
								},
								{
									Kind: OutgoingPathStateKindCompleted{},
								},
							},
						},
						{
							FileId:       "file2_id",
							RelativePath: "dir/file2",
							Bytes:        100,
							Source: OutgoingFileSourceBasePath{
								BasePath: "/tmp",
							},
							States: []OutgoingPathState{
								{
									Kind: OutgoingPathStateKindStarted{
										BytesSent: 100,
									},
								},
								{
									Kind: OutgoingPathStateKindCompleted{},
								},
							},
						},
						{
							FileId:       "file3_id",
							RelativePath: "dir/file3",
							Bytes:        1000,
							Source: OutgoingFileSourceBasePath{
								BasePath: "/tmp",
							},
							States: []OutgoingPathState{
								{
									Kind: OutgoingPathStateKindRejected{},
								},
							},
						},
					},
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
				in.States = []TransferState{
					{
						Kind: TransferStateKindCancel{
							ByPeer: true,
						},
					},
				}
				in.Kind = TransferKindOutgoing{
					Paths: []OutgoingPath{
						{
							FileId:       "file1_id",
							RelativePath: "dir/file1",
							Bytes:        10,
							Source: OutgoingFileSourceBasePath{
								BasePath: "/tmp",
							},
							States: []OutgoingPathState{
								{
									Kind: OutgoingPathStateKindStarted{},
								},
							},
						},
						{
							FileId:       "file2_id",
							RelativePath: "dir/file2",
							Bytes:        100,
							Source: OutgoingFileSourceBasePath{
								BasePath: "/tmp",
							},
							States: []OutgoingPathState{
								{
									Kind: OutgoingPathStateKindStarted{},
								},
							},
						},
						{
							FileId:       "file3_id",
							RelativePath: "dir/file3",
							Bytes:        1000,
							Source: OutgoingFileSourceBasePath{
								BasePath: "/tmp",
							},
							States: []OutgoingPathState{
								{
									Kind: OutgoingPathStateKindRejected{},
								},
							},
						},
					},
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
				in.Kind = TransferKindIncoming{
					Paths: []IncomingPath{
						{
							FileId:       "file1_id",
							RelativePath: "dir/file1",
							Bytes:        10,
							States: []IncomingPathState{
								{
									Kind: IncomingPathStateKindPending{
										BaseDir: "/tmp",
									},
								},
								{
									Kind: IncomingPathStateKindStarted{},
								},
								{
									Kind: IncomingPathStateKindPaused{},
								},
								{
									Kind: IncomingPathStateKindStarted{
										BytesReceived: 5,
									},
								},
							},
						},
						{
							FileId:       "file2_id",
							RelativePath: "dir/file2",
							Bytes:        100,
							States: []IncomingPathState{
								{
									Kind: IncomingPathStateKindPending{
										BaseDir: "/tmp",
									},
								},
								{
									Kind: IncomingPathStateKindStarted{},
								},
								{
									Kind: IncomingPathStateKindPaused{},
								},
								{
									Kind: IncomingPathStateKindStarted{
										BytesReceived: 5,
									},
								},
								{
									Kind: IncomingPathStateKindCompleted{
										FinalPath: "/tmp/dir/file2_(1)",
									},
								},
							},
						},
						{
							FileId:       "file3_id",
							RelativePath: "dir/file3",
							Bytes:        1000,
							States:       []IncomingPathState{},
						},
					},
				}
			},
			out: func(out *pb.Transfer) {
				out.Direction = pb.Direction_INCOMING
				out.Path = "/tmp"
				out.Status = pb.Status_ONGOING
				out.TotalTransferred = 5 + out.Files[1].Size
				out.Files[0].FullPath = exampleFileDir1
				out.Files[0].Status = pb.Status_ONGOING
				out.Files[0].Transferred = 5
				out.Files[1].FullPath = "/tmp/dir/file2_(1)"
				out.Files[1].Status = pb.Status_SUCCESS
				out.Files[1].Transferred = out.Files[1].Size
				out.Files[2].FullPath = exampleFileDir3
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
			result := InternalTransferToPBTransfer(in)

			assert.True(t, proto.Equal(expected, result), fmt.Sprintf("\nExpect: %+v\n\nActual: %+v",
				expected, result))
		})
	}
}
