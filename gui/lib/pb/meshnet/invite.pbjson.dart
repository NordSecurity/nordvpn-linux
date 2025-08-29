// This is a generated file - do not edit.
//
// Generated from invite.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use respondToInviteErrorCodeDescriptor instead')
const RespondToInviteErrorCode$json = {
  '1': 'RespondToInviteErrorCode',
  '2': [
    {'1': 'UNKNOWN', '2': 0},
    {'1': 'NO_SUCH_INVITATION', '2': 1},
    {'1': 'DEVICE_COUNT', '2': 2},
  ],
};

/// Descriptor for `RespondToInviteErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List respondToInviteErrorCodeDescriptor =
    $convert.base64Decode(
        'ChhSZXNwb25kVG9JbnZpdGVFcnJvckNvZGUSCwoHVU5LTk9XThAAEhYKEk5PX1NVQ0hfSU5WSV'
        'RBVElPThABEhAKDERFVklDRV9DT1VOVBAC');

@$core.Deprecated('Use inviteResponseErrorCodeDescriptor instead')
const InviteResponseErrorCode$json = {
  '1': 'InviteResponseErrorCode',
  '2': [
    {'1': 'ALREADY_EXISTS', '2': 0},
    {'1': 'INVALID_EMAIL', '2': 1},
    {'1': 'SAME_ACCOUNT_EMAIL', '2': 2},
    {'1': 'LIMIT_REACHED', '2': 3},
    {'1': 'PEER_COUNT', '2': 4},
  ],
};

/// Descriptor for `InviteResponseErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List inviteResponseErrorCodeDescriptor = $convert.base64Decode(
    'ChdJbnZpdGVSZXNwb25zZUVycm9yQ29kZRISCg5BTFJFQURZX0VYSVNUUxAAEhEKDUlOVkFMSU'
    'RfRU1BSUwQARIWChJTQU1FX0FDQ09VTlRfRU1BSUwQAhIRCg1MSU1JVF9SRUFDSEVEEAMSDgoK'
    'UEVFUl9DT1VOVBAE');

@$core.Deprecated('Use getInvitesResponseDescriptor instead')
const GetInvitesResponse$json = {
  '1': 'GetInvitesResponse',
  '2': [
    {
      '1': 'invites',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.meshpb.InvitesList',
      '9': 0,
      '10': 'invites'
    },
    {
      '1': 'service_error_code',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.meshpb.ServiceErrorCode',
      '9': 0,
      '10': 'serviceErrorCode'
    },
    {
      '1': 'meshnet_error_code',
      '3': 3,
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

/// Descriptor for `GetInvitesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getInvitesResponseDescriptor = $convert.base64Decode(
    'ChJHZXRJbnZpdGVzUmVzcG9uc2USLwoHaW52aXRlcxgBIAEoCzITLm1lc2hwYi5JbnZpdGVzTG'
    'lzdEgAUgdpbnZpdGVzEkgKEnNlcnZpY2VfZXJyb3JfY29kZRgCIAEoDjIYLm1lc2hwYi5TZXJ2'
    'aWNlRXJyb3JDb2RlSABSEHNlcnZpY2VFcnJvckNvZGUSSAoSbWVzaG5ldF9lcnJvcl9jb2RlGA'
    'MgASgOMhgubWVzaHBiLk1lc2huZXRFcnJvckNvZGVIAFIQbWVzaG5ldEVycm9yQ29kZUIKCghy'
    'ZXNwb25zZQ==');

@$core.Deprecated('Use invitesListDescriptor instead')
const InvitesList$json = {
  '1': 'InvitesList',
  '2': [
    {'1': 'sent', '3': 1, '4': 3, '5': 11, '6': '.meshpb.Invite', '10': 'sent'},
    {
      '1': 'received',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.meshpb.Invite',
      '10': 'received'
    },
  ],
};

/// Descriptor for `InvitesList`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List invitesListDescriptor = $convert.base64Decode(
    'CgtJbnZpdGVzTGlzdBIiCgRzZW50GAEgAygLMg4ubWVzaHBiLkludml0ZVIEc2VudBIqCghyZW'
    'NlaXZlZBgCIAMoCzIOLm1lc2hwYi5JbnZpdGVSCHJlY2VpdmVk');

@$core.Deprecated('Use inviteDescriptor instead')
const Invite$json = {
  '1': 'Invite',
  '2': [
    {'1': 'email', '3': 1, '4': 1, '5': 9, '10': 'email'},
    {
      '1': 'expires_at',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'expiresAt'
    },
    {'1': 'os', '3': 3, '4': 1, '5': 9, '10': 'os'},
  ],
};

/// Descriptor for `Invite`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List inviteDescriptor = $convert.base64Decode(
    'CgZJbnZpdGUSFAoFZW1haWwYASABKAlSBWVtYWlsEjkKCmV4cGlyZXNfYXQYAiABKAsyGi5nb2'
    '9nbGUucHJvdG9idWYuVGltZXN0YW1wUglleHBpcmVzQXQSDgoCb3MYAyABKAlSAm9z');

