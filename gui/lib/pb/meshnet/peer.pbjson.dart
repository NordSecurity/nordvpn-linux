// This is a generated file - do not edit.
//
// Generated from peer.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use peerStatusDescriptor instead')
const PeerStatus$json = {
  '1': 'PeerStatus',
  '2': [
    {'1': 'DISCONNECTED', '2': 0},
    {'1': 'CONNECTED', '2': 1},
  ],
};

/// Descriptor for `PeerStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List peerStatusDescriptor = $convert.base64Decode(
    'CgpQZWVyU3RhdHVzEhAKDERJU0NPTk5FQ1RFRBAAEg0KCUNPTk5FQ1RFRBAB');

@$core.Deprecated('Use updatePeerErrorCodeDescriptor instead')
const UpdatePeerErrorCode$json = {
  '1': 'UpdatePeerErrorCode',
  '2': [
    {'1': 'PEER_NOT_FOUND', '2': 0},
  ],
};

/// Descriptor for `UpdatePeerErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List updatePeerErrorCodeDescriptor = $convert
    .base64Decode('ChNVcGRhdGVQZWVyRXJyb3JDb2RlEhIKDlBFRVJfTk9UX0ZPVU5EEAA=');

@$core.Deprecated('Use changeNicknameErrorCodeDescriptor instead')
const ChangeNicknameErrorCode$json = {
  '1': 'ChangeNicknameErrorCode',
  '2': [
    {'1': 'SAME_NICKNAME', '2': 0},
    {'1': 'NICKNAME_ALREADY_EMPTY', '2': 1},
    {'1': 'DOMAIN_NAME_EXISTS', '2': 2},
    {'1': 'RATE_LIMIT_REACH', '2': 3},
    {'1': 'NICKNAME_TOO_LONG', '2': 4},
    {'1': 'DUPLICATE_NICKNAME', '2': 5},
    {'1': 'CONTAINS_FORBIDDEN_WORD', '2': 6},
    {'1': 'SUFFIX_OR_PREFIX_ARE_INVALID', '2': 7},
    {'1': 'NICKNAME_HAS_DOUBLE_HYPHENS', '2': 8},
    {'1': 'INVALID_CHARS', '2': 9},
  ],
};

/// Descriptor for `ChangeNicknameErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List changeNicknameErrorCodeDescriptor = $convert.base64Decode(
    'ChdDaGFuZ2VOaWNrbmFtZUVycm9yQ29kZRIRCg1TQU1FX05JQ0tOQU1FEAASGgoWTklDS05BTU'
    'VfQUxSRUFEWV9FTVBUWRABEhYKEkRPTUFJTl9OQU1FX0VYSVNUUxACEhQKEFJBVEVfTElNSVRf'
    'UkVBQ0gQAxIVChFOSUNLTkFNRV9UT09fTE9ORxAEEhYKEkRVUExJQ0FURV9OSUNLTkFNRRAFEh'
    'sKF0NPTlRBSU5TX0ZPUkJJRERFTl9XT1JEEAYSIAocU1VGRklYX09SX1BSRUZJWF9BUkVfSU5W'
    'QUxJRBAHEh8KG05JQ0tOQU1FX0hBU19ET1VCTEVfSFlQSEVOUxAIEhEKDUlOVkFMSURfQ0hBUl'
    'MQCQ==');

@$core.Deprecated('Use allowRoutingErrorCodeDescriptor instead')
const AllowRoutingErrorCode$json = {
  '1': 'AllowRoutingErrorCode',
  '2': [
    {'1': 'ROUTING_ALREADY_ALLOWED', '2': 0},
  ],
};

/// Descriptor for `AllowRoutingErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List allowRoutingErrorCodeDescriptor = $convert.base64Decode(
    'ChVBbGxvd1JvdXRpbmdFcnJvckNvZGUSGwoXUk9VVElOR19BTFJFQURZX0FMTE9XRUQQAA==');

@$core.Deprecated('Use denyRoutingErrorCodeDescriptor instead')
const DenyRoutingErrorCode$json = {
  '1': 'DenyRoutingErrorCode',
  '2': [
    {'1': 'ROUTING_ALREADY_DENIED', '2': 0},
  ],
};

/// Descriptor for `DenyRoutingErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List denyRoutingErrorCodeDescriptor =
    $convert.base64Decode(
        'ChREZW55Um91dGluZ0Vycm9yQ29kZRIaChZST1VUSU5HX0FMUkVBRFlfREVOSUVEEAA=');

@$core.Deprecated('Use allowIncomingErrorCodeDescriptor instead')
const AllowIncomingErrorCode$json = {
  '1': 'AllowIncomingErrorCode',
  '2': [
    {'1': 'INCOMING_ALREADY_ALLOWED', '2': 0},
  ],
};

/// Descriptor for `AllowIncomingErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List allowIncomingErrorCodeDescriptor =
    $convert.base64Decode(
        'ChZBbGxvd0luY29taW5nRXJyb3JDb2RlEhwKGElOQ09NSU5HX0FMUkVBRFlfQUxMT1dFRBAA');

