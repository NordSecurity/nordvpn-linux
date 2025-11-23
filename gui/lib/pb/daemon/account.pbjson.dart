// This is a generated file - do not edit.
//
// Generated from account.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use dedidcatedIPServiceDescriptor instead')
const DedidcatedIPService$json = {
  '1': 'DedidcatedIPService',
  '2': [
    {'1': 'server_ids', '3': 1, '4': 3, '5': 3, '10': 'serverIds'},
    {
      '1': 'dedicated_ip_expires_at',
      '3': 2,
      '4': 1,
      '5': 9,
      '10': 'dedicatedIpExpiresAt'
    },
  ],
};

/// Descriptor for `DedidcatedIPService`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List dedidcatedIPServiceDescriptor = $convert.base64Decode(
    'ChNEZWRpZGNhdGVkSVBTZXJ2aWNlEh0KCnNlcnZlcl9pZHMYASADKANSCXNlcnZlcklkcxI1Ch'
    'dkZWRpY2F0ZWRfaXBfZXhwaXJlc19hdBgCIAEoCVIUZGVkaWNhdGVkSXBFeHBpcmVzQXQ=');

@$core.Deprecated('Use accountResponseDescriptor instead')
const AccountResponse$json = {
  '1': 'AccountResponse',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 3, '10': 'type'},
    {'1': 'username', '3': 2, '4': 1, '5': 9, '10': 'username'},
    {'1': 'email', '3': 3, '4': 1, '5': 9, '10': 'email'},
    {
      '1': 'subscription_expires_at',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'subscriptionExpiresAt'
    },
    {
      '1': 'dedicated_ip_status',
      '3': 5,
      '4': 1,
      '5': 3,
      '10': 'dedicatedIpStatus'
    },
    {
      '1': 'last_dedicated_ip_expires_at',
      '3': 6,
      '4': 1,
      '5': 9,
      '10': 'lastDedicatedIpExpiresAt'
    },
    {
      '1': 'dedicated_ip_services',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.pb.DedidcatedIPService',
      '10': 'dedicatedIpServices'
    },
    {
      '1': 'mfa_status',
      '3': 8,
      '4': 1,
      '5': 14,
      '6': '.pb.TriState',
      '10': 'mfaStatus'
    },
    {'1': 'created_on', '3': 9, '4': 1, '5': 9, '10': 'createdOn'},
  ],
};

/// Descriptor for `AccountResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List accountResponseDescriptor = $convert.base64Decode(
    'Cg9BY2NvdW50UmVzcG9uc2USEgoEdHlwZRgBIAEoA1IEdHlwZRIaCgh1c2VybmFtZRgCIAEoCV'
    'IIdXNlcm5hbWUSFAoFZW1haWwYAyABKAlSBWVtYWlsEjYKF3N1YnNjcmlwdGlvbl9leHBpcmVz'
    'X2F0GAQgASgJUhVzdWJzY3JpcHRpb25FeHBpcmVzQXQSLgoTZGVkaWNhdGVkX2lwX3N0YXR1cx'
    'gFIAEoA1IRZGVkaWNhdGVkSXBTdGF0dXMSPgocbGFzdF9kZWRpY2F0ZWRfaXBfZXhwaXJlc19h'
    'dBgGIAEoCVIYbGFzdERlZGljYXRlZElwRXhwaXJlc0F0EksKFWRlZGljYXRlZF9pcF9zZXJ2aW'
    'NlcxgHIAMoCzIXLnBiLkRlZGlkY2F0ZWRJUFNlcnZpY2VSE2RlZGljYXRlZElwU2VydmljZXMS'
    'KwoKbWZhX3N0YXR1cxgIIAEoDjIMLnBiLlRyaVN0YXRlUgltZmFTdGF0dXMSHQoKY3JlYXRlZF'
    '9vbhgJIAEoCVIJY3JlYXRlZE9u');

@$core.Deprecated('Use accountRequestDescriptor instead')
const AccountRequest$json = {
  '1': 'AccountRequest',
  '2': [
    {'1': 'full', '3': 1, '4': 1, '5': 8, '10': 'full'},
  ],
};

/// Descriptor for `AccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List accountRequestDescriptor =
    $convert.base64Decode('Cg5BY2NvdW50UmVxdWVzdBISCgRmdWxsGAEgASgIUgRmdWxs');
