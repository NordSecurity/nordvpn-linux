// This is a generated file - do not edit.
//
// Generated from fsnotify.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use newTransferNotificationDescriptor instead')
const NewTransferNotification$json = {
  '1': 'NewTransferNotification',
  '2': [
    {'1': 'identifier', '3': 1, '4': 1, '5': 9, '10': 'identifier'},
    {'1': 'os', '3': 2, '4': 1, '5': 9, '10': 'os'},
    {'1': 'file_name', '3': 3, '4': 1, '5': 9, '10': 'fileName'},
    {'1': 'file_count', '3': 4, '4': 1, '5': 5, '10': 'fileCount'},
    {'1': 'transfer_id', '3': 5, '4': 1, '5': 9, '10': 'transferId'},
  ],
};

/// Descriptor for `NewTransferNotification`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List newTransferNotificationDescriptor = $convert.base64Decode(
    'ChdOZXdUcmFuc2Zlck5vdGlmaWNhdGlvbhIeCgppZGVudGlmaWVyGAEgASgJUgppZGVudGlmaW'
    'VyEg4KAm9zGAIgASgJUgJvcxIbCglmaWxlX25hbWUYAyABKAlSCGZpbGVOYW1lEh0KCmZpbGVf'
    'Y291bnQYBCABKAVSCWZpbGVDb3VudBIfCgt0cmFuc2Zlcl9pZBgFIAEoCVIKdHJhbnNmZXJJZA'
    '==');

@$core.Deprecated('Use notifyNewTransferResponseDescriptor instead')
const NotifyNewTransferResponse$json = {
  '1': 'NotifyNewTransferResponse',
  '2': [
    {
      '1': 'empty',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.meshpb.Empty',
      '9': 0,
      '10': 'empty'
    },
    {
      '1': 'update_peer_error_code',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.meshpb.UpdatePeerErrorCode',
      '9': 0,
      '10': 'updatePeerErrorCode'
    },
    {
      '1': 'service_error_code',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.meshpb.ServiceErrorCode',
      '9': 0,
      '10': 'serviceErrorCode'
    },
    {
      '1': 'meshnet_error_code',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.meshpb.MeshnetErrorCode',
      '9': 0,
      '10': 'meshnetErrorCode'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `NotifyNewTransferResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List notifyNewTransferResponseDescriptor = $convert.base64Decode(
    'ChlOb3RpZnlOZXdUcmFuc2ZlclJlc3BvbnNlEiUKBWVtcHR5GAEgASgLMg0ubWVzaHBiLkVtcH'
    'R5SABSBWVtcHR5ElIKFnVwZGF0ZV9wZWVyX2Vycm9yX2NvZGUYAiABKA4yGy5tZXNocGIuVXBk'
    'YXRlUGVlckVycm9yQ29kZUgAUhN1cGRhdGVQZWVyRXJyb3JDb2RlEkgKEnNlcnZpY2VfZXJyb3'
    'JfY29kZRgDIAEoDjIYLm1lc2hwYi5TZXJ2aWNlRXJyb3JDb2RlSABSEHNlcnZpY2VFcnJvckNv'
    'ZGUSSAoSbWVzaG5ldF9lcnJvcl9jb2RlGAQgASgOMhgubWVzaHBiLk1lc2huZXRFcnJvckNvZG'
    'VIAFIQbWVzaG5ldEVycm9yQ29kZUIKCghyZXNwb25zZQ==');
