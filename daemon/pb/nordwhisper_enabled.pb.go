// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v3.21.6
// source: protobuf/daemon/nordwhisper_enabled.proto

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

type NordWhisperEnabled struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Enabled bool `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty"`
}

func (x *NordWhisperEnabled) Reset() {
	*x = NordWhisperEnabled{}
	mi := &file_protobuf_daemon_nordwhisper_enabled_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *NordWhisperEnabled) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NordWhisperEnabled) ProtoMessage() {}

func (x *NordWhisperEnabled) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_daemon_nordwhisper_enabled_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NordWhisperEnabled.ProtoReflect.Descriptor instead.
func (*NordWhisperEnabled) Descriptor() ([]byte, []int) {
	return file_protobuf_daemon_nordwhisper_enabled_proto_rawDescGZIP(), []int{0}
}

func (x *NordWhisperEnabled) GetEnabled() bool {
	if x != nil {
		return x.Enabled
	}
	return false
}

var File_protobuf_daemon_nordwhisper_enabled_proto protoreflect.FileDescriptor

var file_protobuf_daemon_nordwhisper_enabled_proto_rawDesc = []byte{
	0x0a, 0x29, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x61, 0x65, 0x6d, 0x6f,
	0x6e, 0x2f, 0x6e, 0x6f, 0x72, 0x64, 0x77, 0x68, 0x69, 0x73, 0x70, 0x65, 0x72, 0x5f, 0x65, 0x6e,
	0x61, 0x62, 0x6c, 0x65, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x22,
	0x2e, 0x0a, 0x12, 0x4e, 0x6f, 0x72, 0x64, 0x57, 0x68, 0x69, 0x73, 0x70, 0x65, 0x72, 0x45, 0x6e,
	0x61, 0x62, 0x6c, 0x65, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x42,
	0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4e, 0x6f,
	0x72, 0x64, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x2f, 0x6e, 0x6f, 0x72, 0x64, 0x76,
	0x70, 0x6e, 0x2d, 0x6c, 0x69, 0x6e, 0x75, 0x78, 0x2f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2f,
	0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protobuf_daemon_nordwhisper_enabled_proto_rawDescOnce sync.Once
	file_protobuf_daemon_nordwhisper_enabled_proto_rawDescData = file_protobuf_daemon_nordwhisper_enabled_proto_rawDesc
)

func file_protobuf_daemon_nordwhisper_enabled_proto_rawDescGZIP() []byte {
	file_protobuf_daemon_nordwhisper_enabled_proto_rawDescOnce.Do(func() {
		file_protobuf_daemon_nordwhisper_enabled_proto_rawDescData = protoimpl.X.CompressGZIP(file_protobuf_daemon_nordwhisper_enabled_proto_rawDescData)
	})
	return file_protobuf_daemon_nordwhisper_enabled_proto_rawDescData
}

var file_protobuf_daemon_nordwhisper_enabled_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protobuf_daemon_nordwhisper_enabled_proto_goTypes = []any{
	(*NordWhisperEnabled)(nil), // 0: pb.NordWhisperEnabled
}
var file_protobuf_daemon_nordwhisper_enabled_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_protobuf_daemon_nordwhisper_enabled_proto_init() }
func file_protobuf_daemon_nordwhisper_enabled_proto_init() {
	if File_protobuf_daemon_nordwhisper_enabled_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protobuf_daemon_nordwhisper_enabled_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protobuf_daemon_nordwhisper_enabled_proto_goTypes,
		DependencyIndexes: file_protobuf_daemon_nordwhisper_enabled_proto_depIdxs,
		MessageInfos:      file_protobuf_daemon_nordwhisper_enabled_proto_msgTypes,
	}.Build()
	File_protobuf_daemon_nordwhisper_enabled_proto = out.File
	file_protobuf_daemon_nordwhisper_enabled_proto_rawDesc = nil
	file_protobuf_daemon_nordwhisper_enabled_proto_goTypes = nil
	file_protobuf_daemon_nordwhisper_enabled_proto_depIdxs = nil
}
