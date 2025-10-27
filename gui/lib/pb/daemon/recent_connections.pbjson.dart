// This is a generated file - do not edit.
//
// Generated from recent_connections.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use recentConnectionsResponseDescriptor instead')
const RecentConnectionsResponse$json = {
  '1': 'RecentConnectionsResponse',
  '2': [
    {
      '1': 'connections',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.pb.RecentConnectionModel',
      '10': 'connections'
    },
  ],
};

/// Descriptor for `RecentConnectionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List recentConnectionsResponseDescriptor =
    $convert.base64Decode(
        'ChlSZWNlbnRDb25uZWN0aW9uc1Jlc3BvbnNlEjsKC2Nvbm5lY3Rpb25zGAEgAygLMhkucGIuUm'
        'VjZW50Q29ubmVjdGlvbk1vZGVsUgtjb25uZWN0aW9ucw==');

@$core.Deprecated('Use recentConnectionModelDescriptor instead')
const RecentConnectionModel$json = {
  '1': 'RecentConnectionModel',
  '2': [
    {'1': 'country', '3': 1, '4': 1, '5': 9, '10': 'country'},
    {'1': 'city', '3': 2, '4': 1, '5': 9, '10': 'city'},
    {
      '1': 'group',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.config.ServerGroup',
      '10': 'group'
    },
    {'1': 'country_code', '3': 4, '4': 1, '5': 9, '10': 'countryCode'},
    {
      '1': 'specific_server_name',
      '3': 5,
      '4': 1,
      '5': 9,
      '10': 'specificServerName'
    },
    {'1': 'specific_server', '3': 6, '4': 1, '5': 9, '10': 'specificServer'},
    {
      '1': 'connection_type',
      '3': 7,
      '4': 1,
      '5': 14,
      '6': '.pb.ServerSelectionRule',
      '10': 'connectionType'
    },
    {'1': 'is_virtual', '3': 8, '4': 1, '5': 8, '10': 'isVirtual'},
  ],
};

/// Descriptor for `RecentConnectionModel`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List recentConnectionModelDescriptor = $convert.base64Decode(
    'ChVSZWNlbnRDb25uZWN0aW9uTW9kZWwSGAoHY291bnRyeRgBIAEoCVIHY291bnRyeRISCgRjaX'
    'R5GAIgASgJUgRjaXR5EikKBWdyb3VwGAMgASgOMhMuY29uZmlnLlNlcnZlckdyb3VwUgVncm91'
    'cBIhCgxjb3VudHJ5X2NvZGUYBCABKAlSC2NvdW50cnlDb2RlEjAKFHNwZWNpZmljX3NlcnZlcl'
    '9uYW1lGAUgASgJUhJzcGVjaWZpY1NlcnZlck5hbWUSJwoPc3BlY2lmaWNfc2VydmVyGAYgASgJ'
    'Ug5zcGVjaWZpY1NlcnZlchJACg9jb25uZWN0aW9uX3R5cGUYByABKA4yFy5wYi5TZXJ2ZXJTZW'
    'xlY3Rpb25SdWxlUg5jb25uZWN0aW9uVHlwZRIdCgppc192aXJ0dWFsGAggASgIUglpc1ZpcnR1'
    'YWw=');

@$core.Deprecated('Use recentConnectionsRequestDescriptor instead')
const RecentConnectionsRequest$json = {
  '1': 'RecentConnectionsRequest',
  '2': [
    {'1': 'limit', '3': 1, '4': 1, '5': 3, '9': 0, '10': 'limit', '17': true},
  ],
  '8': [
    {'1': '_limit'},
  ],
};

/// Descriptor for `RecentConnectionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List recentConnectionsRequestDescriptor =
    $convert.base64Decode(
        'ChhSZWNlbnRDb25uZWN0aW9uc1JlcXVlc3QSGQoFbGltaXQYASABKANIAFIFbGltaXSIAQFCCA'
        'oGX2xpbWl0');
