# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: servers.proto
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
    'servers.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from config import group_pb2 as config_dot_group__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\rservers.proto\x12\x02pb\x1a\x12\x63onfig/group.proto\"\x8a\x01\n\x06Server\x12\n\n\x02id\x18\x01 \x01(\x03\x12\x11\n\thost_name\x18\x04 \x01(\t\x12\x0f\n\x07virtual\x18\x05 \x01(\x08\x12*\n\rserver_groups\x18\x06 \x03(\x0e\x32\x13.config.ServerGroup\x12$\n\x0ctechnologies\x18\x07 \x03(\x0e\x32\x0e.pb.Technology\"<\n\nServerCity\x12\x11\n\tcity_name\x18\x01 \x01(\t\x12\x1b\n\x07servers\x18\x02 \x03(\x0b\x32\n.pb.Server\"[\n\rServerCountry\x12\x14\n\x0c\x63ountry_code\x18\x01 \x01(\t\x12\x1e\n\x06\x63ities\x18\x02 \x03(\x0b\x32\x0e.pb.ServerCity\x12\x14\n\x0c\x63ountry_name\x18\x03 \x01(\t\";\n\nServersMap\x12-\n\x12servers_by_country\x18\x01 \x03(\x0b\x32\x11.pb.ServerCountry\"c\n\x0fServersResponse\x12!\n\x07servers\x18\x01 \x01(\x0b\x32\x0e.pb.ServersMapH\x00\x12!\n\x05\x65rror\x18\x02 \x01(\x0e\x32\x10.pb.ServersErrorH\x00\x42\n\n\x08response*L\n\x0cServersError\x12\x0c\n\x08NO_ERROR\x10\x00\x12\x14\n\x10GET_CONFIG_ERROR\x10\x01\x12\x18\n\x14\x46ILTER_SERVERS_ERROR\x10\x02*\x8b\x01\n\nTechnology\x12\x15\n\x11UNKNOWN_TECHNLOGY\x10\x00\x12\x0c\n\x08NORDLYNX\x10\x01\x12\x0f\n\x0bOPENVPN_TCP\x10\x02\x12\x0f\n\x0bOPENVPN_UDP\x10\x03\x12\x1a\n\x16OBFUSCATED_OPENVPN_TCP\x10\x04\x12\x1a\n\x16OBFUSCATED_OPENVPN_UDP\x10\x05\x42\x31Z/github.com/NordSecurity/nordvpn-linux/daemon/pbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'servers_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z/github.com/NordSecurity/nordvpn-linux/daemon/pb'
  _globals['_SERVERSERROR']._serialized_start=499
  _globals['_SERVERSERROR']._serialized_end=575
  _globals['_TECHNOLOGY']._serialized_start=578
  _globals['_TECHNOLOGY']._serialized_end=717
  _globals['_SERVER']._serialized_start=42
  _globals['_SERVER']._serialized_end=180
  _globals['_SERVERCITY']._serialized_start=182
  _globals['_SERVERCITY']._serialized_end=242
  _globals['_SERVERCOUNTRY']._serialized_start=244
  _globals['_SERVERCOUNTRY']._serialized_end=335
  _globals['_SERVERSMAP']._serialized_start=337
  _globals['_SERVERSMAP']._serialized_end=396
  _globals['_SERVERSRESPONSE']._serialized_start=398
  _globals['_SERVERSRESPONSE']._serialized_end=497
# @@protoc_insertion_point(module_scope)
