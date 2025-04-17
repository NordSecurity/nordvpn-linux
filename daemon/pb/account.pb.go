// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v3.21.6
// source: account.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type DedidcatedIPService struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServerIds            []int64 `protobuf:"varint,1,rep,packed,name=server_ids,json=serverIds,proto3" json:"server_ids,omitempty"`
	DedicatedIpExpiresAt string  `protobuf:"bytes,2,opt,name=dedicated_ip_expires_at,json=dedicatedIpExpiresAt,proto3" json:"dedicated_ip_expires_at,omitempty"`
}

func (x *DedidcatedIPService) Reset() {
	*x = DedidcatedIPService{}
	mi := &file_account_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DedidcatedIPService) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DedidcatedIPService) ProtoMessage() {}

func (x *DedidcatedIPService) ProtoReflect() protoreflect.Message {
	mi := &file_account_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DedidcatedIPService.ProtoReflect.Descriptor instead.
func (*DedidcatedIPService) Descriptor() ([]byte, []int) {
	return file_account_proto_rawDescGZIP(), []int{0}
}

func (x *DedidcatedIPService) GetServerIds() []int64 {
	if x != nil {
		return x.ServerIds
	}
	return nil
}

func (x *DedidcatedIPService) GetDedicatedIpExpiresAt() string {
	if x != nil {
		return x.DedicatedIpExpiresAt
	}
	return ""
}

type AccountResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type                     int64                  `protobuf:"varint,1,opt,name=type,proto3" json:"type,omitempty"`
	Username                 string                 `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	Email                    string                 `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
	ExpiresAt                string                 `protobuf:"bytes,4,opt,name=expires_at,json=expiresAt,proto3" json:"expires_at,omitempty"`
	DedicatedIpStatus        int64                  `protobuf:"varint,5,opt,name=dedicated_ip_status,json=dedicatedIpStatus,proto3" json:"dedicated_ip_status,omitempty"`
	LastDedicatedIpExpiresAt string                 `protobuf:"bytes,6,opt,name=last_dedicated_ip_expires_at,json=lastDedicatedIpExpiresAt,proto3" json:"last_dedicated_ip_expires_at,omitempty"`
	DedicatedIpServices      []*DedidcatedIPService `protobuf:"bytes,7,rep,name=dedicated_ip_services,json=dedicatedIpServices,proto3" json:"dedicated_ip_services,omitempty"`
	MfaStatus                TriState               `protobuf:"varint,8,opt,name=mfa_status,json=mfaStatus,proto3,enum=pb.TriState" json:"mfa_status,omitempty"`
}

func (x *AccountResponse) Reset() {
	*x = AccountResponse{}
	mi := &file_account_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AccountResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AccountResponse) ProtoMessage() {}

func (x *AccountResponse) ProtoReflect() protoreflect.Message {
	mi := &file_account_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AccountResponse.ProtoReflect.Descriptor instead.
func (*AccountResponse) Descriptor() ([]byte, []int) {
	return file_account_proto_rawDescGZIP(), []int{1}
}

func (x *AccountResponse) GetType() int64 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *AccountResponse) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *AccountResponse) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *AccountResponse) GetExpiresAt() string {
	if x != nil {
		return x.ExpiresAt
	}
	return ""
}

func (x *AccountResponse) GetDedicatedIpStatus() int64 {
	if x != nil {
		return x.DedicatedIpStatus
	}
	return 0
}

func (x *AccountResponse) GetLastDedicatedIpExpiresAt() string {
	if x != nil {
		return x.LastDedicatedIpExpiresAt
	}
	return ""
}

func (x *AccountResponse) GetDedicatedIpServices() []*DedidcatedIPService {
	if x != nil {
		return x.DedicatedIpServices
	}
	return nil
}

func (x *AccountResponse) GetMfaStatus() TriState {
	if x != nil {
		return x.MfaStatus
	}
	return TriState_UNKNOWN
}

type AccountRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Full bool `protobuf:"varint,1,opt,name=full,proto3" json:"full,omitempty"`
}

func (x *AccountRequest) Reset() {
	*x = AccountRequest{}
	mi := &file_account_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AccountRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AccountRequest) ProtoMessage() {}

