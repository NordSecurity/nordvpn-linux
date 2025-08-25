// This is a generated file - do not edit.
//
// Generated from servers.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use serversErrorDescriptor instead')
const ServersError$json = {
  '1': 'ServersError',
  '2': [
    {'1': 'NO_ERROR', '2': 0},
    {'1': 'GET_CONFIG_ERROR', '2': 1},
    {'1': 'FILTER_SERVERS_ERROR', '2': 2},
  ],
};

/// Descriptor for `ServersError`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List serversErrorDescriptor = $convert.base64Decode(
    'CgxTZXJ2ZXJzRXJyb3ISDAoITk9fRVJST1IQABIUChBHRVRfQ09ORklHX0VSUk9SEAESGAoURk'
    'lMVEVSX1NFUlZFUlNfRVJST1IQAg==');

@$core.Deprecated('Use technologyDescriptor instead')
const Technology$json = {
  '1': 'Technology',
  '2': [
    {'1': 'UNKNOWN_TECHNLOGY', '2': 0},
    {'1': 'NORDLYNX', '2': 1},
    {'1': 'OPENVPN_TCP', '2': 2},
    {'1': 'OPENVPN_UDP', '2': 3},
    {'1': 'OBFUSCATED_OPENVPN_TCP', '2': 4},
    {'1': 'OBFUSCATED_OPENVPN_UDP', '2': 5},
  ],
};

/// Descriptor for `Technology`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List technologyDescriptor = $convert.base64Decode(
    'CgpUZWNobm9sb2d5EhUKEVVOS05PV05fVEVDSE5MT0dZEAASDAoITk9SRExZTlgQARIPCgtPUE'
    'VOVlBOX1RDUBACEg8KC09QRU5WUE5fVURQEAMSGgoWT0JGVVNDQVRFRF9PUEVOVlBOX1RDUBAE'
    'EhoKFk9CRlVTQ0FURURfT1BFTlZQTl9VRFAQBQ==');

@$core.Deprecated('Use serverDescriptor instead')
const Server$json = {
  '1': 'Server',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 3, '10': 'id'},
    {'1': 'host_name', '3': 4, '4': 1, '5': 9, '10': 'hostName'},
    {'1': 'virtual', '3': 5, '4': 1, '5': 8, '10': 'virtual'},
    {
      '1': 'server_groups',
      '3': 6,
      '4': 3,
      '5': 14,
      '6': '.config.ServerGroup',
      '10': 'serverGroups'
    },
    {
      '1': 'technologies',
      '3': 7,
      '4': 3,
      '5': 14,
      '6': '.pb.Technology',
      '10': 'technologies'
    },
  ],
};

/// Descriptor for `Server`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List serverDescriptor = $convert.base64Decode(
    'CgZTZXJ2ZXISDgoCaWQYASABKANSAmlkEhsKCWhvc3RfbmFtZRgEIAEoCVIIaG9zdE5hbWUSGA'
    'oHdmlydHVhbBgFIAEoCFIHdmlydHVhbBI4Cg1zZXJ2ZXJfZ3JvdXBzGAYgAygOMhMuY29uZmln'
    'LlNlcnZlckdyb3VwUgxzZXJ2ZXJHcm91cHMSMgoMdGVjaG5vbG9naWVzGAcgAygOMg4ucGIuVG'
    'VjaG5vbG9neVIMdGVjaG5vbG9naWVz');

@$core.Deprecated('Use serverCityDescriptor instead')
const ServerCity$json = {
  '1': 'ServerCity',
  '2': [
    {'1': 'city_name', '3': 1, '4': 1, '5': 9, '10': 'cityName'},
    {
      '1': 'servers',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.pb.Server',
      '10': 'servers'
    },
  ],
};

/// Descriptor for `ServerCity`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List serverCityDescriptor = $convert.base64Decode(
    'CgpTZXJ2ZXJDaXR5EhsKCWNpdHlfbmFtZRgBIAEoCVIIY2l0eU5hbWUSJAoHc2VydmVycxgCIA'
    'MoCzIKLnBiLlNlcnZlclIHc2VydmVycw==');

@$core.Deprecated('Use serverCountryDescriptor instead')
const ServerCountry$json = {
  '1': 'ServerCountry',
  '2': [
    {'1': 'country_code', '3': 1, '4': 1, '5': 9, '10': 'countryCode'},
    {
      '1': 'cities',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.pb.ServerCity',
      '10': 'cities'
    },
    {'1': 'country_name', '3': 3, '4': 1, '5': 9, '10': 'countryName'},
  ],
};

/// Descriptor for `ServerCountry`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List serverCountryDescriptor = $convert.base64Decode(
    'Cg1TZXJ2ZXJDb3VudHJ5EiEKDGNvdW50cnlfY29kZRgBIAEoCVILY291bnRyeUNvZGUSJgoGY2'
    'l0aWVzGAIgAygLMg4ucGIuU2VydmVyQ2l0eVIGY2l0aWVzEiEKDGNvdW50cnlfbmFtZRgDIAEo'
    'CVILY291bnRyeU5hbWU=');

@$core.Deprecated('Use serversMapDescriptor instead')
const ServersMap$json = {
  '1': 'ServersMap',
  '2': [
    {
      '1': 'servers_by_country',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.pb.ServerCountry',
      '10': 'serversByCountry'
    },
  ],
};

/// Descriptor for `ServersMap`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List serversMapDescriptor = $convert.base64Decode(
    'CgpTZXJ2ZXJzTWFwEj8KEnNlcnZlcnNfYnlfY291bnRyeRgBIAMoCzIRLnBiLlNlcnZlckNvdW'
    '50cnlSEHNlcnZlcnNCeUNvdW50cnk=');

@$core.Deprecated('Use serversResponseDescriptor instead')
const ServersResponse$json = {
  '1': 'ServersResponse',
  '2': [
    {
      '1': 'servers',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.pb.ServersMap',
      '9': 0,
      '10': 'servers'
    },
    {
      '1': 'error',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.pb.ServersError',
      '9': 0,
      '10': 'error'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `ServersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List serversResponseDescriptor = $convert.base64Decode(
    'Cg9TZXJ2ZXJzUmVzcG9uc2USKgoHc2VydmVycxgBIAEoCzIOLnBiLlNlcnZlcnNNYXBIAFIHc2'
    'VydmVycxIoCgVlcnJvchgCIAEoDjIQLnBiLlNlcnZlcnNFcnJvckgAUgVlcnJvckIKCghyZXNw'
    'b25zZQ==');
