// This is a generated file - do not edit.
//
// Generated from pause.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use pauseInvervalDescriptor instead')
const PauseInverval$json = {
  '1': 'PauseInverval',
  '2': [
    {'1': 'PAUSE_5_MIN', '2': 0},
    {'1': 'PAUSE_15_MIN', '2': 1},
    {'1': 'PAUSE_30_MIN', '2': 2},
    {'1': 'PAUSE_1_HOUR', '2': 3},
    {'1': 'PAUSE_24_HOURS', '2': 4},
  ],
};

/// Descriptor for `PauseInverval`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List pauseInvervalDescriptor = $convert.base64Decode(
    'Cg1QYXVzZUludmVydmFsEg8KC1BBVVNFXzVfTUlOEAASEAoMUEFVU0VfMTVfTUlOEAESEAoMUE'
    'FVU0VfMzBfTUlOEAISEAoMUEFVU0VfMV9IT1VSEAMSEgoOUEFVU0VfMjRfSE9VUlMQBA==');

@$core.Deprecated('Use pauseRequestDescriptor instead')
const PauseRequest$json = {
  '1': 'PauseRequest',
  '2': [
    {
      '1': 'interval',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.pb.PauseInverval',
      '10': 'interval'
    },
  ],
};

/// Descriptor for `PauseRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List pauseRequestDescriptor = $convert.base64Decode(
    'CgxQYXVzZVJlcXVlc3QSLQoIaW50ZXJ2YWwYASABKA4yES5wYi5QYXVzZUludmVydmFsUghpbn'
    'RlcnZhbA==');