@$core.Deprecated('Use denyIncomingErrorCodeDescriptor instead')
const DenyIncomingErrorCode$json = {
  '1': 'DenyIncomingErrorCode',
  '2': [
    {'1': 'INCOMING_ALREADY_DENIED', '2': 0},
  ],
};

/// Descriptor for `DenyIncomingErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List denyIncomingErrorCodeDescriptor = $convert.base64Decode(
    'ChVEZW55SW5jb21pbmdFcnJvckNvZGUSGwoXSU5DT01JTkdfQUxSRUFEWV9ERU5JRUQQAA==');

@$core.Deprecated('Use allowLocalNetworkErrorCodeDescriptor instead')
const AllowLocalNetworkErrorCode$json = {
  '1': 'AllowLocalNetworkErrorCode',
  '2': [
    {'1': 'LOCAL_NETWORK_ALREADY_ALLOWED', '2': 0},
  ],
};

/// Descriptor for `AllowLocalNetworkErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List allowLocalNetworkErrorCodeDescriptor =
    $convert.base64Decode(
        'ChpBbGxvd0xvY2FsTmV0d29ya0Vycm9yQ29kZRIhCh1MT0NBTF9ORVRXT1JLX0FMUkVBRFlfQU'
        'xMT1dFRBAA');

@$core.Deprecated('Use denyLocalNetworkErrorCodeDescriptor instead')
const DenyLocalNetworkErrorCode$json = {
  '1': 'DenyLocalNetworkErrorCode',
  '2': [
    {'1': 'LOCAL_NETWORK_ALREADY_DENIED', '2': 0},
  ],
};

/// Descriptor for `DenyLocalNetworkErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List denyLocalNetworkErrorCodeDescriptor =
    $convert.base64Decode(
        'ChlEZW55TG9jYWxOZXR3b3JrRXJyb3JDb2RlEiAKHExPQ0FMX05FVFdPUktfQUxSRUFEWV9ERU'
        '5JRUQQAA==');

@$core.Deprecated('Use allowFileshareErrorCodeDescriptor instead')
const AllowFileshareErrorCode$json = {
  '1': 'AllowFileshareErrorCode',
  '2': [
    {'1': 'SEND_ALREADY_ALLOWED', '2': 0},
  ],
};

/// Descriptor for `AllowFileshareErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List allowFileshareErrorCodeDescriptor =
    $convert.base64Decode(
        'ChdBbGxvd0ZpbGVzaGFyZUVycm9yQ29kZRIYChRTRU5EX0FMUkVBRFlfQUxMT1dFRBAA');

@$core.Deprecated('Use denyFileshareErrorCodeDescriptor instead')
const DenyFileshareErrorCode$json = {
  '1': 'DenyFileshareErrorCode',
  '2': [
    {'1': 'SEND_ALREADY_DENIED', '2': 0},
  ],
};

/// Descriptor for `DenyFileshareErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List denyFileshareErrorCodeDescriptor =
    $convert.base64Decode(
        'ChZEZW55RmlsZXNoYXJlRXJyb3JDb2RlEhcKE1NFTkRfQUxSRUFEWV9ERU5JRUQQAA==');

@$core.Deprecated('Use enableAutomaticFileshareErrorCodeDescriptor instead')
const EnableAutomaticFileshareErrorCode$json = {
  '1': 'EnableAutomaticFileshareErrorCode',
  '2': [
    {'1': 'AUTOMATIC_FILESHARE_ALREADY_ENABLED', '2': 0},
  ],
};

/// Descriptor for `EnableAutomaticFileshareErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List enableAutomaticFileshareErrorCodeDescriptor =
    $convert.base64Decode(
        'CiFFbmFibGVBdXRvbWF0aWNGaWxlc2hhcmVFcnJvckNvZGUSJwojQVVUT01BVElDX0ZJTEVTSE'
        'FSRV9BTFJFQURZX0VOQUJMRUQQAA==');

@$core.Deprecated('Use disableAutomaticFileshareErrorCodeDescriptor instead')
const DisableAutomaticFileshareErrorCode$json = {
  '1': 'DisableAutomaticFileshareErrorCode',
  '2': [
    {'1': 'AUTOMATIC_FILESHARE_ALREADY_DISABLED', '2': 0},
  ],
};

/// Descriptor for `DisableAutomaticFileshareErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List disableAutomaticFileshareErrorCodeDescriptor =
    $convert.base64Decode(
        'CiJEaXNhYmxlQXV0b21hdGljRmlsZXNoYXJlRXJyb3JDb2RlEigKJEFVVE9NQVRJQ19GSUxFU0'
        'hBUkVfQUxSRUFEWV9ESVNBQkxFRBAA');

@$core.Deprecated('Use connectErrorCodeDescriptor instead')
const ConnectErrorCode$json = {
  '1': 'ConnectErrorCode',
  '2': [
    {'1': 'PEER_DOES_NOT_ALLOW_ROUTING', '2': 0},
    {'1': 'ALREADY_CONNECTED', '2': 1},
    {'1': 'CONNECT_FAILED', '2': 2},
    {'1': 'PEER_NO_IP', '2': 3},
    {'1': 'ALREADY_CONNECTING', '2': 4},
    {'1': 'CANCELED', '2': 5},
  ],
};

