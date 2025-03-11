// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v3.21.6
// source: protobuf/fileshare/transfer.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Direction int32

const (
	Direction_UNKNOWN_DIRECTION Direction = 0
	Direction_INCOMING          Direction = 1
	Direction_OUTGOING          Direction = 2
)

// Enum value maps for Direction.
var (
	Direction_name = map[int32]string{
		0: "UNKNOWN_DIRECTION",
		1: "INCOMING",
		2: "OUTGOING",
	}
	Direction_value = map[string]int32{
		"UNKNOWN_DIRECTION": 0,
		"INCOMING":          1,
		"OUTGOING":          2,
	}
)

func (x Direction) Enum() *Direction {
	p := new(Direction)
	*p = x
	return p
}

func (x Direction) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Direction) Descriptor() protoreflect.EnumDescriptor {
	return file_protobuf_fileshare_transfer_proto_enumTypes[0].Descriptor()
}

func (Direction) Type() protoreflect.EnumType {
	return &file_protobuf_fileshare_transfer_proto_enumTypes[0]
}

func (x Direction) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Direction.Descriptor instead.
func (Direction) EnumDescriptor() ([]byte, []int) {
	return file_protobuf_fileshare_transfer_proto_rawDescGZIP(), []int{0}
}

type Status int32

const (
	// Libdrop statuses for finished transfers
	Status_SUCCESS                  Status = 0
	Status_CANCELED                 Status = 1
	Status_BAD_PATH                 Status = 2
	Status_BAD_FILE                 Status = 3
	Status_TRANSPORT                Status = 4 // unused
	Status_BAD_STATUS               Status = 5 // unused
	Status_SERVICE_STOP             Status = 6
	Status_BAD_TRANSFER             Status = 7
	Status_BAD_TRANSFER_STATE       Status = 8
	Status_BAD_FILE_ID              Status = 9
	Status_BAD_SYSTEM_TIME          Status = 10 // unused
	Status_TRUNCATED_FILE           Status = 11 // unused
	Status_EVENT_SEND               Status = 12 // unused
	Status_BAD_UUID                 Status = 13 // unused
	Status_CHANNEL_CLOSED           Status = 14 // unused
	Status_IO                       Status = 15
	Status_DATA_SEND                Status = 16 // unused
	Status_DIRECTORY_NOT_EXPECTED   Status = 17
	Status_EMPTY_TRANSFER           Status = 18 // unused
	Status_TRANSFER_CLOSED_BY_PEER  Status = 19 // unused
	Status_TRANSFER_LIMITS_EXCEEDED Status = 20
	Status_MISMATCHED_SIZE          Status = 21
	Status_UNEXPECTED_DATA          Status = 22
	Status_INVALID_ARGUMENT         Status = 23 // unused
	Status_TRANSFER_TIMEOUT         Status = 24
	Status_WS_SERVER                Status = 25
	Status_WS_CLIENT                Status = 26
	// UNUSED = 27;
	Status_FILE_MODIFIED          Status = 28
	Status_FILENAME_TOO_LONG      Status = 29
	Status_AUTHENTICATION_FAILED  Status = 30
	Status_FILE_CHECKSUM_MISMATCH Status = 33
	Status_FILE_REJECTED          Status = 34
	// Internally defined statuses for unfinished transfers
	Status_REQUESTED            Status = 100
	Status_ONGOING              Status = 101
	Status_FINISHED_WITH_ERRORS Status = 102
	Status_ACCEPT_FAILURE       Status = 103
	Status_CANCELED_BY_PEER     Status = 104
	Status_INTERRUPTED          Status = 105
	Status_PAUSED               Status = 106
	Status_PENDING              Status = 107
)

