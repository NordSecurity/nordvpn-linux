// This is a generated file - do not edit.
//
// Generated from fileshare.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use serviceErrorCodeDescriptor instead')
const ServiceErrorCode$json = {
  '1': 'ServiceErrorCode',
  '2': [
    {'1': 'MESH_NOT_ENABLED', '2': 0},
    {'1': 'INTERNAL_FAILURE', '2': 1},
  ],
};

/// Descriptor for `ServiceErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List serviceErrorCodeDescriptor = $convert.base64Decode(
    'ChBTZXJ2aWNlRXJyb3JDb2RlEhQKEE1FU0hfTk9UX0VOQUJMRUQQABIUChBJTlRFUk5BTF9GQU'
    'lMVVJFEAE=');

@$core.Deprecated('Use fileshareErrorCodeDescriptor instead')
const FileshareErrorCode$json = {
  '1': 'FileshareErrorCode',
  '2': [
    {'1': 'LIB_FAILURE', '2': 0},
    {'1': 'TRANSFER_NOT_FOUND', '2': 1},
    {'1': 'INVALID_PEER', '2': 2},
    {'1': 'FILE_NOT_FOUND', '2': 3},
    {'1': 'ACCEPT_ALL_FILES_FAILED', '2': 5},
    {'1': 'ACCEPT_OUTGOING', '2': 6},
    {'1': 'ALREADY_ACCEPTED', '2': 7},
    {'1': 'FILE_INVALIDATED', '2': 8},
    {'1': 'TRANSFER_INVALIDATED', '2': 9},
    {'1': 'TOO_MANY_FILES', '2': 10},
    {'1': 'DIRECTORY_TOO_DEEP', '2': 11},
    {'1': 'SENDING_NOT_ALLOWED', '2': 12},
    {'1': 'PEER_DISCONNECTED', '2': 13},
    {'1': 'FILE_NOT_IN_PROGRESS', '2': 14},
    {'1': 'TRANSFER_NOT_CREATED', '2': 15},
    {'1': 'NOT_ENOUGH_SPACE', '2': 16},
    {'1': 'ACCEPT_DIR_NOT_FOUND', '2': 17},
    {'1': 'ACCEPT_DIR_IS_A_SYMLINK', '2': 18},
    {'1': 'ACCEPT_DIR_IS_NOT_A_DIRECTORY', '2': 19},
    {'1': 'NO_FILES', '2': 20},
    {'1': 'ACCEPT_DIR_NO_PERMISSIONS', '2': 21},
    {'1': 'PURGE_FAILURE', '2': 22},
  ],
};

/// Descriptor for `FileshareErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List fileshareErrorCodeDescriptor = $convert.base64Decode(
    'ChJGaWxlc2hhcmVFcnJvckNvZGUSDwoLTElCX0ZBSUxVUkUQABIWChJUUkFOU0ZFUl9OT1RfRk'
    '9VTkQQARIQCgxJTlZBTElEX1BFRVIQAhISCg5GSUxFX05PVF9GT1VORBADEhsKF0FDQ0VQVF9B'
    'TExfRklMRVNfRkFJTEVEEAUSEwoPQUNDRVBUX09VVEdPSU5HEAYSFAoQQUxSRUFEWV9BQ0NFUF'
    'RFRBAHEhQKEEZJTEVfSU5WQUxJREFURUQQCBIYChRUUkFOU0ZFUl9JTlZBTElEQVRFRBAJEhIK'
    'DlRPT19NQU5ZX0ZJTEVTEAoSFgoSRElSRUNUT1JZX1RPT19ERUVQEAsSFwoTU0VORElOR19OT1'
    'RfQUxMT1dFRBAMEhUKEVBFRVJfRElTQ09OTkVDVEVEEA0SGAoURklMRV9OT1RfSU5fUFJPR1JF'
    'U1MQDhIYChRUUkFOU0ZFUl9OT1RfQ1JFQVRFRBAPEhQKEE5PVF9FTk9VR0hfU1BBQ0UQEBIYCh'
    'RBQ0NFUFRfRElSX05PVF9GT1VORBAREhsKF0FDQ0VQVF9ESVJfSVNfQV9TWU1MSU5LEBISIQod'
    'QUNDRVBUX0RJUl9JU19OT1RfQV9ESVJFQ1RPUlkQExIMCghOT19GSUxFUxAUEh0KGUFDQ0VQVF'
    '9ESVJfTk9fUEVSTUlTU0lPTlMQFRIRCg1QVVJHRV9GQUlMVVJFEBY=');

