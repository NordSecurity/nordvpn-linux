// This is a generated file - do not edit.
//
// Generated from common.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use daemonApiVersionDescriptor instead')
const DaemonApiVersion$json = {
  '1': 'DaemonApiVersion',
  '2': [
    {'1': 'UNKNOWN_VERSION', '2': 0},
    {'1': 'CURRENT_VERSION', '2': 4},
  ],
};

/// Descriptor for `DaemonApiVersion`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List daemonApiVersionDescriptor = $convert.base64Decode(
    'ChBEYWVtb25BcGlWZXJzaW9uEhMKD1VOS05PV05fVkVSU0lPThAAEhMKD0NVUlJFTlRfVkVSU0'
    'lPThAE');

@$core.Deprecated('Use triStateDescriptor instead')
const TriState$json = {
  '1': 'TriState',
  '2': [
    {'1': 'UNKNOWN', '2': 0},
    {'1': 'DISABLED', '2': 1},
    {'1': 'ENABLED', '2': 2},
  ],
};

/// Descriptor for `TriState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List triStateDescriptor = $convert.base64Decode(
    'CghUcmlTdGF0ZRILCgdVTktOT1dOEAASDAoIRElTQUJMRUQQARILCgdFTkFCTEVEEAI=');

@$core.Deprecated('Use clientIDDescriptor instead')
const ClientID$json = {
  '1': 'ClientID',
  '2': [
    {'1': 'UNKNOWN_CLIENT', '2': 0},
    {'1': 'CLI', '2': 1},
    {'1': 'GUI', '2': 2},
    {'1': 'TRAY', '2': 3},
  ],
};

/// Descriptor for `ClientID`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List clientIDDescriptor = $convert.base64Decode(
    'CghDbGllbnRJRBISCg5VTktOT1dOX0NMSUVOVBAAEgcKA0NMSRABEgcKA0dVSRACEggKBFRSQV'
    'kQAw==');

@$core.Deprecated('Use getDaemonApiVersionRequestDescriptor instead')
const GetDaemonApiVersionRequest$json = {
  '1': 'GetDaemonApiVersionRequest',
};

/// Descriptor for `GetDaemonApiVersionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getDaemonApiVersionRequestDescriptor =
    $convert.base64Decode('ChpHZXREYWVtb25BcGlWZXJzaW9uUmVxdWVzdA==');

@$core.Deprecated('Use getDaemonApiVersionResponseDescriptor instead')
const GetDaemonApiVersionResponse$json = {
  '1': 'GetDaemonApiVersionResponse',
  '2': [
    {'1': 'apiVersion', '3': 1, '4': 1, '5': 13, '10': 'apiVersion'},
  ],
};

/// Descriptor for `GetDaemonApiVersionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getDaemonApiVersionResponseDescriptor =
    $convert.base64Decode(
        'ChtHZXREYWVtb25BcGlWZXJzaW9uUmVzcG9uc2USHgoKYXBpVmVyc2lvbhgBIAEoDVIKYXBpVm'
        'Vyc2lvbg==');

@$core.Deprecated('Use emptyDescriptor instead')
const Empty$json = {
  '1': 'Empty',
};

/// Descriptor for `Empty`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List emptyDescriptor =
    $convert.base64Decode('CgVFbXB0eQ==');

@$core.Deprecated('Use boolDescriptor instead')
const Bool$json = {
  '1': 'Bool',
  '2': [
    {'1': 'value', '3': 1, '4': 1, '5': 8, '10': 'value'},
  ],
};

/// Descriptor for `Bool`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List boolDescriptor =
    $convert.base64Decode('CgRCb29sEhQKBXZhbHVlGAEgASgIUgV2YWx1ZQ==');

@$core.Deprecated('Use payloadDescriptor instead')
const Payload$json = {
  '1': 'Payload',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 3, '10': 'type'},
    {'1': 'data', '3': 2, '4': 3, '5': 9, '10': 'data'},
  ],
};

/// Descriptor for `Payload`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List payloadDescriptor = $convert.base64Decode(
    'CgdQYXlsb2FkEhIKBHR5cGUYASABKANSBHR5cGUSEgoEZGF0YRgCIAMoCVIEZGF0YQ==');

@$core.Deprecated('Use allowlistDescriptor instead')
const Allowlist$json = {
  '1': 'Allowlist',
  '2': [
    {'1': 'ports', '3': 1, '4': 1, '5': 11, '6': '.pb.Ports', '10': 'ports'},
    {'1': 'subnets', '3': 2, '4': 3, '5': 9, '10': 'subnets'},
  ],
};

/// Descriptor for `Allowlist`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List allowlistDescriptor = $convert.base64Decode(
    'CglBbGxvd2xpc3QSHwoFcG9ydHMYASABKAsyCS5wYi5Qb3J0c1IFcG9ydHMSGAoHc3VibmV0cx'
    'gCIAMoCVIHc3VibmV0cw==');

@$core.Deprecated('Use portsDescriptor instead')
const Ports$json = {
  '1': 'Ports',
  '2': [
    {'1': 'udp', '3': 1, '4': 3, '5': 3, '10': 'udp'},
    {'1': 'tcp', '3': 2, '4': 3, '5': 3, '10': 'tcp'},
  ],
};

/// Descriptor for `Ports`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List portsDescriptor = $convert.base64Decode(
    'CgVQb3J0cxIQCgN1ZHAYASADKANSA3VkcBIQCgN0Y3AYAiADKANSA3RjcA==');

@$core.Deprecated('Use serverGroupDescriptor instead')
const ServerGroup$json = {
  '1': 'ServerGroup',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'virtualLocation', '3': 2, '4': 1, '5': 8, '10': 'virtualLocation'},
  ],
};

/// Descriptor for `ServerGroup`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List serverGroupDescriptor = $convert.base64Decode(
    'CgtTZXJ2ZXJHcm91cBISCgRuYW1lGAEgASgJUgRuYW1lEigKD3ZpcnR1YWxMb2NhdGlvbhgCIA'
    'EoCFIPdmlydHVhbExvY2F0aW9u');

@$core.Deprecated('Use serverGroupsListDescriptor instead')
const ServerGroupsList$json = {
  '1': 'ServerGroupsList',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 3, '10': 'type'},
    {
      '1': 'servers',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.pb.ServerGroup',
      '10': 'servers'
    },
  ],
};

/// Descriptor for `ServerGroupsList`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List serverGroupsListDescriptor = $convert.base64Decode(
    'ChBTZXJ2ZXJHcm91cHNMaXN0EhIKBHR5cGUYASABKANSBHR5cGUSKQoHc2VydmVycxgCIAMoCz'
    'IPLnBiLlNlcnZlckdyb3VwUgdzZXJ2ZXJz');
