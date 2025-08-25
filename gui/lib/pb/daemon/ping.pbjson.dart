// This is a generated file - do not edit.
//
// Generated from ping.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use pingResponseDescriptor instead')
const PingResponse$json = {
  '1': 'PingResponse',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 3, '10': 'type'},
    {'1': 'major', '3': 2, '4': 1, '5': 3, '10': 'major'},
    {'1': 'minor', '3': 3, '4': 1, '5': 3, '10': 'minor'},
    {'1': 'patch', '3': 4, '4': 1, '5': 3, '10': 'patch'},
    {'1': 'metadata', '3': 5, '4': 1, '5': 9, '10': 'metadata'},
  ],
};

/// Descriptor for `PingResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List pingResponseDescriptor = $convert.base64Decode(
    'CgxQaW5nUmVzcG9uc2USEgoEdHlwZRgBIAEoA1IEdHlwZRIUCgVtYWpvchgCIAEoA1IFbWFqb3'
    'ISFAoFbWlub3IYAyABKANSBW1pbm9yEhQKBXBhdGNoGAQgASgDUgVwYXRjaBIaCghtZXRhZGF0'
    'YRgFIAEoCVIIbWV0YWRhdGE=');
