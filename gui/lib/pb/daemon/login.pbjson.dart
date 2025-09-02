// This is a generated file - do not edit.
//
// Generated from login.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use loginTypeDescriptor instead')
const LoginType$json = {
  '1': 'LoginType',
  '2': [
    {'1': 'LoginType_UNKNOWN', '2': 0},
    {'1': 'LoginType_LOGIN', '2': 1},
    {'1': 'LoginType_SIGNUP', '2': 2},
  ],
};

/// Descriptor for `LoginType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List loginTypeDescriptor = $convert.base64Decode(
    'CglMb2dpblR5cGUSFQoRTG9naW5UeXBlX1VOS05PV04QABITCg9Mb2dpblR5cGVfTE9HSU4QAR'
    'IUChBMb2dpblR5cGVfU0lHTlVQEAI=');

@$core.Deprecated('Use loginStatusDescriptor instead')
const LoginStatus$json = {
  '1': 'LoginStatus',
  '2': [
    {'1': 'SUCCESS', '2': 0},
    {'1': 'UNKNOWN_OAUTH2_ERROR', '2': 1},
    {'1': 'ALREADY_LOGGED_IN', '2': 2},
    {'1': 'NO_NET', '2': 3},
    {'1': 'CONSENT_MISSING', '2': 4},
  ],
};

/// Descriptor for `LoginStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List loginStatusDescriptor = $convert.base64Decode(
    'CgtMb2dpblN0YXR1cxILCgdTVUNDRVNTEAASGAoUVU5LTk9XTl9PQVVUSDJfRVJST1IQARIVCh'
    'FBTFJFQURZX0xPR0dFRF9JThACEgoKBk5PX05FVBADEhMKD0NPTlNFTlRfTUlTU0lORxAE');

@$core.Deprecated('Use loginOAuth2RequestDescriptor instead')
const LoginOAuth2Request$json = {
  '1': 'LoginOAuth2Request',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 14, '6': '.pb.LoginType', '10': 'type'},
  ],
};

/// Descriptor for `LoginOAuth2Request`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginOAuth2RequestDescriptor = $convert.base64Decode(
    'ChJMb2dpbk9BdXRoMlJlcXVlc3QSIQoEdHlwZRgBIAEoDjINLnBiLkxvZ2luVHlwZVIEdHlwZQ'
    '==');

@$core.Deprecated('Use loginOAuth2CallbackRequestDescriptor instead')
const LoginOAuth2CallbackRequest$json = {
  '1': 'LoginOAuth2CallbackRequest',
  '2': [
    {'1': 'token', '3': 1, '4': 1, '5': 9, '10': 'token'},
    {'1': 'type', '3': 2, '4': 1, '5': 14, '6': '.pb.LoginType', '10': 'type'},
  ],
};

/// Descriptor for `LoginOAuth2CallbackRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginOAuth2CallbackRequestDescriptor =
    $convert.base64Decode(
        'ChpMb2dpbk9BdXRoMkNhbGxiYWNrUmVxdWVzdBIUCgV0b2tlbhgBIAEoCVIFdG9rZW4SIQoEdH'
        'lwZRgCIAEoDjINLnBiLkxvZ2luVHlwZVIEdHlwZQ==');

@$core.Deprecated('Use loginResponseDescriptor instead')
const LoginResponse$json = {
  '1': 'LoginResponse',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 3, '10': 'type'},
    {'1': 'url', '3': 5, '4': 1, '5': 9, '10': 'url'},
  ],
};

/// Descriptor for `LoginResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginResponseDescriptor = $convert.base64Decode(
    'Cg1Mb2dpblJlc3BvbnNlEhIKBHR5cGUYASABKANSBHR5cGUSEAoDdXJsGAUgASgJUgN1cmw=');

@$core.Deprecated('Use loginOAuth2ResponseDescriptor instead')
const LoginOAuth2Response$json = {
  '1': 'LoginOAuth2Response',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.pb.LoginStatus',
      '10': 'status'
    },
    {'1': 'url', '3': 2, '4': 1, '5': 9, '10': 'url'},
  ],
};

/// Descriptor for `LoginOAuth2Response`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginOAuth2ResponseDescriptor = $convert.base64Decode(
    'ChNMb2dpbk9BdXRoMlJlc3BvbnNlEicKBnN0YXR1cxgBIAEoDjIPLnBiLkxvZ2luU3RhdHVzUg'
    'ZzdGF0dXMSEAoDdXJsGAIgASgJUgN1cmw=');

@$core.Deprecated('Use loginOAuth2CallbackResponseDescriptor instead')
const LoginOAuth2CallbackResponse$json = {
  '1': 'LoginOAuth2CallbackResponse',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.pb.LoginStatus',
      '10': 'status'
    },
  ],
};

/// Descriptor for `LoginOAuth2CallbackResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginOAuth2CallbackResponseDescriptor =
    $convert.base64Decode(
        'ChtMb2dpbk9BdXRoMkNhbGxiYWNrUmVzcG9uc2USJwoGc3RhdHVzGAEgASgOMg8ucGIuTG9naW'
        '5TdGF0dXNSBnN0YXR1cw==');

@$core.Deprecated('Use isLoggedInResponseDescriptor instead')
const IsLoggedInResponse$json = {
  '1': 'IsLoggedInResponse',
  '2': [
    {'1': 'is_logged_in', '3': 1, '4': 1, '5': 8, '10': 'isLoggedIn'},
    {
      '1': 'status',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.pb.LoginStatus',
      '10': 'status'
    },
  ],
};

/// Descriptor for `IsLoggedInResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List isLoggedInResponseDescriptor = $convert.base64Decode(
    'ChJJc0xvZ2dlZEluUmVzcG9uc2USIAoMaXNfbG9nZ2VkX2luGAEgASgIUgppc0xvZ2dlZEluEi'
    'cKBnN0YXR1cxgCIAEoDjIPLnBiLkxvZ2luU3RhdHVzUgZzdGF0dXM=');
