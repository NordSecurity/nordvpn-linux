// This is a generated file - do not edit.
//
// Generated from status.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use connectionSourceDescriptor instead')
const ConnectionSource$json = {
  '1': 'ConnectionSource',
  '2': [
    {'1': 'UNKNOWN_SOURCE', '2': 0},
    {'1': 'MANUAL', '2': 1},
    {'1': 'AUTO', '2': 2},
  ],
};

/// Descriptor for `ConnectionSource`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List connectionSourceDescriptor = $convert.base64Decode(
    'ChBDb25uZWN0aW9uU291cmNlEhIKDlVOS05PV05fU09VUkNFEAASCgoGTUFOVUFMEAESCAoEQV'
    'VUTxAC');

@$core.Deprecated('Use connectionStateDescriptor instead')
const ConnectionState$json = {
  '1': 'ConnectionState',
  '2': [
    {'1': 'UNKNOWN_STATE', '2': 0},
    {'1': 'DISCONNECTED', '2': 1},
    {'1': 'CONNECTING', '2': 2},
    {'1': 'CONNECTED', '2': 3},
  ],
};

/// Descriptor for `ConnectionState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List connectionStateDescriptor = $convert.base64Decode(
    'Cg9Db25uZWN0aW9uU3RhdGUSEQoNVU5LTk9XTl9TVEFURRAAEhAKDERJU0NPTk5FQ1RFRBABEg'
    '4KCkNPTk5FQ1RJTkcQAhINCglDT05ORUNURUQQAw==');

@$core.Deprecated('Use connectionParametersDescriptor instead')
const ConnectionParameters$json = {
  '1': 'ConnectionParameters',
  '2': [
    {
      '1': 'source',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.pb.ConnectionSource',
      '10': 'source'
    },
    {'1': 'country', '3': 2, '4': 1, '5': 9, '10': 'country'},
    {'1': 'city', '3': 3, '4': 1, '5': 9, '10': 'city'},
    {
      '1': 'group',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.config.ServerGroup',
      '10': 'group'
    },
    {'1': 'server_name', '3': 5, '4': 1, '5': 9, '10': 'serverName'},
    {'1': 'country_code', '3': 6, '4': 1, '5': 9, '10': 'countryCode'},
  ],
};

/// Descriptor for `ConnectionParameters`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List connectionParametersDescriptor = $convert.base64Decode(
    'ChRDb25uZWN0aW9uUGFyYW1ldGVycxIsCgZzb3VyY2UYASABKA4yFC5wYi5Db25uZWN0aW9uU2'
    '91cmNlUgZzb3VyY2USGAoHY291bnRyeRgCIAEoCVIHY291bnRyeRISCgRjaXR5GAMgASgJUgRj'
    'aXR5EikKBWdyb3VwGAQgASgOMhMuY29uZmlnLlNlcnZlckdyb3VwUgVncm91cBIfCgtzZXJ2ZX'
    'JfbmFtZRgFIAEoCVIKc2VydmVyTmFtZRIhCgxjb3VudHJ5X2NvZGUYBiABKAlSC2NvdW50cnlD'
    'b2Rl');

@$core.Deprecated('Use statusResponseDescriptor instead')
const StatusResponse$json = {
  '1': 'StatusResponse',
  '2': [
    {
      '1': 'state',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.pb.ConnectionState',
      '10': 'state'
    },
    {
      '1': 'technology',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.config.Technology',
      '10': 'technology'
    },
    {
      '1': 'protocol',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.config.Protocol',
      '10': 'protocol'
    },
    {'1': 'ip', '3': 4, '4': 1, '5': 9, '10': 'ip'},
    {'1': 'hostname', '3': 5, '4': 1, '5': 9, '10': 'hostname'},
    {'1': 'country', '3': 6, '4': 1, '5': 9, '10': 'country'},
    {'1': 'city', '3': 7, '4': 1, '5': 9, '10': 'city'},
    {'1': 'download', '3': 8, '4': 1, '5': 4, '10': 'download'},
    {'1': 'upload', '3': 9, '4': 1, '5': 4, '10': 'upload'},
    {'1': 'uptime', '3': 10, '4': 1, '5': 3, '10': 'uptime'},
    {'1': 'name', '3': 11, '4': 1, '5': 9, '10': 'name'},
    {'1': 'virtualLocation', '3': 12, '4': 1, '5': 8, '10': 'virtualLocation'},
    {
      '1': 'parameters',
      '3': 13,
      '4': 1,
      '5': 11,
      '6': '.pb.ConnectionParameters',
      '10': 'parameters'
    },
    {'1': 'postQuantum', '3': 14, '4': 1, '5': 8, '10': 'postQuantum'},
    {'1': 'is_mesh_peer', '3': 15, '4': 1, '5': 8, '10': 'isMeshPeer'},
    {'1': 'by_user', '3': 16, '4': 1, '5': 8, '10': 'byUser'},
    {'1': 'country_code', '3': 17, '4': 1, '5': 9, '10': 'countryCode'},
    {'1': 'obfuscated', '3': 18, '4': 1, '5': 8, '10': 'obfuscated'},
  ],
};

/// Descriptor for `StatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List statusResponseDescriptor = $convert.base64Decode(
    'Cg5TdGF0dXNSZXNwb25zZRIpCgVzdGF0ZRgBIAEoDjITLnBiLkNvbm5lY3Rpb25TdGF0ZVIFc3'
    'RhdGUSMgoKdGVjaG5vbG9neRgCIAEoDjISLmNvbmZpZy5UZWNobm9sb2d5Ugp0ZWNobm9sb2d5'
    'EiwKCHByb3RvY29sGAMgASgOMhAuY29uZmlnLlByb3RvY29sUghwcm90b2NvbBIOCgJpcBgEIA'
    'EoCVICaXASGgoIaG9zdG5hbWUYBSABKAlSCGhvc3RuYW1lEhgKB2NvdW50cnkYBiABKAlSB2Nv'
    'dW50cnkSEgoEY2l0eRgHIAEoCVIEY2l0eRIaCghkb3dubG9hZBgIIAEoBFIIZG93bmxvYWQSFg'
    'oGdXBsb2FkGAkgASgEUgZ1cGxvYWQSFgoGdXB0aW1lGAogASgDUgZ1cHRpbWUSEgoEbmFtZRgL'
    'IAEoCVIEbmFtZRIoCg92aXJ0dWFsTG9jYXRpb24YDCABKAhSD3ZpcnR1YWxMb2NhdGlvbhI4Cg'
    'pwYXJhbWV0ZXJzGA0gASgLMhgucGIuQ29ubmVjdGlvblBhcmFtZXRlcnNSCnBhcmFtZXRlcnMS'
    'IAoLcG9zdFF1YW50dW0YDiABKAhSC3Bvc3RRdWFudHVtEiAKDGlzX21lc2hfcGVlchgPIAEoCF'
    'IKaXNNZXNoUGVlchIXCgdieV91c2VyGBAgASgIUgZieVVzZXISIQoMY291bnRyeV9jb2RlGBEg'
    'ASgJUgtjb3VudHJ5Q29kZRIeCgpvYmZ1c2NhdGVkGBIgASgIUgpvYmZ1c2NhdGVk');
