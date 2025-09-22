// This is a generated file - do not edit.
//
// Generated from service_response.proto.

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
    {'1': 'NOT_LOGGED_IN', '2': 0},
    {'1': 'API_FAILURE', '2': 1},
    {'1': 'CONFIG_FAILURE', '2': 2},
  ],
};

/// Descriptor for `ServiceErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List serviceErrorCodeDescriptor = $convert.base64Decode(
    'ChBTZXJ2aWNlRXJyb3JDb2RlEhEKDU5PVF9MT0dHRURfSU4QABIPCgtBUElfRkFJTFVSRRABEh'
    'IKDkNPTkZJR19GQUlMVVJFEAI=');

@$core.Deprecated('Use meshnetErrorCodeDescriptor instead')
const MeshnetErrorCode$json = {
  '1': 'MeshnetErrorCode',
  '2': [
    {'1': 'NOT_REGISTERED', '2': 0},
    {'1': 'LIB_FAILURE', '2': 1},
    {'1': 'ALREADY_ENABLED', '2': 3},
    {'1': 'ALREADY_DISABLED', '2': 4},
    {'1': 'NOT_ENABLED', '2': 5},
    {'1': 'TECH_FAILURE', '2': 6},
    {'1': 'TUNNEL_CLOSED', '2': 7},
    {'1': 'CONFLICT_WITH_PQ', '2': 8},
    {'1': 'CONFLICT_WITH_PQ_SERVER', '2': 9},
  ],
};

/// Descriptor for `MeshnetErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List meshnetErrorCodeDescriptor = $convert.base64Decode(
    'ChBNZXNobmV0RXJyb3JDb2RlEhIKDk5PVF9SRUdJU1RFUkVEEAASDwoLTElCX0ZBSUxVUkUQAR'
    'ITCg9BTFJFQURZX0VOQUJMRUQQAxIUChBBTFJFQURZX0RJU0FCTEVEEAQSDwoLTk9UX0VOQUJM'
    'RUQQBRIQCgxURUNIX0ZBSUxVUkUQBhIRCg1UVU5ORUxfQ0xPU0VEEAcSFAoQQ09ORkxJQ1RfV0'
    'lUSF9QURAIEhsKF0NPTkZMSUNUX1dJVEhfUFFfU0VSVkVSEAk=');

@$core.Deprecated('Use meshnetResponseDescriptor instead')
const MeshnetResponse$json = {
  '1': 'MeshnetResponse',
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
      '1': 'service_error',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.meshpb.ServiceErrorCode',
      '9': 0,
      '10': 'serviceError'
    },
    {
      '1': 'meshnet_error',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.meshpb.MeshnetErrorCode',
      '9': 0,
      '10': 'meshnetError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `MeshnetResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List meshnetResponseDescriptor = $convert.base64Decode(
    'Cg9NZXNobmV0UmVzcG9uc2USJQoFZW1wdHkYASABKAsyDS5tZXNocGIuRW1wdHlIAFIFZW1wdH'
    'kSPwoNc2VydmljZV9lcnJvchgCIAEoDjIYLm1lc2hwYi5TZXJ2aWNlRXJyb3JDb2RlSABSDHNl'
    'cnZpY2VFcnJvchI/Cg1tZXNobmV0X2Vycm9yGAMgASgOMhgubWVzaHBiLk1lc2huZXRFcnJvck'
    'NvZGVIAFIMbWVzaG5ldEVycm9yQgoKCHJlc3BvbnNl');

@$core.Deprecated('Use serviceResponseDescriptor instead')
const ServiceResponse$json = {
  '1': 'ServiceResponse',
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
      '1': 'error_code',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.meshpb.ServiceErrorCode',
      '9': 0,
      '10': 'errorCode'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `ServiceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List serviceResponseDescriptor = $convert.base64Decode(
    'Cg9TZXJ2aWNlUmVzcG9uc2USJQoFZW1wdHkYASABKAsyDS5tZXNocGIuRW1wdHlIAFIFZW1wdH'
    'kSOQoKZXJyb3JfY29kZRgCIAEoDjIYLm1lc2hwYi5TZXJ2aWNlRXJyb3JDb2RlSABSCWVycm9y'
    'Q29kZUIKCghyZXNwb25zZQ==');

@$core.Deprecated('Use serviceBoolResponseDescriptor instead')
const ServiceBoolResponse$json = {
  '1': 'ServiceBoolResponse',
  '2': [
    {'1': 'value', '3': 1, '4': 1, '5': 8, '9': 0, '10': 'value'},
    {
      '1': 'error_code',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.meshpb.ServiceErrorCode',
      '9': 0,
      '10': 'errorCode'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `ServiceBoolResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List serviceBoolResponseDescriptor = $convert.base64Decode(
    'ChNTZXJ2aWNlQm9vbFJlc3BvbnNlEhYKBXZhbHVlGAEgASgISABSBXZhbHVlEjkKCmVycm9yX2'
    'NvZGUYAiABKA4yGC5tZXNocGIuU2VydmljZUVycm9yQ29kZUgAUgllcnJvckNvZGVCCgoIcmVz'
    'cG9uc2U=');

@$core.Deprecated('Use enabledStatusDescriptor instead')
const EnabledStatus$json = {
  '1': 'EnabledStatus',
  '2': [
    {'1': 'value', '3': 1, '4': 1, '5': 8, '10': 'value'},
    {'1': 'uid', '3': 2, '4': 1, '5': 13, '10': 'uid'},
  ],
};

/// Descriptor for `EnabledStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List enabledStatusDescriptor = $convert.base64Decode(
    'Cg1FbmFibGVkU3RhdHVzEhQKBXZhbHVlGAEgASgIUgV2YWx1ZRIQCgN1aWQYAiABKA1SA3VpZA'
    '==');

@$core.Deprecated('Use isEnabledResponseDescriptor instead')
const IsEnabledResponse$json = {
  '1': 'IsEnabledResponse',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.meshpb.EnabledStatus',
      '9': 0,
      '10': 'status'
    },
    {
      '1': 'error_code',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.meshpb.ServiceErrorCode',
      '9': 0,
      '10': 'errorCode'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `IsEnabledResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List isEnabledResponseDescriptor = $convert.base64Decode(
    'ChFJc0VuYWJsZWRSZXNwb25zZRIvCgZzdGF0dXMYASABKAsyFS5tZXNocGIuRW5hYmxlZFN0YX'
    'R1c0gAUgZzdGF0dXMSOQoKZXJyb3JfY29kZRgCIAEoDjIYLm1lc2hwYi5TZXJ2aWNlRXJyb3JD'
    'b2RlSABSCWVycm9yQ29kZUIKCghyZXNwb25zZQ==');
