// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.6
// source: servers.proto

package pb

import (
	config "github.com/NordSecurity/nordvpn-linux/config"
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

type ServersError int32

const (
	ServersError_NO_ERROR             ServersError = 0
	ServersError_GET_CONFIG_ERROR     ServersError = 1
	ServersError_FILTER_SERVERS_ERROR ServersError = 2
)

// Enum value maps for ServersError.
var (
	ServersError_name = map[int32]string{
		0: "NO_ERROR",
		1: "GET_CONFIG_ERROR",
		2: "FILTER_SERVERS_ERROR",
	}
	ServersError_value = map[string]int32{
		"NO_ERROR":             0,
		"GET_CONFIG_ERROR":     1,
		"FILTER_SERVERS_ERROR": 2,
	}
)

func (x ServersError) Enum() *ServersError {
	p := new(ServersError)
	*p = x
	return p
}

func (x ServersError) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ServersError) Descriptor() protoreflect.EnumDescriptor {
	return file_servers_proto_enumTypes[0].Descriptor()
}

func (ServersError) Type() protoreflect.EnumType {
	return &file_servers_proto_enumTypes[0]
}

func (x ServersError) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ServersError.Descriptor instead.
func (ServersError) EnumDescriptor() ([]byte, []int) {
	return file_servers_proto_rawDescGZIP(), []int{0}
}

type Technology int32

const (
	Technology_UNKNOWN_TECHNLOGY      Technology = 0
	Technology_NORDLYNX               Technology = 1
	Technology_OPENVPN_TCP            Technology = 2
	Technology_OPENVPN_UDP            Technology = 3
	Technology_OBFUSCATED_OPENVPN_TCP Technology = 4
	Technology_OBFUSCATED_OPENVPN_UDP Technology = 5
)

// Enum value maps for Technology.
var (
	Technology_name = map[int32]string{
		0: "UNKNOWN_TECHNLOGY",
		1: "NORDLYNX",
		2: "OPENVPN_TCP",
		3: "OPENVPN_UDP",
		4: "OBFUSCATED_OPENVPN_TCP",
		5: "OBFUSCATED_OPENVPN_UDP",
	}
	Technology_value = map[string]int32{
		"UNKNOWN_TECHNLOGY":      0,
		"NORDLYNX":               1,
		"OPENVPN_TCP":            2,
		"OPENVPN_UDP":            3,
		"OBFUSCATED_OPENVPN_TCP": 4,
		"OBFUSCATED_OPENVPN_UDP": 5,
	}
)

func (x Technology) Enum() *Technology {
	p := new(Technology)
	*p = x
	return p
}

func (x Technology) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Technology) Descriptor() protoreflect.EnumDescriptor {
	return file_servers_proto_enumTypes[1].Descriptor()
}

func (Technology) Type() protoreflect.EnumType {
	return &file_servers_proto_enumTypes[1]
}

func (x Technology) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Technology.Descriptor instead.
func (Technology) EnumDescriptor() ([]byte, []int) {
	return file_servers_proto_rawDescGZIP(), []int{1}
}

