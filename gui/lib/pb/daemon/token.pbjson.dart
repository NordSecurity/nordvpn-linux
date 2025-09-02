// This is a generated file - do not edit.
//
// Generated from token.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use tokenInfoResponseDescriptor instead')
const TokenInfoResponse$json = {
  '1': 'TokenInfoResponse',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 3, '10': 'type'},
    {'1': 'token', '3': 2, '4': 1, '5': 9, '10': 'token'},
    {'1': 'expires_at', '3': 3, '4': 1, '5': 9, '10': 'expiresAt'},
    {
      '1': 'trusted_pass_token',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'trustedPassToken'
    },
    {
      '1': 'trusted_pass_owner_id',
      '3': 5,
      '4': 1,
      '5': 9,
      '10': 'trustedPassOwnerId'
    },
  ],
};

/// Descriptor for `TokenInfoResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List tokenInfoResponseDescriptor = $convert.base64Decode(
    'ChFUb2tlbkluZm9SZXNwb25zZRISCgR0eXBlGAEgASgDUgR0eXBlEhQKBXRva2VuGAIgASgJUg'
    'V0b2tlbhIdCgpleHBpcmVzX2F0GAMgASgJUglleHBpcmVzQXQSLAoSdHJ1c3RlZF9wYXNzX3Rv'
    'a2VuGAQgASgJUhB0cnVzdGVkUGFzc1Rva2VuEjEKFXRydXN0ZWRfcGFzc19vd25lcl9pZBgFIA'
    'EoCVISdHJ1c3RlZFBhc3NPd25lcklk');