@$core.Deprecated('Use setNotificationsStatusDescriptor instead')
const SetNotificationsStatus$json = {
  '1': 'SetNotificationsStatus',
  '2': [
    {'1': 'SET_SUCCESS', '2': 0},
    {'1': 'NOTHING_TO_DO', '2': 1},
    {'1': 'SET_FAILURE', '2': 2},
  ],
};

/// Descriptor for `SetNotificationsStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List setNotificationsStatusDescriptor =
    $convert.base64Decode(
        'ChZTZXROb3RpZmljYXRpb25zU3RhdHVzEg8KC1NFVF9TVUNDRVNTEAASEQoNTk9USElOR19UT1'
        '9ETxABEg8KC1NFVF9GQUlMVVJFEAI=');

@$core.Deprecated('Use emptyDescriptor instead')
const Empty$json = {
  '1': 'Empty',
};

/// Descriptor for `Empty`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List emptyDescriptor =
    $convert.base64Decode('CgVFbXB0eQ==');

@$core.Deprecated('Use errorDescriptor instead')
const Error$json = {
  '1': 'Error',
  '2': [
    {
      '1': 'empty',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.filesharepb.Empty',
      '9': 0,
      '10': 'empty'
    },
    {
      '1': 'service_error',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.filesharepb.ServiceErrorCode',
      '9': 0,
      '10': 'serviceError'
    },
    {
      '1': 'fileshare_error',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.filesharepb.FileshareErrorCode',
      '9': 0,
      '10': 'fileshareError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `Error`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List errorDescriptor = $convert.base64Decode(
    'CgVFcnJvchIqCgVlbXB0eRgBIAEoCzISLmZpbGVzaGFyZXBiLkVtcHR5SABSBWVtcHR5EkQKDX'
    'NlcnZpY2VfZXJyb3IYAiABKA4yHS5maWxlc2hhcmVwYi5TZXJ2aWNlRXJyb3JDb2RlSABSDHNl'
    'cnZpY2VFcnJvchJKCg9maWxlc2hhcmVfZXJyb3IYAyABKA4yHy5maWxlc2hhcmVwYi5GaWxlc2'
    'hhcmVFcnJvckNvZGVIAFIOZmlsZXNoYXJlRXJyb3JCCgoIcmVzcG9uc2U=');

@$core.Deprecated('Use sendRequestDescriptor instead')
const SendRequest$json = {
  '1': 'SendRequest',
  '2': [
    {'1': 'peer', '3': 1, '4': 1, '5': 9, '10': 'peer'},
    {'1': 'paths', '3': 2, '4': 3, '5': 9, '10': 'paths'},
    {'1': 'silent', '3': 3, '4': 1, '5': 8, '10': 'silent'},
  ],
};

/// Descriptor for `SendRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List sendRequestDescriptor = $convert.base64Decode(
    'CgtTZW5kUmVxdWVzdBISCgRwZWVyGAEgASgJUgRwZWVyEhQKBXBhdGhzGAIgAygJUgVwYXRocx'
    'IWCgZzaWxlbnQYAyABKAhSBnNpbGVudA==');

@$core.Deprecated('Use acceptRequestDescriptor instead')
const AcceptRequest$json = {
  '1': 'AcceptRequest',
  '2': [
    {'1': 'transfer_id', '3': 1, '4': 1, '5': 9, '10': 'transferId'},
    {'1': 'dst_path', '3': 2, '4': 1, '5': 9, '10': 'dstPath'},
    {'1': 'silent', '3': 3, '4': 1, '5': 8, '10': 'silent'},
    {'1': 'files', '3': 4, '4': 3, '5': 9, '10': 'files'},
  ],
};

/// Descriptor for `AcceptRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List acceptRequestDescriptor = $convert.base64Decode(
    'Cg1BY2NlcHRSZXF1ZXN0Eh8KC3RyYW5zZmVyX2lkGAEgASgJUgp0cmFuc2ZlcklkEhkKCGRzdF'
    '9wYXRoGAIgASgJUgdkc3RQYXRoEhYKBnNpbGVudBgDIAEoCFIGc2lsZW50EhQKBWZpbGVzGAQg'
    'AygJUgVmaWxlcw==');

@$core.Deprecated('Use statusResponseDescriptor instead')
const StatusResponse$json = {
  '1': 'StatusResponse',
  '2': [
    {
      '1': 'error',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.filesharepb.Error',
      '10': 'error'
    },
    {'1': 'transfer_id', '3': 2, '4': 1, '5': 9, '10': 'transferId'},
    {'1': 'progress', '3': 3, '4': 1, '5': 13, '10': 'progress'},
    {
      '1': 'status',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.filesharepb.Status',
      '10': 'status'
    },
  ],
};

