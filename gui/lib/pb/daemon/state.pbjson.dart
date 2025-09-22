// This is a generated file - do not edit.
//
// Generated from state.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use appStateErrorDescriptor instead')
const AppStateError$json = {
  '1': 'AppStateError',
  '2': [
    {'1': 'FAILED_TO_GET_UID', '2': 0},
  ],
};

/// Descriptor for `AppStateError`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List appStateErrorDescriptor = $convert
    .base64Decode('Cg1BcHBTdGF0ZUVycm9yEhUKEUZBSUxFRF9UT19HRVRfVUlEEAA=');

@$core.Deprecated('Use loginEventTypeDescriptor instead')
const LoginEventType$json = {
  '1': 'LoginEventType',
  '2': [
    {'1': 'LOGIN', '2': 0},
    {'1': 'LOGOUT', '2': 1},
  ],
};

/// Descriptor for `LoginEventType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List loginEventTypeDescriptor = $convert
    .base64Decode('Cg5Mb2dpbkV2ZW50VHlwZRIJCgVMT0dJThAAEgoKBkxPR09VVBAB');

@$core.Deprecated('Use updateEventDescriptor instead')
const UpdateEvent$json = {
  '1': 'UpdateEvent',
  '2': [
    {'1': 'SERVERS_LIST_UPDATE', '2': 0},
  ],
};

/// Descriptor for `UpdateEvent`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List updateEventDescriptor = $convert
    .base64Decode('CgtVcGRhdGVFdmVudBIXChNTRVJWRVJTX0xJU1RfVVBEQVRFEAA=');

@$core.Deprecated('Use loginEventDescriptor instead')
const LoginEvent$json = {
  '1': 'LoginEvent',
  '2': [
    {
      '1': 'type',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.pb.LoginEventType',
      '10': 'type'
    },
  ],
};

/// Descriptor for `LoginEvent`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginEventDescriptor = $convert.base64Decode(
    'CgpMb2dpbkV2ZW50EiYKBHR5cGUYASABKA4yEi5wYi5Mb2dpbkV2ZW50VHlwZVIEdHlwZQ==');

@$core.Deprecated('Use accountModificationDescriptor instead')
const AccountModification$json = {
  '1': 'AccountModification',
  '2': [
    {
      '1': 'expires_at',
      '3': 1,
      '4': 1,
      '5': 9,
      '9': 0,
      '10': 'expiresAt',
      '17': true
    },
  ],
  '8': [
    {'1': '_expires_at'},
  ],
};

/// Descriptor for `AccountModification`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List accountModificationDescriptor = $convert.base64Decode(
    'ChNBY2NvdW50TW9kaWZpY2F0aW9uEiIKCmV4cGlyZXNfYXQYASABKAlIAFIJZXhwaXJlc0F0iA'
    'EBQg0KC19leHBpcmVzX2F0');

@$core.Deprecated('Use appStateDescriptor instead')
const AppState$json = {
  '1': 'AppState',
  '2': [
    {
      '1': 'error',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.pb.AppStateError',
      '9': 0,
      '10': 'error'
    },
    {
      '1': 'connection_status',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.pb.StatusResponse',
      '9': 0,
      '10': 'connectionStatus'
    },
    {
      '1': 'login_event',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.pb.LoginEvent',
      '9': 0,
      '10': 'loginEvent'
    },
    {
      '1': 'settings_change',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.pb.Settings',
      '9': 0,
      '10': 'settingsChange'
    },
    {
      '1': 'update_event',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.pb.UpdateEvent',
      '9': 0,
      '10': 'updateEvent'
    },
    {
      '1': 'account_modification',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.pb.AccountModification',
      '9': 0,
      '10': 'accountModification'
    },
  ],
  '8': [
    {'1': 'state'},
  ],
};

/// Descriptor for `AppState`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List appStateDescriptor = $convert.base64Decode(
    'CghBcHBTdGF0ZRIpCgVlcnJvchgBIAEoDjIRLnBiLkFwcFN0YXRlRXJyb3JIAFIFZXJyb3ISQQ'
    'oRY29ubmVjdGlvbl9zdGF0dXMYAiABKAsyEi5wYi5TdGF0dXNSZXNwb25zZUgAUhBjb25uZWN0'
    'aW9uU3RhdHVzEjEKC2xvZ2luX2V2ZW50GAMgASgLMg4ucGIuTG9naW5FdmVudEgAUgpsb2dpbk'
    'V2ZW50EjcKD3NldHRpbmdzX2NoYW5nZRgEIAEoCzIMLnBiLlNldHRpbmdzSABSDnNldHRpbmdz'
    'Q2hhbmdlEjQKDHVwZGF0ZV9ldmVudBgFIAEoDjIPLnBiLlVwZGF0ZUV2ZW50SABSC3VwZGF0ZU'
    'V2ZW50EkwKFGFjY291bnRfbW9kaWZpY2F0aW9uGAYgASgLMhcucGIuQWNjb3VudE1vZGlmaWNh'
    'dGlvbkgAUhNhY2NvdW50TW9kaWZpY2F0aW9uQgcKBXN0YXRl');