// Enum value maps for Status.
var (
	Status_name = map[int32]string{
		0:   "SUCCESS",
		1:   "CANCELED",
		2:   "BAD_PATH",
		3:   "BAD_FILE",
		4:   "TRANSPORT",
		5:   "BAD_STATUS",
		6:   "SERVICE_STOP",
		7:   "BAD_TRANSFER",
		8:   "BAD_TRANSFER_STATE",
		9:   "BAD_FILE_ID",
		10:  "BAD_SYSTEM_TIME",
		11:  "TRUNCATED_FILE",
		12:  "EVENT_SEND",
		13:  "BAD_UUID",
		14:  "CHANNEL_CLOSED",
		15:  "IO",
		16:  "DATA_SEND",
		17:  "DIRECTORY_NOT_EXPECTED",
		18:  "EMPTY_TRANSFER",
		19:  "TRANSFER_CLOSED_BY_PEER",
		20:  "TRANSFER_LIMITS_EXCEEDED",
		21:  "MISMATCHED_SIZE",
		22:  "UNEXPECTED_DATA",
		23:  "INVALID_ARGUMENT",
		24:  "TRANSFER_TIMEOUT",
		25:  "WS_SERVER",
		26:  "WS_CLIENT",
		28:  "FILE_MODIFIED",
		29:  "FILENAME_TOO_LONG",
		30:  "AUTHENTICATION_FAILED",
		33:  "FILE_CHECKSUM_MISMATCH",
		34:  "FILE_REJECTED",
		100: "REQUESTED",
		101: "ONGOING",
		102: "FINISHED_WITH_ERRORS",
		103: "ACCEPT_FAILURE",
		104: "CANCELED_BY_PEER",
		105: "INTERRUPTED",
		106: "PAUSED",
		107: "PENDING",
	}
	Status_value = map[string]int32{
		"SUCCESS":                  0,
		"CANCELED":                 1,
		"BAD_PATH":                 2,
		"BAD_FILE":                 3,
		"TRANSPORT":                4,
		"BAD_STATUS":               5,
		"SERVICE_STOP":             6,
		"BAD_TRANSFER":             7,
		"BAD_TRANSFER_STATE":       8,
		"BAD_FILE_ID":              9,
		"BAD_SYSTEM_TIME":          10,
		"TRUNCATED_FILE":           11,
		"EVENT_SEND":               12,
		"BAD_UUID":                 13,
		"CHANNEL_CLOSED":           14,
		"IO":                       15,
		"DATA_SEND":                16,
		"DIRECTORY_NOT_EXPECTED":   17,
		"EMPTY_TRANSFER":           18,
		"TRANSFER_CLOSED_BY_PEER":  19,
		"TRANSFER_LIMITS_EXCEEDED": 20,
		"MISMATCHED_SIZE":          21,
		"UNEXPECTED_DATA":          22,
		"INVALID_ARGUMENT":         23,
		"TRANSFER_TIMEOUT":         24,
		"WS_SERVER":                25,
		"WS_CLIENT":                26,
		"FILE_MODIFIED":            28,
		"FILENAME_TOO_LONG":        29,
		"AUTHENTICATION_FAILED":    30,
		"FILE_CHECKSUM_MISMATCH":   33,
		"FILE_REJECTED":            34,
		"REQUESTED":                100,
		"ONGOING":                  101,
		"FINISHED_WITH_ERRORS":     102,
		"ACCEPT_FAILURE":           103,
		"CANCELED_BY_PEER":         104,
		"INTERRUPTED":              105,
		"PAUSED":                   106,
		"PENDING":                  107,
	}
)

func (x Status) Enum() *Status {
	p := new(Status)
	*p = x
	return p
}

func (x Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status) Descriptor() protoreflect.EnumDescriptor {
	return file_protobuf_fileshare_transfer_proto_enumTypes[1].Descriptor()
}

func (Status) Type() protoreflect.EnumType {
	return &file_protobuf_fileshare_transfer_proto_enumTypes[1]
}