type Server struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id           int64                `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	HostName     string               `protobuf:"bytes,4,opt,name=host_name,json=hostName,proto3" json:"host_name,omitempty"`
	Virtual      bool                 `protobuf:"varint,5,opt,name=virtual,proto3" json:"virtual,omitempty"`
	ServerGroups []config.ServerGroup `protobuf:"varint,6,rep,packed,name=server_groups,json=serverGroups,proto3,enum=config.ServerGroup" json:"server_groups,omitempty"`
	Technologies []Technology         `protobuf:"varint,7,rep,packed,name=technologies,proto3,enum=pb.Technology" json:"technologies,omitempty"`
}

func (x *Server) Reset() {
	*x = Server{}
	if protoimpl.UnsafeEnabled {
		mi := &file_servers_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Server) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Server) ProtoMessage() {}

func (x *Server) ProtoReflect() protoreflect.Message {
	mi := &file_servers_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Server.ProtoReflect.Descriptor instead.
func (*Server) Descriptor() ([]byte, []int) {
	return file_servers_proto_rawDescGZIP(), []int{0}
}

func (x *Server) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Server) GetHostName() string {
	if x != nil {
		return x.HostName
	}
	return ""
}

func (x *Server) GetVirtual() bool {
	if x != nil {
		return x.Virtual
	}
	return false
}

func (x *Server) GetServerGroups() []config.ServerGroup {
	if x != nil {
		return x.ServerGroups
	}
	return nil
}

func (x *Server) GetTechnologies() []Technology {
	if x != nil {
		return x.Technologies
	}
	return nil
}

type ServerCity struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CityName string    `protobuf:"bytes,1,opt,name=city_name,json=cityName,proto3" json:"city_name,omitempty"`
	Servers  []*Server `protobuf:"bytes,2,rep,name=servers,proto3" json:"servers,omitempty"`
}

func (x *ServerCity) Reset() {
	*x = ServerCity{}
	if protoimpl.UnsafeEnabled {
		mi := &file_servers_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerCity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerCity) ProtoMessage() {}

func (x *ServerCity) ProtoReflect() protoreflect.Message {
	mi := &file_servers_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerCity.ProtoReflect.Descriptor instead.
func (*ServerCity) Descriptor() ([]byte, []int) {
	return file_servers_proto_rawDescGZIP(), []int{1}
}

func (x *ServerCity) GetCityName() string {
	if x != nil {
		return x.CityName
	}
	return ""
}

func (x *ServerCity) GetServers() []*Server {
	if x != nil {
		return x.Servers
	}
	return nil
}

type ServerCountry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CountryCode string        `protobuf:"bytes,1,opt,name=country_code,json=countryCode,proto3" json:"country_code,omitempty"`
	Cities      []*ServerCity `protobuf:"bytes,2,rep,name=cities,proto3" json:"cities,omitempty"`
}

func (x *ServerCountry) Reset() {
	*x = ServerCountry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_servers_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerCountry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerCountry) ProtoMessage() {}

func (x *ServerCountry) ProtoReflect() protoreflect.Message {
	mi := &file_servers_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerCountry.ProtoReflect.Descriptor instead.
func (*ServerCountry) Descriptor() ([]byte, []int) {
	return file_servers_proto_rawDescGZIP(), []int{2}
}

func (x *ServerCountry) GetCountryCode() string {
	if x != nil {
		return x.CountryCode
	}
	return ""
}

func (x *ServerCountry) GetCities() []*ServerCity {
	if x != nil {
		return x.Cities
	}
	return nil
}

type ServersMap struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServersByCountry []*ServerCountry `protobuf:"bytes,1,rep,name=servers_by_country,json=serversByCountry,proto3" json:"servers_by_country,omitempty"`
}

func (x *ServersMap) Reset() {
	*x = ServersMap{}
	if protoimpl.UnsafeEnabled {
		mi := &file_servers_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServersMap) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServersMap) ProtoMessage() {}

func (x *ServersMap) ProtoReflect() protoreflect.Message {
	mi := &file_servers_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServersMap.ProtoReflect.Descriptor instead.
func (*ServersMap) Descriptor() ([]byte, []int) {
	return file_servers_proto_rawDescGZIP(), []int{3}
}

func (x *ServersMap) GetServersByCountry() []*ServerCountry {
	if x != nil {
		return x.ServersByCountry
	}
	return nil
}

type ServersResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Response:
	//
	//	*ServersResponse_Servers
	//	*ServersResponse_Error
	Response isServersResponse_Response `protobuf_oneof:"response"`
}

func (x *ServersResponse) Reset() {
	*x = ServersResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_servers_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServersResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServersResponse) ProtoMessage() {}

func (x *ServersResponse) ProtoReflect() protoreflect.Message {
	mi := &file_servers_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServersResponse.ProtoReflect.Descriptor instead.
func (*ServersResponse) Descriptor() ([]byte, []int) {
	return file_servers_proto_rawDescGZIP(), []int{4}
}

func (m *ServersResponse) GetResponse() isServersResponse_Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (x *ServersResponse) GetServers() *ServersMap {
	if x, ok := x.GetResponse().(*ServersResponse_Servers); ok {
		return x.Servers
	}
	return nil
}

func (x *ServersResponse) GetError() ServersError {
	if x, ok := x.GetResponse().(*ServersResponse_Error); ok {
		return x.Error
	}
	return ServersError_NO_ERROR
}

type isServersResponse_Response interface {
	isServersResponse_Response()
}

type ServersResponse_Servers struct {
	Servers *ServersMap `protobuf:"bytes,1,opt,name=servers,proto3,oneof"`
}

type ServersResponse_Error struct {
	Error ServersError `protobuf:"varint,2,opt,name=error,proto3,enum=pb.ServersError,oneof"`
}

func (*ServersResponse_Servers) isServersResponse_Response() {}

func (*ServersResponse_Error) isServersResponse_Response() {}

var File_servers_proto protoreflect.FileDescriptor

var file_servers_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x02, 0x70, 0x62, 0x1a, 0x12, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x67, 0x72, 0x6f, 0x75,
	0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbd, 0x01, 0x0a, 0x06, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x68, 0x6f, 0x73, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12,
	0x18, 0x0a, 0x07, 0x76, 0x69, 0x72, 0x74, 0x75, 0x61, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x07, 0x76, 0x69, 0x72, 0x74, 0x75, 0x61, 0x6c, 0x12, 0x38, 0x0a, 0x0d, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x5f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0e,
	0x32, 0x13, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x47, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x0c, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x73, 0x12, 0x32, 0x0a, 0x0c, 0x74, 0x65, 0x63, 0x68, 0x6e, 0x6f, 0x6c, 0x6f, 0x67,
	0x69, 0x65, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x54,
	0x65, 0x63, 0x68, 0x6e, 0x6f, 0x6c, 0x6f, 0x67, 0x79, 0x52, 0x0c, 0x74, 0x65, 0x63, 0x68, 0x6e,
	0x6f, 0x6c, 0x6f, 0x67, 0x69, 0x65, 0x73, 0x22, 0x4f, 0x0a, 0x0a, 0x53, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x43, 0x69, 0x74, 0x79, 0x12, 0x1b, 0x0a, 0x09, 0x63, 0x69, 0x74, 0x79, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x69, 0x74, 0x79, 0x4e, 0x61,
	0x6d, 0x65, 0x12, 0x24, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x52,
	0x07, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x22, 0x5a, 0x0a, 0x0d, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x72, 0x79, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0b, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x26, 0x0a, 0x06,
	0x63, 0x69, 0x74, 0x69, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70,
	0x62, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x43, 0x69, 0x74, 0x79, 0x52, 0x06, 0x63, 0x69,
	0x74, 0x69, 0x65, 0x73, 0x22, 0x4d, 0x0a, 0x0a, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x4d,
	0x61, 0x70, 0x12, 0x3f, 0x0a, 0x12, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x5f, 0x62, 0x79,
	0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11,
	0x2e, 0x70, 0x62, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x10, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x42, 0x79, 0x43, 0x6f, 0x75, 0x6e,
	0x74, 0x72, 0x79, 0x22, 0x73, 0x0a, 0x0f, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x73, 0x4d, 0x61, 0x70, 0x48, 0x00, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x73, 0x12, 0x28, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x10, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x45, 0x72,
	0x72, 0x6f, 0x72, 0x48, 0x00, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x42, 0x0a, 0x0a, 0x08,
	0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2a, 0x4c, 0x0a, 0x0c, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x73, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x0c, 0x0a, 0x08, 0x4e, 0x4f, 0x5f, 0x45,
	0x52, 0x52, 0x4f, 0x52, 0x10, 0x00, 0x12, 0x14, 0x0a, 0x10, 0x47, 0x45, 0x54, 0x5f, 0x43, 0x4f,
	0x4e, 0x46, 0x49, 0x47, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x01, 0x12, 0x18, 0x0a, 0x14,
	0x46, 0x49, 0x4c, 0x54, 0x45, 0x52, 0x5f, 0x53, 0x45, 0x52, 0x56, 0x45, 0x52, 0x53, 0x5f, 0x45,
	0x52, 0x52, 0x4f, 0x52, 0x10, 0x02, 0x2a, 0x8b, 0x01, 0x0a, 0x0a, 0x54, 0x65, 0x63, 0x68, 0x6e,
	0x6f, 0x6c, 0x6f, 0x67, 0x79, 0x12, 0x15, 0x0a, 0x11, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e,
	0x5f, 0x54, 0x45, 0x43, 0x48, 0x4e, 0x4c, 0x4f, 0x47, 0x59, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08,
	0x4e, 0x4f, 0x52, 0x44, 0x4c, 0x59, 0x4e, 0x58, 0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b, 0x4f, 0x50,
	0x45, 0x4e, 0x56, 0x50, 0x4e, 0x5f, 0x54, 0x43, 0x50, 0x10, 0x02, 0x12, 0x0f, 0x0a, 0x0b, 0x4f,
	0x50, 0x45, 0x4e, 0x56, 0x50, 0x4e, 0x5f, 0x55, 0x44, 0x50, 0x10, 0x03, 0x12, 0x1a, 0x0a, 0x16,
	0x4f, 0x42, 0x46, 0x55, 0x53, 0x43, 0x41, 0x54, 0x45, 0x44, 0x5f, 0x4f, 0x50, 0x45, 0x4e, 0x56,
	0x50, 0x4e, 0x5f, 0x54, 0x43, 0x50, 0x10, 0x04, 0x12, 0x1a, 0x0a, 0x16, 0x4f, 0x42, 0x46, 0x55,
	0x53, 0x43, 0x41, 0x54, 0x45, 0x44, 0x5f, 0x4f, 0x50, 0x45, 0x4e, 0x56, 0x50, 0x4e, 0x5f, 0x55,
	0x44, 0x50, 0x10, 0x05, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x4e, 0x6f, 0x72, 0x64, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x2f,
	0x6e, 0x6f, 0x72, 0x64, 0x76, 0x70, 0x6e, 0x2d, 0x6c, 0x69, 0x6e, 0x75, 0x78, 0x2f, 0x64, 0x61,
	0x65, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_servers_proto_rawDescOnce sync.Once
	file_servers_proto_rawDescData = file_servers_proto_rawDesc
)

func file_servers_proto_rawDescGZIP() []byte {
	file_servers_proto_rawDescOnce.Do(func() {
		file_servers_proto_rawDescData = protoimpl.X.CompressGZIP(file_servers_proto_rawDescData)
	})
	return file_servers_proto_rawDescData
}

var file_servers_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_servers_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_servers_proto_goTypes = []interface{}{
	(ServersError)(0),       // 0: pb.ServersError
	(Technology)(0),         // 1: pb.Technology
	(*Server)(nil),          // 2: pb.Server
	(*ServerCity)(nil),      // 3: pb.ServerCity
	(*ServerCountry)(nil),   // 4: pb.ServerCountry
	(*ServersMap)(nil),      // 5: pb.ServersMap
	(*ServersResponse)(nil), // 6: pb.ServersResponse
	(config.ServerGroup)(0), // 7: config.ServerGroup
}
var file_servers_proto_depIdxs = []int32{
	7, // 0: pb.Server.server_groups:type_name -> config.ServerGroup
	1, // 1: pb.Server.technologies:type_name -> pb.Technology
	2, // 2: pb.ServerCity.servers:type_name -> pb.Server
	3, // 3: pb.ServerCountry.cities:type_name -> pb.ServerCity
	4, // 4: pb.ServersMap.servers_by_country:type_name -> pb.ServerCountry
	5, // 5: pb.ServersResponse.servers:type_name -> pb.ServersMap
	0, // 6: pb.ServersResponse.error:type_name -> pb.ServersError
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_servers_proto_init() }
func file_servers_proto_init() {
	if File_servers_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_servers_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Server); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_servers_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServerCity); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_servers_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServerCountry); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_servers_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServersMap); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_servers_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServersResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_servers_proto_msgTypes[4].OneofWrappers = []interface{}{
		(*ServersResponse_Servers)(nil),
		(*ServersResponse_Error)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_servers_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_servers_proto_goTypes,
		DependencyIndexes: file_servers_proto_depIdxs,
		EnumInfos:         file_servers_proto_enumTypes,
		MessageInfos:      file_servers_proto_msgTypes,
	}.Build()
	File_servers_proto = out.File
	file_servers_proto_rawDesc = nil
	file_servers_proto_goTypes = nil
	file_servers_proto_depIdxs = nil
}