/// Descriptor for `ConnectErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List connectErrorCodeDescriptor = $convert.base64Decode(
    'ChBDb25uZWN0RXJyb3JDb2RlEh8KG1BFRVJfRE9FU19OT1RfQUxMT1dfUk9VVElORxAAEhUKEU'
    'FMUkVBRFlfQ09OTkVDVEVEEAESEgoOQ09OTkVDVF9GQUlMRUQQAhIOCgpQRUVSX05PX0lQEAMS'
    'FgoSQUxSRUFEWV9DT05ORUNUSU5HEAQSDAoIQ0FOQ0VMRUQQBQ==');

@$core.Deprecated('Use getPeersResponseDescriptor instead')
const GetPeersResponse$json = {
  '1': 'GetPeersResponse',
  '2': [
    {
      '1': 'peers',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.meshpb.PeerList',
      '9': 0,
      '10': 'peers'
    },
    {
      '1': 'error',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.meshpb.Error',
      '9': 0,
      '10': 'error'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `GetPeersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getPeersResponseDescriptor = $convert.base64Decode(
    'ChBHZXRQZWVyc1Jlc3BvbnNlEigKBXBlZXJzGAEgASgLMhAubWVzaHBiLlBlZXJMaXN0SABSBX'
    'BlZXJzEiUKBWVycm9yGAQgASgLMg0ubWVzaHBiLkVycm9ySABSBWVycm9yQgoKCHJlc3BvbnNl');

@$core.Deprecated('Use peerListDescriptor instead')
const PeerList$json = {
  '1': 'PeerList',
  '2': [
    {'1': 'self', '3': 1, '4': 1, '5': 11, '6': '.meshpb.Peer', '10': 'self'},
    {'1': 'local', '3': 2, '4': 3, '5': 11, '6': '.meshpb.Peer', '10': 'local'},
    {
      '1': 'external',
      '3': 3,
      '4': 3,
      '5': 11,
      '6': '.meshpb.Peer',
      '10': 'external'
    },
  ],
};

/// Descriptor for `PeerList`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List peerListDescriptor = $convert.base64Decode(
    'CghQZWVyTGlzdBIgCgRzZWxmGAEgASgLMgwubWVzaHBiLlBlZXJSBHNlbGYSIgoFbG9jYWwYAi'
    'ADKAsyDC5tZXNocGIuUGVlclIFbG9jYWwSKAoIZXh0ZXJuYWwYAyADKAsyDC5tZXNocGIuUGVl'
    'clIIZXh0ZXJuYWw=');

@$core.Deprecated('Use peerDescriptor instead')
const Peer$json = {
  '1': 'Peer',
  '2': [
    {'1': 'identifier', '3': 1, '4': 1, '5': 9, '10': 'identifier'},
    {'1': 'pubkey', '3': 2, '4': 1, '5': 9, '10': 'pubkey'},
    {'1': 'ip', '3': 3, '4': 1, '5': 9, '10': 'ip'},
    {'1': 'endpoints', '3': 4, '4': 3, '5': 9, '10': 'endpoints'},
    {'1': 'os', '3': 5, '4': 1, '5': 9, '10': 'os'},
    {'1': 'os_version', '3': 6, '4': 1, '5': 9, '10': 'osVersion'},
    {'1': 'hostname', '3': 7, '4': 1, '5': 9, '10': 'hostname'},
    {'1': 'distro', '3': 8, '4': 1, '5': 9, '10': 'distro'},
    {'1': 'email', '3': 9, '4': 1, '5': 9, '10': 'email'},
    {
      '1': 'is_inbound_allowed',
      '3': 10,
      '4': 1,
      '5': 8,
      '10': 'isInboundAllowed'
    },
    {'1': 'is_routable', '3': 11, '4': 1, '5': 8, '10': 'isRoutable'},
    {
      '1': 'is_local_network_allowed',
      '3': 15,
      '4': 1,
      '5': 8,
      '10': 'isLocalNetworkAllowed'
    },
    {
      '1': 'is_fileshare_allowed',
      '3': 17,
      '4': 1,
      '5': 8,
      '10': 'isFileshareAllowed'
    },
    {
      '1': 'do_i_allow_inbound',
      '3': 12,
      '4': 1,
      '5': 8,
      '10': 'doIAllowInbound'
    },
    {
      '1': 'do_i_allow_routing',
      '3': 13,
      '4': 1,
      '5': 8,
      '10': 'doIAllowRouting'
    },
    {
      '1': 'do_i_allow_local_network',
      '3': 16,
      '4': 1,
      '5': 8,
      '10': 'doIAllowLocalNetwork'
    },
    {
      '1': 'do_i_allow_fileshare',
      '3': 18,
      '4': 1,
      '5': 8,
      '10': 'doIAllowFileshare'
    },
    {
      '1': 'always_accept_files',
      '3': 19,
      '4': 1,
      '5': 8,
      '10': 'alwaysAcceptFiles'
    },
    {
      '1': 'status',
      '3': 14,
      '4': 1,
      '5': 14,
      '6': '.meshpb.PeerStatus',
      '10': 'status'
    },
    {'1': 'nickname', '3': 20, '4': 1, '5': 9, '10': 'nickname'},
  ],
};

/// Descriptor for `Peer`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List peerDescriptor = $convert.base64Decode(
    'CgRQZWVyEh4KCmlkZW50aWZpZXIYASABKAlSCmlkZW50aWZpZXISFgoGcHVia2V5GAIgASgJUg'
    'ZwdWJrZXkSDgoCaXAYAyABKAlSAmlwEhwKCWVuZHBvaW50cxgEIAMoCVIJZW5kcG9pbnRzEg4K'
    'Am9zGAUgASgJUgJvcxIdCgpvc192ZXJzaW9uGAYgASgJUglvc1ZlcnNpb24SGgoIaG9zdG5hbW'
    'UYByABKAlSCGhvc3RuYW1lEhYKBmRpc3RybxgIIAEoCVIGZGlzdHJvEhQKBWVtYWlsGAkgASgJ'
    'UgVlbWFpbBIsChJpc19pbmJvdW5kX2FsbG93ZWQYCiABKAhSEGlzSW5ib3VuZEFsbG93ZWQSHw'
    'oLaXNfcm91dGFibGUYCyABKAhSCmlzUm91dGFibGUSNwoYaXNfbG9jYWxfbmV0d29ya19hbGxv'
    'd2VkGA8gASgIUhVpc0xvY2FsTmV0d29ya0FsbG93ZWQSMAoUaXNfZmlsZXNoYXJlX2FsbG93ZW'
    'QYESABKAhSEmlzRmlsZXNoYXJlQWxsb3dlZBIrChJkb19pX2FsbG93X2luYm91bmQYDCABKAhS'
    'D2RvSUFsbG93SW5ib3VuZBIrChJkb19pX2FsbG93X3JvdXRpbmcYDSABKAhSD2RvSUFsbG93Um'
    '91dGluZxI2Chhkb19pX2FsbG93X2xvY2FsX25ldHdvcmsYECABKAhSFGRvSUFsbG93TG9jYWxO'
    'ZXR3b3JrEi8KFGRvX2lfYWxsb3dfZmlsZXNoYXJlGBIgASgIUhFkb0lBbGxvd0ZpbGVzaGFyZR'
    'IuChNhbHdheXNfYWNjZXB0X2ZpbGVzGBMgASgIUhFhbHdheXNBY2NlcHRGaWxlcxIqCgZzdGF0'
    'dXMYDiABKA4yEi5tZXNocGIuUGVlclN0YXR1c1IGc3RhdHVzEhoKCG5pY2tuYW1lGBQgASgJUg'
    'huaWNrbmFtZQ==');

@$core.Deprecated('Use updatePeerRequestDescriptor instead')
const UpdatePeerRequest$json = {
  '1': 'UpdatePeerRequest',
  '2': [
    {'1': 'identifier', '3': 1, '4': 1, '5': 9, '10': 'identifier'},
  ],
};

/// Descriptor for `UpdatePeerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updatePeerRequestDescriptor = $convert.base64Decode(
    'ChFVcGRhdGVQZWVyUmVxdWVzdBIeCgppZGVudGlmaWVyGAEgASgJUgppZGVudGlmaWVy');

@$core.Deprecated('Use errorDescriptor instead')
const Error$json = {
  '1': 'Error',
  '2': [
    {
      '1': 'service_error_code',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.meshpb.ServiceErrorCode',
      '9': 0,
      '10': 'serviceErrorCode'
    },
    {
      '1': 'meshnet_error_code',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.meshpb.MeshnetErrorCode',
      '9': 0,
      '10': 'meshnetErrorCode'
    },
  ],
  '8': [
    {'1': 'error'},
  ],
};

/// Descriptor for `Error`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List errorDescriptor = $convert.base64Decode(
    'CgVFcnJvchJIChJzZXJ2aWNlX2Vycm9yX2NvZGUYASABKA4yGC5tZXNocGIuU2VydmljZUVycm'
    '9yQ29kZUgAUhBzZXJ2aWNlRXJyb3JDb2RlEkgKEm1lc2huZXRfZXJyb3JfY29kZRgCIAEoDjIY'
    'Lm1lc2hwYi5NZXNobmV0RXJyb3JDb2RlSABSEG1lc2huZXRFcnJvckNvZGVCBwoFZXJyb3I=');

@$core.Deprecated('Use updatePeerErrorDescriptor instead')
const UpdatePeerError$json = {
  '1': 'UpdatePeerError',
  '2': [
    {
      '1': 'general_error',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.meshpb.Error',
      '9': 0,
      '10': 'generalError'
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
  ],
  '8': [
    {'1': 'error'},
  ],
};

/// Descriptor for `UpdatePeerError`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updatePeerErrorDescriptor = $convert.base64Decode(
    'Cg9VcGRhdGVQZWVyRXJyb3ISNAoNZ2VuZXJhbF9lcnJvchgBIAEoCzINLm1lc2hwYi5FcnJvck'
    'gAUgxnZW5lcmFsRXJyb3ISUgoWdXBkYXRlX3BlZXJfZXJyb3JfY29kZRgCIAEoDjIbLm1lc2hw'
    'Yi5VcGRhdGVQZWVyRXJyb3JDb2RlSABSE3VwZGF0ZVBlZXJFcnJvckNvZGVCBwoFZXJyb3I=');

@$core.Deprecated('Use removePeerResponseDescriptor instead')
const RemovePeerResponse$json = {
  '1': 'RemovePeerResponse',
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
      '1': 'update_peer_error',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.meshpb.UpdatePeerError',
      '9': 0,
      '10': 'updatePeerError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `RemovePeerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List removePeerResponseDescriptor = $convert.base64Decode(
    'ChJSZW1vdmVQZWVyUmVzcG9uc2USJQoFZW1wdHkYASABKAsyDS5tZXNocGIuRW1wdHlIAFIFZW'
    '1wdHkSRQoRdXBkYXRlX3BlZXJfZXJyb3IYBSABKAsyFy5tZXNocGIuVXBkYXRlUGVlckVycm9y'
    'SABSD3VwZGF0ZVBlZXJFcnJvckIKCghyZXNwb25zZQ==');

@$core.Deprecated('Use changePeerNicknameRequestDescriptor instead')
const ChangePeerNicknameRequest$json = {
  '1': 'ChangePeerNicknameRequest',
  '2': [
    {'1': 'identifier', '3': 1, '4': 1, '5': 9, '10': 'identifier'},
    {'1': 'nickname', '3': 2, '4': 1, '5': 9, '10': 'nickname'},
  ],
};

/// Descriptor for `ChangePeerNicknameRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List changePeerNicknameRequestDescriptor =
    $convert.base64Decode(
        'ChlDaGFuZ2VQZWVyTmlja25hbWVSZXF1ZXN0Eh4KCmlkZW50aWZpZXIYASABKAlSCmlkZW50aW'
        'ZpZXISGgoIbmlja25hbWUYAiABKAlSCG5pY2tuYW1l');

@$core.Deprecated('Use changeMachineNicknameRequestDescriptor instead')
const ChangeMachineNicknameRequest$json = {
  '1': 'ChangeMachineNicknameRequest',
  '2': [
    {'1': 'nickname', '3': 1, '4': 1, '5': 9, '10': 'nickname'},
  ],
};

/// Descriptor for `ChangeMachineNicknameRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List changeMachineNicknameRequestDescriptor =
    $convert.base64Decode(
        'ChxDaGFuZ2VNYWNoaW5lTmlja25hbWVSZXF1ZXN0EhoKCG5pY2tuYW1lGAEgASgJUghuaWNrbm'
        'FtZQ==');

@$core.Deprecated('Use changeNicknameResponseDescriptor instead')
const ChangeNicknameResponse$json = {
  '1': 'ChangeNicknameResponse',
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
      '1': 'change_nickname_error_code',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.meshpb.ChangeNicknameErrorCode',
      '9': 0,
      '10': 'changeNicknameErrorCode'
    },
    {
      '1': 'update_peer_error',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.meshpb.UpdatePeerError',
      '9': 0,
      '10': 'updatePeerError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `ChangeNicknameResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List changeNicknameResponseDescriptor = $convert.base64Decode(
    'ChZDaGFuZ2VOaWNrbmFtZVJlc3BvbnNlEiUKBWVtcHR5GAEgASgLMg0ubWVzaHBiLkVtcHR5SA'
    'BSBWVtcHR5El4KGmNoYW5nZV9uaWNrbmFtZV9lcnJvcl9jb2RlGAUgASgOMh8ubWVzaHBiLkNo'
    'YW5nZU5pY2tuYW1lRXJyb3JDb2RlSABSF2NoYW5nZU5pY2tuYW1lRXJyb3JDb2RlEkUKEXVwZG'
    'F0ZV9wZWVyX2Vycm9yGAYgASgLMhcubWVzaHBiLlVwZGF0ZVBlZXJFcnJvckgAUg91cGRhdGVQ'
    'ZWVyRXJyb3JCCgoIcmVzcG9uc2U=');

@$core.Deprecated('Use allowRoutingResponseDescriptor instead')
const AllowRoutingResponse$json = {
  '1': 'AllowRoutingResponse',
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
      '1': 'allow_routing_error_code',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.meshpb.AllowRoutingErrorCode',
      '9': 0,
      '10': 'allowRoutingErrorCode'
    },
    {
      '1': 'update_peer_error',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.meshpb.UpdatePeerError',
      '9': 0,
      '10': 'updatePeerError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `AllowRoutingResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List allowRoutingResponseDescriptor = $convert.base64Decode(
    'ChRBbGxvd1JvdXRpbmdSZXNwb25zZRIlCgVlbXB0eRgBIAEoCzINLm1lc2hwYi5FbXB0eUgAUg'
    'VlbXB0eRJYChhhbGxvd19yb3V0aW5nX2Vycm9yX2NvZGUYAyABKA4yHS5tZXNocGIuQWxsb3dS'
    'b3V0aW5nRXJyb3JDb2RlSABSFWFsbG93Um91dGluZ0Vycm9yQ29kZRJFChF1cGRhdGVfcGVlcl'
    '9lcnJvchgGIAEoCzIXLm1lc2hwYi5VcGRhdGVQZWVyRXJyb3JIAFIPdXBkYXRlUGVlckVycm9y'
    'QgoKCHJlc3BvbnNl');

@$core.Deprecated('Use denyRoutingResponseDescriptor instead')
const DenyRoutingResponse$json = {
  '1': 'DenyRoutingResponse',
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
      '1': 'deny_routing_error_code',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.meshpb.DenyRoutingErrorCode',
      '9': 0,
      '10': 'denyRoutingErrorCode'
    },
    {
      '1': 'update_peer_error',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.meshpb.UpdatePeerError',
      '9': 0,
      '10': 'updatePeerError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `DenyRoutingResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List denyRoutingResponseDescriptor = $convert.base64Decode(
    'ChNEZW55Um91dGluZ1Jlc3BvbnNlEiUKBWVtcHR5GAEgASgLMg0ubWVzaHBiLkVtcHR5SABSBW'
    'VtcHR5ElUKF2Rlbnlfcm91dGluZ19lcnJvcl9jb2RlGAMgASgOMhwubWVzaHBiLkRlbnlSb3V0'
    'aW5nRXJyb3JDb2RlSABSFGRlbnlSb3V0aW5nRXJyb3JDb2RlEkUKEXVwZGF0ZV9wZWVyX2Vycm'
    '9yGAYgASgLMhcubWVzaHBiLlVwZGF0ZVBlZXJFcnJvckgAUg91cGRhdGVQZWVyRXJyb3JCCgoI'
    'cmVzcG9uc2U=');

@$core.Deprecated('Use allowIncomingResponseDescriptor instead')
const AllowIncomingResponse$json = {
  '1': 'AllowIncomingResponse',
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
      '1': 'allow_incoming_error_code',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.meshpb.AllowIncomingErrorCode',
      '9': 0,
      '10': 'allowIncomingErrorCode'
    },
    {
      '1': 'update_peer_error',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.meshpb.UpdatePeerError',
      '9': 0,
      '10': 'updatePeerError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `AllowIncomingResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List allowIncomingResponseDescriptor = $convert.base64Decode(
    'ChVBbGxvd0luY29taW5nUmVzcG9uc2USJQoFZW1wdHkYASABKAsyDS5tZXNocGIuRW1wdHlIAF'
    'IFZW1wdHkSWwoZYWxsb3dfaW5jb21pbmdfZXJyb3JfY29kZRgDIAEoDjIeLm1lc2hwYi5BbGxv'
    'd0luY29taW5nRXJyb3JDb2RlSABSFmFsbG93SW5jb21pbmdFcnJvckNvZGUSRQoRdXBkYXRlX3'
    'BlZXJfZXJyb3IYBiABKAsyFy5tZXNocGIuVXBkYXRlUGVlckVycm9ySABSD3VwZGF0ZVBlZXJF'
    'cnJvckIKCghyZXNwb25zZQ==');

@$core.Deprecated('Use denyIncomingResponseDescriptor instead')
const DenyIncomingResponse$json = {
  '1': 'DenyIncomingResponse',
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
      '1': 'deny_incoming_error_code',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.meshpb.DenyIncomingErrorCode',
      '9': 0,
      '10': 'denyIncomingErrorCode'
    },
    {
      '1': 'update_peer_error',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.meshpb.UpdatePeerError',
      '9': 0,
      '10': 'updatePeerError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `DenyIncomingResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List denyIncomingResponseDescriptor = $convert.base64Decode(
    'ChREZW55SW5jb21pbmdSZXNwb25zZRIlCgVlbXB0eRgBIAEoCzINLm1lc2hwYi5FbXB0eUgAUg'
    'VlbXB0eRJYChhkZW55X2luY29taW5nX2Vycm9yX2NvZGUYAyABKA4yHS5tZXNocGIuRGVueUlu'
    'Y29taW5nRXJyb3JDb2RlSABSFWRlbnlJbmNvbWluZ0Vycm9yQ29kZRJFChF1cGRhdGVfcGVlcl'
    '9lcnJvchgGIAEoCzIXLm1lc2hwYi5VcGRhdGVQZWVyRXJyb3JIAFIPdXBkYXRlUGVlckVycm9y'
    'QgoKCHJlc3BvbnNl');

@$core.Deprecated('Use allowLocalNetworkResponseDescriptor instead')
const AllowLocalNetworkResponse$json = {
  '1': 'AllowLocalNetworkResponse',
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
      '1': 'allow_local_network_error_code',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.meshpb.AllowLocalNetworkErrorCode',
      '9': 0,
      '10': 'allowLocalNetworkErrorCode'
    },
    {
      '1': 'update_peer_error',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.meshpb.UpdatePeerError',
      '9': 0,
      '10': 'updatePeerError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `AllowLocalNetworkResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List allowLocalNetworkResponseDescriptor = $convert.base64Decode(
    'ChlBbGxvd0xvY2FsTmV0d29ya1Jlc3BvbnNlEiUKBWVtcHR5GAEgASgLMg0ubWVzaHBiLkVtcH'
    'R5SABSBWVtcHR5EmgKHmFsbG93X2xvY2FsX25ldHdvcmtfZXJyb3JfY29kZRgDIAEoDjIiLm1l'
    'c2hwYi5BbGxvd0xvY2FsTmV0d29ya0Vycm9yQ29kZUgAUhphbGxvd0xvY2FsTmV0d29ya0Vycm'
    '9yQ29kZRJFChF1cGRhdGVfcGVlcl9lcnJvchgGIAEoCzIXLm1lc2hwYi5VcGRhdGVQZWVyRXJy'
    'b3JIAFIPdXBkYXRlUGVlckVycm9yQgoKCHJlc3BvbnNl');

@$core.Deprecated('Use denyLocalNetworkResponseDescriptor instead')
const DenyLocalNetworkResponse$json = {
  '1': 'DenyLocalNetworkResponse',
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
      '1': 'deny_local_network_error_code',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.meshpb.DenyLocalNetworkErrorCode',
      '9': 0,
      '10': 'denyLocalNetworkErrorCode'
    },
    {
      '1': 'update_peer_error',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.meshpb.UpdatePeerError',
      '9': 0,
      '10': 'updatePeerError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `DenyLocalNetworkResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List denyLocalNetworkResponseDescriptor = $convert.base64Decode(
    'ChhEZW55TG9jYWxOZXR3b3JrUmVzcG9uc2USJQoFZW1wdHkYASABKAsyDS5tZXNocGIuRW1wdH'
    'lIAFIFZW1wdHkSZQodZGVueV9sb2NhbF9uZXR3b3JrX2Vycm9yX2NvZGUYAyABKA4yIS5tZXNo'
    'cGIuRGVueUxvY2FsTmV0d29ya0Vycm9yQ29kZUgAUhlkZW55TG9jYWxOZXR3b3JrRXJyb3JDb2'
    'RlEkUKEXVwZGF0ZV9wZWVyX2Vycm9yGAYgASgLMhcubWVzaHBiLlVwZGF0ZVBlZXJFcnJvckgA'
    'Ug91cGRhdGVQZWVyRXJyb3JCCgoIcmVzcG9uc2U=');

@$core.Deprecated('Use allowFileshareResponseDescriptor instead')
const AllowFileshareResponse$json = {
  '1': 'AllowFileshareResponse',
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
      '1': 'allow_send_error_code',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.meshpb.AllowFileshareErrorCode',
      '9': 0,
      '10': 'allowSendErrorCode'
    },
    {
      '1': 'update_peer_error',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.meshpb.UpdatePeerError',
      '9': 0,
      '10': 'updatePeerError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `AllowFileshareResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List allowFileshareResponseDescriptor = $convert.base64Decode(
    'ChZBbGxvd0ZpbGVzaGFyZVJlc3BvbnNlEiUKBWVtcHR5GAEgASgLMg0ubWVzaHBiLkVtcHR5SA'
    'BSBWVtcHR5ElQKFWFsbG93X3NlbmRfZXJyb3JfY29kZRgDIAEoDjIfLm1lc2hwYi5BbGxvd0Zp'
    'bGVzaGFyZUVycm9yQ29kZUgAUhJhbGxvd1NlbmRFcnJvckNvZGUSRQoRdXBkYXRlX3BlZXJfZX'
    'Jyb3IYBiABKAsyFy5tZXNocGIuVXBkYXRlUGVlckVycm9ySABSD3VwZGF0ZVBlZXJFcnJvckIK'
    'CghyZXNwb25zZQ==');

@$core.Deprecated('Use denyFileshareResponseDescriptor instead')
const DenyFileshareResponse$json = {
  '1': 'DenyFileshareResponse',
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
      '1': 'deny_send_error_code',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.meshpb.DenyFileshareErrorCode',
      '9': 0,
      '10': 'denySendErrorCode'
    },
    {
      '1': 'update_peer_error',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.meshpb.UpdatePeerError',
      '9': 0,
      '10': 'updatePeerError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `DenyFileshareResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List denyFileshareResponseDescriptor = $convert.base64Decode(
    'ChVEZW55RmlsZXNoYXJlUmVzcG9uc2USJQoFZW1wdHkYASABKAsyDS5tZXNocGIuRW1wdHlIAF'
    'IFZW1wdHkSUQoUZGVueV9zZW5kX2Vycm9yX2NvZGUYAyABKA4yHi5tZXNocGIuRGVueUZpbGVz'
    'aGFyZUVycm9yQ29kZUgAUhFkZW55U2VuZEVycm9yQ29kZRJFChF1cGRhdGVfcGVlcl9lcnJvch'
    'gGIAEoCzIXLm1lc2hwYi5VcGRhdGVQZWVyRXJyb3JIAFIPdXBkYXRlUGVlckVycm9yQgoKCHJl'
    'c3BvbnNl');

@$core.Deprecated('Use enableAutomaticFileshareResponseDescriptor instead')
const EnableAutomaticFileshareResponse$json = {
  '1': 'EnableAutomaticFileshareResponse',
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
      '1': 'enable_automatic_fileshare_error_code',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.meshpb.EnableAutomaticFileshareErrorCode',
      '9': 0,
      '10': 'enableAutomaticFileshareErrorCode'
    },
    {
      '1': 'update_peer_error',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.meshpb.UpdatePeerError',
      '9': 0,
      '10': 'updatePeerError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `EnableAutomaticFileshareResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List enableAutomaticFileshareResponseDescriptor = $convert.base64Decode(
    'CiBFbmFibGVBdXRvbWF0aWNGaWxlc2hhcmVSZXNwb25zZRIlCgVlbXB0eRgBIAEoCzINLm1lc2'
    'hwYi5FbXB0eUgAUgVlbXB0eRJ9CiVlbmFibGVfYXV0b21hdGljX2ZpbGVzaGFyZV9lcnJvcl9j'
    'b2RlGAMgASgOMikubWVzaHBiLkVuYWJsZUF1dG9tYXRpY0ZpbGVzaGFyZUVycm9yQ29kZUgAUi'
    'FlbmFibGVBdXRvbWF0aWNGaWxlc2hhcmVFcnJvckNvZGUSRQoRdXBkYXRlX3BlZXJfZXJyb3IY'
    'BiABKAsyFy5tZXNocGIuVXBkYXRlUGVlckVycm9ySABSD3VwZGF0ZVBlZXJFcnJvckIKCghyZX'
    'Nwb25zZQ==');

@$core.Deprecated('Use disableAutomaticFileshareResponseDescriptor instead')
const DisableAutomaticFileshareResponse$json = {
  '1': 'DisableAutomaticFileshareResponse',
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
      '1': 'disable_automatic_fileshare_error_code',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.meshpb.DisableAutomaticFileshareErrorCode',
      '9': 0,
      '10': 'disableAutomaticFileshareErrorCode'
    },
    {
      '1': 'update_peer_error',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.meshpb.UpdatePeerError',
      '9': 0,
      '10': 'updatePeerError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `DisableAutomaticFileshareResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List disableAutomaticFileshareResponseDescriptor = $convert.base64Decode(
    'CiFEaXNhYmxlQXV0b21hdGljRmlsZXNoYXJlUmVzcG9uc2USJQoFZW1wdHkYASABKAsyDS5tZX'
    'NocGIuRW1wdHlIAFIFZW1wdHkSgAEKJmRpc2FibGVfYXV0b21hdGljX2ZpbGVzaGFyZV9lcnJv'
    'cl9jb2RlGAMgASgOMioubWVzaHBiLkRpc2FibGVBdXRvbWF0aWNGaWxlc2hhcmVFcnJvckNvZG'
    'VIAFIiZGlzYWJsZUF1dG9tYXRpY0ZpbGVzaGFyZUVycm9yQ29kZRJFChF1cGRhdGVfcGVlcl9l'
    'cnJvchgGIAEoCzIXLm1lc2hwYi5VcGRhdGVQZWVyRXJyb3JIAFIPdXBkYXRlUGVlckVycm9yQg'
    'oKCHJlc3BvbnNl');

@$core.Deprecated('Use connectResponseDescriptor instead')
const ConnectResponse$json = {
  '1': 'ConnectResponse',
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
      '1': 'connect_error_code',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.meshpb.ConnectErrorCode',
      '9': 0,
      '10': 'connectErrorCode'
    },
    {
      '1': 'update_peer_error',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.meshpb.UpdatePeerError',
      '9': 0,
      '10': 'updatePeerError'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `ConnectResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List connectResponseDescriptor = $convert.base64Decode(
    'Cg9Db25uZWN0UmVzcG9uc2USJQoFZW1wdHkYASABKAsyDS5tZXNocGIuRW1wdHlIAFIFZW1wdH'
    'kSSAoSY29ubmVjdF9lcnJvcl9jb2RlGAMgASgOMhgubWVzaHBiLkNvbm5lY3RFcnJvckNvZGVI'
    'AFIQY29ubmVjdEVycm9yQ29kZRJFChF1cGRhdGVfcGVlcl9lcnJvchgGIAEoCzIXLm1lc2hwYi'
    '5VcGRhdGVQZWVyRXJyb3JIAFIPdXBkYXRlUGVlckVycm9yQgoKCHJlc3BvbnNl');

@$core.Deprecated('Use privateKeyResponseDescriptor instead')
const PrivateKeyResponse$json = {
  '1': 'PrivateKeyResponse',
  '2': [
    {'1': 'private_key', '3': 1, '4': 1, '5': 9, '9': 0, '10': 'privateKey'},
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

/// Descriptor for `PrivateKeyResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List privateKeyResponseDescriptor = $convert.base64Decode(
    'ChJQcml2YXRlS2V5UmVzcG9uc2USIQoLcHJpdmF0ZV9rZXkYASABKAlIAFIKcHJpdmF0ZUtleR'
    'JIChJzZXJ2aWNlX2Vycm9yX2NvZGUYAiABKA4yGC5tZXNocGIuU2VydmljZUVycm9yQ29kZUgA'
    'UhBzZXJ2aWNlRXJyb3JDb2RlEkgKEm1lc2huZXRfZXJyb3JfY29kZRgDIAEoDjIYLm1lc2hwYi'
    '5NZXNobmV0RXJyb3JDb2RlSABSEG1lc2huZXRFcnJvckNvZGVCCgoIcmVzcG9uc2U=');