/// Descriptor for `StatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List statusResponseDescriptor = $convert.base64Decode(
    'Cg5TdGF0dXNSZXNwb25zZRIoCgVlcnJvchgBIAEoCzISLmZpbGVzaGFyZXBiLkVycm9yUgVlcn'
    'JvchIfCgt0cmFuc2Zlcl9pZBgCIAEoCVIKdHJhbnNmZXJJZBIaCghwcm9ncmVzcxgDIAEoDVII'
    'cHJvZ3Jlc3MSKwoGc3RhdHVzGAQgASgOMhMuZmlsZXNoYXJlcGIuU3RhdHVzUgZzdGF0dXM=');

@$core.Deprecated('Use cancelRequestDescriptor instead')
const CancelRequest$json = {
  '1': 'CancelRequest',
  '2': [
    {'1': 'transfer_id', '3': 1, '4': 1, '5': 9, '10': 'transferId'},
  ],
};

/// Descriptor for `CancelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelRequestDescriptor = $convert.base64Decode(
    'Cg1DYW5jZWxSZXF1ZXN0Eh8KC3RyYW5zZmVyX2lkGAEgASgJUgp0cmFuc2Zlcklk');

@$core.Deprecated('Use listResponseDescriptor instead')
const ListResponse$json = {
  '1': 'ListResponse',
  '2': [
    {
      '1': 'error',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.filesharepb.Error',
      '10': 'error'
    },
    {
      '1': 'transfers',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.filesharepb.Transfer',
      '10': 'transfers'
    },
  ],
};

/// Descriptor for `ListResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listResponseDescriptor = $convert.base64Decode(
    'CgxMaXN0UmVzcG9uc2USKAoFZXJyb3IYASABKAsyEi5maWxlc2hhcmVwYi5FcnJvclIFZXJyb3'
    'ISMwoJdHJhbnNmZXJzGAIgAygLMhUuZmlsZXNoYXJlcGIuVHJhbnNmZXJSCXRyYW5zZmVycw==');

@$core.Deprecated('Use cancelFileRequestDescriptor instead')
const CancelFileRequest$json = {
  '1': 'CancelFileRequest',
  '2': [
    {'1': 'transfer_id', '3': 1, '4': 1, '5': 9, '10': 'transferId'},
    {'1': 'file_path', '3': 2, '4': 1, '5': 9, '10': 'filePath'},
  ],
};

/// Descriptor for `CancelFileRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelFileRequestDescriptor = $convert.base64Decode(
    'ChFDYW5jZWxGaWxlUmVxdWVzdBIfCgt0cmFuc2Zlcl9pZBgBIAEoCVIKdHJhbnNmZXJJZBIbCg'
    'lmaWxlX3BhdGgYAiABKAlSCGZpbGVQYXRo');

@$core.Deprecated('Use setNotificationsRequestDescriptor instead')
const SetNotificationsRequest$json = {
  '1': 'SetNotificationsRequest',
  '2': [
    {'1': 'enable', '3': 1, '4': 1, '5': 8, '10': 'enable'},
  ],
};

/// Descriptor for `SetNotificationsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setNotificationsRequestDescriptor =
    $convert.base64Decode(
        'ChdTZXROb3RpZmljYXRpb25zUmVxdWVzdBIWCgZlbmFibGUYASABKAhSBmVuYWJsZQ==');

@$core.Deprecated('Use setNotificationsResponseDescriptor instead')
const SetNotificationsResponse$json = {
  '1': 'SetNotificationsResponse',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.filesharepb.SetNotificationsStatus',
      '10': 'status'
    },
  ],
};

/// Descriptor for `SetNotificationsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setNotificationsResponseDescriptor =
    $convert.base64Decode(
        'ChhTZXROb3RpZmljYXRpb25zUmVzcG9uc2USOwoGc3RhdHVzGAEgASgOMiMuZmlsZXNoYXJlcG'
        'IuU2V0Tm90aWZpY2F0aW9uc1N0YXR1c1IGc3RhdHVz');

@$core.Deprecated('Use purgeTransfersUntilRequestDescriptor instead')
const PurgeTransfersUntilRequest$json = {
  '1': 'PurgeTransfersUntilRequest',
  '2': [
    {
      '1': 'until',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'until'
    },
  ],
};

/// Descriptor for `PurgeTransfersUntilRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List purgeTransfersUntilRequestDescriptor =
    $convert.base64Decode(
        'ChpQdXJnZVRyYW5zZmVyc1VudGlsUmVxdWVzdBIwCgV1bnRpbBgBIAEoCzIaLmdvb2dsZS5wcm'
        '90b2J1Zi5UaW1lc3RhbXBSBXVudGls');