func (x *AccountRequest) ProtoReflect() protoreflect.Message {
	mi := &file_account_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AccountRequest.ProtoReflect.Descriptor instead.
func (*AccountRequest) Descriptor() ([]byte, []int) {
	return file_account_proto_rawDescGZIP(), []int{2}
}

func (x *AccountRequest) GetFull() bool {
	if x != nil {
		return x.Full
	}
	return false
}

var File_account_proto protoreflect.FileDescriptor

var file_account_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x02, 0x70, 0x62, 0x1a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x6b, 0x0a, 0x13, 0x44, 0x65, 0x64, 0x69, 0x64, 0x63, 0x61, 0x74, 0x65, 0x64, 0x49,
	0x50, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x03, 0x52, 0x09, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x49, 0x64, 0x73, 0x12, 0x35, 0x0a, 0x17, 0x64, 0x65, 0x64, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x64, 0x5f, 0x69, 0x70, 0x5f, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0x5f,
	0x61, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x14, 0x64, 0x65, 0x64, 0x69, 0x63, 0x61,
	0x74, 0x65, 0x64, 0x49, 0x70, 0x45, 0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0x41, 0x74, 0x22, 0xe0,
	0x02, 0x0a, 0x0f, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x12, 0x1d, 0x0a, 0x0a, 0x65, 0x78, 0x70, 0x69,
	0x72, 0x65, 0x73, 0x5f, 0x61, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x65, 0x78,
	0x70, 0x69, 0x72, 0x65, 0x73, 0x41, 0x74, 0x12, 0x2e, 0x0a, 0x13, 0x64, 0x65, 0x64, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x64, 0x5f, 0x69, 0x70, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x11, 0x64, 0x65, 0x64, 0x69, 0x63, 0x61, 0x74, 0x65, 0x64, 0x49,
	0x70, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x3e, 0x0a, 0x1c, 0x6c, 0x61, 0x73, 0x74, 0x5f,
	0x64, 0x65, 0x64, 0x69, 0x63, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x69, 0x70, 0x5f, 0x65, 0x78, 0x70,
	0x69, 0x72, 0x65, 0x73, 0x5f, 0x61, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x18, 0x6c,
	0x61, 0x73, 0x74, 0x44, 0x65, 0x64, 0x69, 0x63, 0x61, 0x74, 0x65, 0x64, 0x49, 0x70, 0x45, 0x78,
	0x70, 0x69, 0x72, 0x65, 0x73, 0x41, 0x74, 0x12, 0x4b, 0x0a, 0x15, 0x64, 0x65, 0x64, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x64, 0x5f, 0x69, 0x70, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73,
	0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x70, 0x62, 0x2e, 0x44, 0x65, 0x64, 0x69,
	0x64, 0x63, 0x61, 0x74, 0x65, 0x64, 0x49, 0x50, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52,
	0x13, 0x64, 0x65, 0x64, 0x69, 0x63, 0x61, 0x74, 0x65, 0x64, 0x49, 0x70, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x73, 0x12, 0x2b, 0x0a, 0x0a, 0x6d, 0x66, 0x61, 0x5f, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e, 0x70, 0x62, 0x2e, 0x54, 0x72,
	0x69, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x09, 0x6d, 0x66, 0x61, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x22, 0x24, 0x0a, 0x0e, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x75, 0x6c, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x04, 0x66, 0x75, 0x6c, 0x6c, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4e, 0x6f, 0x72, 0x64, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69,
	0x74, 0x79, 0x2f, 0x6e, 0x6f, 0x72, 0x64, 0x76, 0x70, 0x6e, 0x2d, 0x6c, 0x69, 0x6e, 0x75, 0x78,
	0x2f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_account_proto_rawDescOnce sync.Once
	file_account_proto_rawDescData = file_account_proto_rawDesc
)

func file_account_proto_rawDescGZIP() []byte {
	file_account_proto_rawDescOnce.Do(func() {
		file_account_proto_rawDescData = protoimpl.X.CompressGZIP(file_account_proto_rawDescData)
	})
	return file_account_proto_rawDescData
}

var file_account_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_account_proto_goTypes = []any{
	(*DedidcatedIPService)(nil), // 0: pb.DedidcatedIPService
	(*AccountResponse)(nil),     // 1: pb.AccountResponse
	(*AccountRequest)(nil),      // 2: pb.AccountRequest
	(TriState)(0),               // 3: pb.TriState
}
var file_account_proto_depIdxs = []int32{
	0, // 0: pb.AccountResponse.dedicated_ip_services:type_name -> pb.DedidcatedIPService
	3, // 1: pb.AccountResponse.mfa_status:type_name -> pb.TriState
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_account_proto_init() }
func file_account_proto_init() {
	if File_account_proto != nil {
		return
	}
	file_common_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_account_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_account_proto_goTypes,
		DependencyIndexes: file_account_proto_depIdxs,
		MessageInfos:      file_account_proto_msgTypes,
	}.Build()
	File_account_proto = out.File
	file_account_proto_rawDesc = nil
	file_account_proto_goTypes = nil
	file_account_proto_depIdxs = nil
}
