# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: state.proto
# Protobuf Python Version: 5.28.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    28,
    1,
    '',
    'state.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


import settings_pb2 as settings__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x0bstate.proto\x12\x02pb\x1a\x0esettings.proto\"\xcb\x01\n\x10\x43onnectionStatus\x12\"\n\x05state\x18\x01 \x01(\x0e\x32\x13.pb.ConnectionState\x12\x11\n\tserver_ip\x18\x02 \x01(\t\x12\x16\n\x0eserver_country\x18\x03 \x01(\t\x12\x13\n\x0bserver_city\x18\x04 \x01(\t\x12\x17\n\x0fserver_hostname\x18\x05 \x01(\t\x12\x13\n\x0bserver_name\x18\x06 \x01(\t\x12\x14\n\x0cis_mesh_peer\x18\x07 \x01(\x08\x12\x0f\n\x07\x62y_user\x18\x08 \x01(\x08\".\n\nLoginEvent\x12 \n\x04type\x18\x01 \x01(\x0e\x32\x12.pb.LoginEventType\"=\n\x13\x41\x63\x63ountModification\x12\x17\n\nexpires_at\x18\x01 \x01(\tH\x00\x88\x01\x01\x42\r\n\x0b_expires_at\"\x9c\x02\n\x08\x41ppState\x12\"\n\x05\x65rror\x18\x01 \x01(\x0e\x32\x11.pb.AppStateErrorH\x00\x12\x31\n\x11\x63onnection_status\x18\x02 \x01(\x0b\x32\x14.pb.ConnectionStatusH\x00\x12%\n\x0blogin_event\x18\x03 \x01(\x0b\x32\x0e.pb.LoginEventH\x00\x12\'\n\x0fsettings_change\x18\x04 \x01(\x0b\x32\x0c.pb.SettingsH\x00\x12\'\n\x0cupdate_event\x18\x05 \x01(\x0e\x32\x0f.pb.UpdateEventH\x00\x12\x37\n\x14\x61\x63\x63ount_modification\x18\x06 \x01(\x0b\x32\x17.pb.AccountModificationH\x00\x42\x07\n\x05state*&\n\rAppStateError\x12\x15\n\x11\x46\x41ILED_TO_GET_UID\x10\x00*B\n\x0f\x43onnectionState\x12\x10\n\x0c\x44ISCONNECTED\x10\x00\x12\x0e\n\nCONNECTING\x10\x01\x12\r\n\tCONNECTED\x10\x02*\'\n\x0eLoginEventType\x12\t\n\x05LOGIN\x10\x00\x12\n\n\x06LOGOUT\x10\x01*&\n\x0bUpdateEvent\x12\x17\n\x13SERVERS_LIST_UPDATE\x10\x00\x42\x31Z/github.com/NordSecurity/nordvpn-linux/daemon/pbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'state_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z/github.com/NordSecurity/nordvpn-linux/daemon/pb'
  _globals['_APPSTATEERROR']._serialized_start=639
  _globals['_APPSTATEERROR']._serialized_end=677
  _globals['_CONNECTIONSTATE']._serialized_start=679
  _globals['_CONNECTIONSTATE']._serialized_end=745
  _globals['_LOGINEVENTTYPE']._serialized_start=747
  _globals['_LOGINEVENTTYPE']._serialized_end=786
  _globals['_UPDATEEVENT']._serialized_start=788
  _globals['_UPDATEEVENT']._serialized_end=826
  _globals['_CONNECTIONSTATUS']._serialized_start=36
  _globals['_CONNECTIONSTATUS']._serialized_end=239
  _globals['_LOGINEVENT']._serialized_start=241
  _globals['_LOGINEVENT']._serialized_end=287
  _globals['_ACCOUNTMODIFICATION']._serialized_start=289
  _globals['_ACCOUNTMODIFICATION']._serialized_end=350
  _globals['_APPSTATE']._serialized_start=353
  _globals['_APPSTATE']._serialized_end=637
# @@protoc_insertion_point(module_scope)