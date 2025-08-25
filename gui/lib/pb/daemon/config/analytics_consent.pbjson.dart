// This is a generated file - do not edit.
//
// Generated from analytics_consent.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use consentModeDescriptor instead')
const ConsentMode$json = {
  '1': 'ConsentMode',
  '2': [
    {'1': 'UNDEFINED', '2': 0},
    {'1': 'GRANTED', '2': 1},
    {'1': 'DENIED', '2': 2},
  ],
};

/// Descriptor for `ConsentMode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List consentModeDescriptor = $convert.base64Decode(
    'CgtDb25zZW50TW9kZRINCglVTkRFRklORUQQABILCgdHUkFOVEVEEAESCgoGREVOSUVEEAI=');