@$core.Deprecated('Use inviteRequestDescriptor instead')
const InviteRequest$json = {
  '1': 'InviteRequest',
  '2': [
    {'1': 'email', '3': 1, '4': 1, '5': 9, '10': 'email'},
    {
      '1': 'allowIncomingTraffic',
      '3': 2,
      '4': 1,
      '5': 8,
      '10': 'allowIncomingTraffic'
    },
    {
      '1': 'allowTrafficRouting',
      '3': 3,
      '4': 1,
      '5': 8,
      '10': 'allowTrafficRouting'
    },
    {
      '1': 'allowLocalNetwork',
      '3': 4,
      '4': 1,
      '5': 8,
      '10': 'allowLocalNetwork'
    },
    {'1': 'allowFileshare', '3': 5, '4': 1, '5': 8, '10': 'allowFileshare'},
  ],
};

/// Descriptor for `InviteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List inviteRequestDescriptor = $convert.base64Decode(
    'Cg1JbnZpdGVSZXF1ZXN0EhQKBWVtYWlsGAEgASgJUgVlbWFpbBIyChRhbGxvd0luY29taW5nVH'
    'JhZmZpYxgCIAEoCFIUYWxsb3dJbmNvbWluZ1RyYWZmaWMSMAoTYWxsb3dUcmFmZmljUm91dGlu'
    'ZxgDIAEoCFITYWxsb3dUcmFmZmljUm91dGluZxIsChFhbGxvd0xvY2FsTmV0d29yaxgEIAEoCF'
    'IRYWxsb3dMb2NhbE5ldHdvcmsSJgoOYWxsb3dGaWxlc2hhcmUYBSABKAhSDmFsbG93RmlsZXNo'
    'YXJl');

@$core.Deprecated('Use denyInviteRequestDescriptor instead')
const DenyInviteRequest$json = {
  '1': 'DenyInviteRequest',
  '2': [
    {'1': 'email', '3': 1, '4': 1, '5': 9, '10': 'email'},
  ],
};

/// Descriptor for `DenyInviteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List denyInviteRequestDescriptor = $convert
    .base64Decode('ChFEZW55SW52aXRlUmVxdWVzdBIUCgVlbWFpbBgBIAEoCVIFZW1haWw=');

@$core.Deprecated('Use respondToInviteResponseDescriptor instead')
const RespondToInviteResponse$json = {
  '1': 'RespondToInviteResponse',
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
      '1': 'respond_to_invite_error_code',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.meshpb.RespondToInviteErrorCode',
      '9': 0,
      '10': 'respondToInviteErrorCode'
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

/// Descriptor for `RespondToInviteResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List respondToInviteResponseDescriptor = $convert.base64Decode(
    'ChdSZXNwb25kVG9JbnZpdGVSZXNwb25zZRIlCgVlbXB0eRgBIAEoCzINLm1lc2hwYi5FbXB0eU'
    'gAUgVlbXB0eRJiChxyZXNwb25kX3RvX2ludml0ZV9lcnJvcl9jb2RlGAIgASgOMiAubWVzaHBi'
    'LlJlc3BvbmRUb0ludml0ZUVycm9yQ29kZUgAUhhyZXNwb25kVG9JbnZpdGVFcnJvckNvZGUSSA'
    'oSc2VydmljZV9lcnJvcl9jb2RlGAMgASgOMhgubWVzaHBiLlNlcnZpY2VFcnJvckNvZGVIAFIQ'
    'c2VydmljZUVycm9yQ29kZRJIChJtZXNobmV0X2Vycm9yX2NvZGUYBCABKA4yGC5tZXNocGIuTW'
    'VzaG5ldEVycm9yQ29kZUgAUhBtZXNobmV0RXJyb3JDb2RlQgoKCHJlc3BvbnNl');

@$core.Deprecated('Use inviteResponseDescriptor instead')
const InviteResponse$json = {
  '1': 'InviteResponse',
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
      '1': 'invite_response_error_code',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.meshpb.InviteResponseErrorCode',
      '9': 0,
      '10': 'inviteResponseErrorCode'
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

/// Descriptor for `InviteResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List inviteResponseDescriptor = $convert.base64Decode(
    'Cg5JbnZpdGVSZXNwb25zZRIlCgVlbXB0eRgBIAEoCzINLm1lc2hwYi5FbXB0eUgAUgVlbXB0eR'
    'JeChppbnZpdGVfcmVzcG9uc2VfZXJyb3JfY29kZRgCIAEoDjIfLm1lc2hwYi5JbnZpdGVSZXNw'
    'b25zZUVycm9yQ29kZUgAUhdpbnZpdGVSZXNwb25zZUVycm9yQ29kZRJIChJzZXJ2aWNlX2Vycm'
    '9yX2NvZGUYAyABKA4yGC5tZXNocGIuU2VydmljZUVycm9yQ29kZUgAUhBzZXJ2aWNlRXJyb3JD'
    'b2RlEkgKEm1lc2huZXRfZXJyb3JfY29kZRgEIAEoDjIYLm1lc2hwYi5NZXNobmV0RXJyb3JDb2'
    'RlSABSEG1lc2huZXRFcnJvckNvZGVCCgoIcmVzcG9uc2U=');