func (x Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status.Descriptor instead.
func (Status) EnumDescriptor() ([]byte, []int) {
	return file_protobuf_fileshare_transfer_proto_rawDescGZIP(), []int{1}
}

type Transfer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Direction Direction              `protobuf:"varint,2,opt,name=direction,proto3,enum=filesharepb.Direction" json:"direction,omitempty"`
	Peer      string                 `protobuf:"bytes,3,opt,name=peer,proto3" json:"peer,omitempty"`
	Status    Status                 `protobuf:"varint,4,opt,name=status,proto3,enum=filesharepb.Status" json:"status,omitempty"` // Calculated from status of all files in the transfer
	Created   *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=created,proto3" json:"created,omitempty"`
	Files     []*File                `protobuf:"bytes,6,rep,name=files,proto3" json:"files,omitempty"`
	// For outgoing transfers the user provided path to be sent
	// For incoming transfers path where the files will be downloaded to
	Path             string `protobuf:"bytes,7,opt,name=path,proto3" json:"path,omitempty"`
	TotalSize        uint64 `protobuf:"varint,8,opt,name=total_size,json=totalSize,proto3" json:"total_size,omitempty"`
	TotalTransferred uint64 `protobuf:"varint,9,opt,name=total_transferred,json=totalTransferred,proto3" json:"total_transferred,omitempty"`
}

func (x *Transfer) Reset() {
	*x = Transfer{}
	mi := &file_protobuf_fileshare_transfer_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Transfer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transfer) ProtoMessage() {}

func (x *Transfer) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_fileshare_transfer_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transfer.ProtoReflect.Descriptor instead.
func (*Transfer) Descriptor() ([]byte, []int) {
	return file_protobuf_fileshare_transfer_proto_rawDescGZIP(), []int{0}
}

func (x *Transfer) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Transfer) GetDirection() Direction {
	if x != nil {
		return x.Direction
	}
	return Direction_UNKNOWN_DIRECTION
}

func (x *Transfer) GetPeer() string {
	if x != nil {
		return x.Peer
	}
	return ""
}

func (x *Transfer) GetStatus() Status {
	if x != nil {
		return x.Status
	}
	return Status_SUCCESS
}

func (x *Transfer) GetCreated() *timestamppb.Timestamp {
	if x != nil {
		return x.Created
	}
	return nil
}

func (x *Transfer) GetFiles() []*File {
	if x != nil {
		return x.Files
	}
	return nil
}

func (x *Transfer) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *Transfer) GetTotalSize() uint64 {
	if x != nil {
		return x.TotalSize
	}
	return 0
}

func (x *Transfer) GetTotalTransferred() uint64 {
	if x != nil {
		return x.TotalTransferred
	}
	return 0
}

type File struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Path        string `protobuf:"bytes,6,opt,name=path,proto3" json:"path,omitempty"`         // Used for display and user input. Can be just file name, or relative to a dir that is sent.
	FullPath    string `protobuf:"bytes,7,opt,name=fullPath,proto3" json:"fullPath,omitempty"` // Absolute path
	Size        uint64 `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`
	Transferred uint64 `protobuf:"varint,3,opt,name=transferred,proto3" json:"transferred,omitempty"`
	Status      Status `protobuf:"varint,4,opt,name=status,proto3,enum=filesharepb.Status" json:"status,omitempty"` // Received from the events for specific set of files
	// Not used anymore, file lists should always be flat, kept for history file compatibility
	Children map[string]*File `protobuf:"bytes,5,rep,name=children,proto3" json:"children,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *File) Reset() {
	*x = File{}
	mi := &file_protobuf_fileshare_transfer_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *File) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*File) ProtoMessage() {}

func (x *File) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_fileshare_transfer_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use File.ProtoReflect.Descriptor instead.
func (*File) Descriptor() ([]byte, []int) {
	return file_protobuf_fileshare_transfer_proto_rawDescGZIP(), []int{1}
}

func (x *File) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *File) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *File) GetFullPath() string {
	if x != nil {
		return x.FullPath
	}
	return ""
}

func (x *File) GetSize() uint64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *File) GetTransferred() uint64 {
	if x != nil {
		return x.Transferred
	}
	return 0
}

func (x *File) GetStatus() Status {
	if x != nil {
		return x.Status
	}
	return Status_SUCCESS
}

func (x *File) GetChildren() map[string]*File {
	if x != nil {
		return x.Children
	}
	return nil
}

var File_protobuf_fileshare_transfer_proto protoreflect.FileDescriptor

var file_protobuf_fileshare_transfer_proto_rawDesc = []byte{
	0x0a, 0x21, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x73,
	0x68, 0x61, 0x72, 0x65, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x68, 0x61, 0x72, 0x65, 0x70, 0x62,
	0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xd0, 0x02, 0x0a, 0x08, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x34,
	0x0a, 0x09, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x16, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x68, 0x61, 0x72, 0x65, 0x70, 0x62, 0x2e,
	0x44, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x09, 0x64, 0x69, 0x72, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x65, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x70, 0x65, 0x65, 0x72, 0x12, 0x2b, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x73,
	0x68, 0x61, 0x72, 0x65, 0x70, 0x62, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x34, 0x0a, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x52, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x12, 0x27, 0x0a, 0x05, 0x66,
	0x69, 0x6c, 0x65, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x66, 0x69, 0x6c,
	0x65, 0x73, 0x68, 0x61, 0x72, 0x65, 0x70, 0x62, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x05, 0x66,
	0x69, 0x6c, 0x65, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x1d, 0x0a, 0x0a, 0x74, 0x6f, 0x74, 0x61,
	0x6c, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x74, 0x6f,
	0x74, 0x61, 0x6c, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x2b, 0x0a, 0x11, 0x74, 0x6f, 0x74, 0x61, 0x6c,
	0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x72, 0x65, 0x64, 0x18, 0x09, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x10, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65,
	0x72, 0x72, 0x65, 0x64, 0x22, 0xb6, 0x02, 0x0a, 0x04, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a,
	0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74,
	0x68, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x75, 0x6c, 0x6c, 0x50, 0x61, 0x74, 0x68, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x75, 0x6c, 0x6c, 0x50, 0x61, 0x74, 0x68, 0x12, 0x12, 0x0a,
	0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x04, 0x73, 0x69, 0x7a,
	0x65, 0x12, 0x20, 0x0a, 0x0b, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x72, 0x65, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72,
	0x72, 0x65, 0x64, 0x12, 0x2b, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x68, 0x61, 0x72, 0x65, 0x70,
	0x62, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x3b, 0x0a, 0x08, 0x63, 0x68, 0x69, 0x6c, 0x64, 0x72, 0x65, 0x6e, 0x18, 0x05, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x68, 0x61, 0x72, 0x65, 0x70, 0x62,
	0x2e, 0x46, 0x69, 0x6c, 0x65, 0x2e, 0x43, 0x68, 0x69, 0x6c, 0x64, 0x72, 0x65, 0x6e, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x08, 0x63, 0x68, 0x69, 0x6c, 0x64, 0x72, 0x65, 0x6e, 0x1a, 0x4e, 0x0a,
	0x0d, 0x43, 0x68, 0x69, 0x6c, 0x64, 0x72, 0x65, 0x6e, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x27, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x11, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x68, 0x61, 0x72, 0x65, 0x70, 0x62, 0x2e, 0x46, 0x69,
	0x6c, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x2a, 0x3e, 0x0a,
	0x09, 0x44, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x15, 0x0a, 0x11, 0x55, 0x4e,
	0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x5f, 0x44, 0x49, 0x52, 0x45, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x10,
	0x00, 0x12, 0x0c, 0x0a, 0x08, 0x49, 0x4e, 0x43, 0x4f, 0x4d, 0x49, 0x4e, 0x47, 0x10, 0x01, 0x12,
	0x0c, 0x0a, 0x08, 0x4f, 0x55, 0x54, 0x47, 0x4f, 0x49, 0x4e, 0x47, 0x10, 0x02, 0x2a, 0xfa, 0x05,
	0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x55, 0x43, 0x43,
	0x45, 0x53, 0x53, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x43, 0x41, 0x4e, 0x43, 0x45, 0x4c, 0x45,
	0x44, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x42, 0x41, 0x44, 0x5f, 0x50, 0x41, 0x54, 0x48, 0x10,
	0x02, 0x12, 0x0c, 0x0a, 0x08, 0x42, 0x41, 0x44, 0x5f, 0x46, 0x49, 0x4c, 0x45, 0x10, 0x03, 0x12,
	0x0d, 0x0a, 0x09, 0x54, 0x52, 0x41, 0x4e, 0x53, 0x50, 0x4f, 0x52, 0x54, 0x10, 0x04, 0x12, 0x0e,
	0x0a, 0x0a, 0x42, 0x41, 0x44, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x10, 0x05, 0x12, 0x10,
	0x0a, 0x0c, 0x53, 0x45, 0x52, 0x56, 0x49, 0x43, 0x45, 0x5f, 0x53, 0x54, 0x4f, 0x50, 0x10, 0x06,
	0x12, 0x10, 0x0a, 0x0c, 0x42, 0x41, 0x44, 0x5f, 0x54, 0x52, 0x41, 0x4e, 0x53, 0x46, 0x45, 0x52,
	0x10, 0x07, 0x12, 0x16, 0x0a, 0x12, 0x42, 0x41, 0x44, 0x5f, 0x54, 0x52, 0x41, 0x4e, 0x53, 0x46,
	0x45, 0x52, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x10, 0x08, 0x12, 0x0f, 0x0a, 0x0b, 0x42, 0x41,
	0x44, 0x5f, 0x46, 0x49, 0x4c, 0x45, 0x5f, 0x49, 0x44, 0x10, 0x09, 0x12, 0x13, 0x0a, 0x0f, 0x42,
	0x41, 0x44, 0x5f, 0x53, 0x59, 0x53, 0x54, 0x45, 0x4d, 0x5f, 0x54, 0x49, 0x4d, 0x45, 0x10, 0x0a,
	0x12, 0x12, 0x0a, 0x0e, 0x54, 0x52, 0x55, 0x4e, 0x43, 0x41, 0x54, 0x45, 0x44, 0x5f, 0x46, 0x49,
	0x4c, 0x45, 0x10, 0x0b, 0x12, 0x0e, 0x0a, 0x0a, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x5f, 0x53, 0x45,
	0x4e, 0x44, 0x10, 0x0c, 0x12, 0x0c, 0x0a, 0x08, 0x42, 0x41, 0x44, 0x5f, 0x55, 0x55, 0x49, 0x44,
	0x10, 0x0d, 0x12, 0x12, 0x0a, 0x0e, 0x43, 0x48, 0x41, 0x4e, 0x4e, 0x45, 0x4c, 0x5f, 0x43, 0x4c,
	0x4f, 0x53, 0x45, 0x44, 0x10, 0x0e, 0x12, 0x06, 0x0a, 0x02, 0x49, 0x4f, 0x10, 0x0f, 0x12, 0x0d,
	0x0a, 0x09, 0x44, 0x41, 0x54, 0x41, 0x5f, 0x53, 0x45, 0x4e, 0x44, 0x10, 0x10, 0x12, 0x1a, 0x0a,
	0x16, 0x44, 0x49, 0x52, 0x45, 0x43, 0x54, 0x4f, 0x52, 0x59, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x45,
	0x58, 0x50, 0x45, 0x43, 0x54, 0x45, 0x44, 0x10, 0x11, 0x12, 0x12, 0x0a, 0x0e, 0x45, 0x4d, 0x50,
	0x54, 0x59, 0x5f, 0x54, 0x52, 0x41, 0x4e, 0x53, 0x46, 0x45, 0x52, 0x10, 0x12, 0x12, 0x1b, 0x0a,
	0x17, 0x54, 0x52, 0x41, 0x4e, 0x53, 0x46, 0x45, 0x52, 0x5f, 0x43, 0x4c, 0x4f, 0x53, 0x45, 0x44,
	0x5f, 0x42, 0x59, 0x5f, 0x50, 0x45, 0x45, 0x52, 0x10, 0x13, 0x12, 0x1c, 0x0a, 0x18, 0x54, 0x52,
	0x41, 0x4e, 0x53, 0x46, 0x45, 0x52, 0x5f, 0x4c, 0x49, 0x4d, 0x49, 0x54, 0x53, 0x5f, 0x45, 0x58,
	0x43, 0x45, 0x45, 0x44, 0x45, 0x44, 0x10, 0x14, 0x12, 0x13, 0x0a, 0x0f, 0x4d, 0x49, 0x53, 0x4d,
	0x41, 0x54, 0x43, 0x48, 0x45, 0x44, 0x5f, 0x53, 0x49, 0x5a, 0x45, 0x10, 0x15, 0x12, 0x13, 0x0a,
	0x0f, 0x55, 0x4e, 0x45, 0x58, 0x50, 0x45, 0x43, 0x54, 0x45, 0x44, 0x5f, 0x44, 0x41, 0x54, 0x41,
	0x10, 0x16, 0x12, 0x14, 0x0a, 0x10, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x41, 0x52,
	0x47, 0x55, 0x4d, 0x45, 0x4e, 0x54, 0x10, 0x17, 0x12, 0x14, 0x0a, 0x10, 0x54, 0x52, 0x41, 0x4e,
	0x53, 0x46, 0x45, 0x52, 0x5f, 0x54, 0x49, 0x4d, 0x45, 0x4f, 0x55, 0x54, 0x10, 0x18, 0x12, 0x0d,
	0x0a, 0x09, 0x57, 0x53, 0x5f, 0x53, 0x45, 0x52, 0x56, 0x45, 0x52, 0x10, 0x19, 0x12, 0x0d, 0x0a,
	0x09, 0x57, 0x53, 0x5f, 0x43, 0x4c, 0x49, 0x45, 0x4e, 0x54, 0x10, 0x1a, 0x12, 0x11, 0x0a, 0x0d,
	0x46, 0x49, 0x4c, 0x45, 0x5f, 0x4d, 0x4f, 0x44, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x1c, 0x12,
	0x15, 0x0a, 0x11, 0x46, 0x49, 0x4c, 0x45, 0x4e, 0x41, 0x4d, 0x45, 0x5f, 0x54, 0x4f, 0x4f, 0x5f,
	0x4c, 0x4f, 0x4e, 0x47, 0x10, 0x1d, 0x12, 0x19, 0x0a, 0x15, 0x41, 0x55, 0x54, 0x48, 0x45, 0x4e,
	0x54, 0x49, 0x43, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10,
	0x1e, 0x12, 0x1a, 0x0a, 0x16, 0x46, 0x49, 0x4c, 0x45, 0x5f, 0x43, 0x48, 0x45, 0x43, 0x4b, 0x53,
	0x55, 0x4d, 0x5f, 0x4d, 0x49, 0x53, 0x4d, 0x41, 0x54, 0x43, 0x48, 0x10, 0x21, 0x12, 0x11, 0x0a,
	0x0d, 0x46, 0x49, 0x4c, 0x45, 0x5f, 0x52, 0x45, 0x4a, 0x45, 0x43, 0x54, 0x45, 0x44, 0x10, 0x22,
	0x12, 0x0d, 0x0a, 0x09, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x45, 0x44, 0x10, 0x64, 0x12,
	0x0b, 0x0a, 0x07, 0x4f, 0x4e, 0x47, 0x4f, 0x49, 0x4e, 0x47, 0x10, 0x65, 0x12, 0x18, 0x0a, 0x14,
	0x46, 0x49, 0x4e, 0x49, 0x53, 0x48, 0x45, 0x44, 0x5f, 0x57, 0x49, 0x54, 0x48, 0x5f, 0x45, 0x52,
	0x52, 0x4f, 0x52, 0x53, 0x10, 0x66, 0x12, 0x12, 0x0a, 0x0e, 0x41, 0x43, 0x43, 0x45, 0x50, 0x54,
	0x5f, 0x46, 0x41, 0x49, 0x4c, 0x55, 0x52, 0x45, 0x10, 0x67, 0x12, 0x14, 0x0a, 0x10, 0x43, 0x41,
	0x4e, 0x43, 0x45, 0x4c, 0x45, 0x44, 0x5f, 0x42, 0x59, 0x5f, 0x50, 0x45, 0x45, 0x52, 0x10, 0x68,
	0x12, 0x0f, 0x0a, 0x0b, 0x49, 0x4e, 0x54, 0x45, 0x52, 0x52, 0x55, 0x50, 0x54, 0x45, 0x44, 0x10,
	0x69, 0x12, 0x0a, 0x0a, 0x06, 0x50, 0x41, 0x55, 0x53, 0x45, 0x44, 0x10, 0x6a, 0x12, 0x0b, 0x0a,
	0x07, 0x50, 0x45, 0x4e, 0x44, 0x49, 0x4e, 0x47, 0x10, 0x6b, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4e, 0x6f, 0x72, 0x64, 0x53, 0x65, 0x63,
	0x75, 0x72, 0x69, 0x74, 0x79, 0x2f, 0x6e, 0x6f, 0x72, 0x64, 0x76, 0x70, 0x6e, 0x2d, 0x6c, 0x69,
	0x6e, 0x75, 0x78, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x68, 0x61, 0x72, 0x65, 0x2f, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protobuf_fileshare_transfer_proto_rawDescOnce sync.Once
	file_protobuf_fileshare_transfer_proto_rawDescData = file_protobuf_fileshare_transfer_proto_rawDesc
)

func file_protobuf_fileshare_transfer_proto_rawDescGZIP() []byte {
	file_protobuf_fileshare_transfer_proto_rawDescOnce.Do(func() {
		file_protobuf_fileshare_transfer_proto_rawDescData = protoimpl.X.CompressGZIP(file_protobuf_fileshare_transfer_proto_rawDescData)
	})
	return file_protobuf_fileshare_transfer_proto_rawDescData
}

var file_protobuf_fileshare_transfer_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_protobuf_fileshare_transfer_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_protobuf_fileshare_transfer_proto_goTypes = []any{
	(Direction)(0),                // 0: filesharepb.Direction
	(Status)(0),                   // 1: filesharepb.Status
	(*Transfer)(nil),              // 2: filesharepb.Transfer
	(*File)(nil),                  // 3: filesharepb.File
	nil,                           // 4: filesharepb.File.ChildrenEntry
	(*timestamppb.Timestamp)(nil), // 5: google.protobuf.Timestamp
}
var file_protobuf_fileshare_transfer_proto_depIdxs = []int32{
	0, // 0: filesharepb.Transfer.direction:type_name -> filesharepb.Direction
	1, // 1: filesharepb.Transfer.status:type_name -> filesharepb.Status
	5, // 2: filesharepb.Transfer.created:type_name -> google.protobuf.Timestamp
	3, // 3: filesharepb.Transfer.files:type_name -> filesharepb.File
	1, // 4: filesharepb.File.status:type_name -> filesharepb.Status
	4, // 5: filesharepb.File.children:type_name -> filesharepb.File.ChildrenEntry
	3, // 6: filesharepb.File.ChildrenEntry.value:type_name -> filesharepb.File
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_protobuf_fileshare_transfer_proto_init() }
func file_protobuf_fileshare_transfer_proto_init() {
	if File_protobuf_fileshare_transfer_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protobuf_fileshare_transfer_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protobuf_fileshare_transfer_proto_goTypes,
		DependencyIndexes: file_protobuf_fileshare_transfer_proto_depIdxs,
		EnumInfos:         file_protobuf_fileshare_transfer_proto_enumTypes,
		MessageInfos:      file_protobuf_fileshare_transfer_proto_msgTypes,
	}.Build()
	File_protobuf_fileshare_transfer_proto = out.File
	file_protobuf_fileshare_transfer_proto_rawDesc = nil
	file_protobuf_fileshare_transfer_proto_goTypes = nil
	file_protobuf_fileshare_transfer_proto_depIdxs = nil
}
