// This is a generated file - do not edit.
//
// Generated from fields.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use displayProtocolDescriptor instead')
const DisplayProtocol$json = {
  '1': 'DisplayProtocol',
  '2': [
    {'1': 'DISPLAY_PROTOCOL_UNSPECIFIED', '2': 0},
    {'1': 'DISPLAY_PROTOCOL_UNKNOWN', '2': 1},
    {'1': 'DISPLAY_PROTOCOL_X11', '2': 2},
    {'1': 'DISPLAY_PROTOCOL_WAYLAND', '2': 3},
  ],
};

/// Descriptor for `DisplayProtocol`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List displayProtocolDescriptor = $convert.base64Decode(
    'Cg9EaXNwbGF5UHJvdG9jb2wSIAocRElTUExBWV9QUk9UT0NPTF9VTlNQRUNJRklFRBAAEhwKGE'
    'RJU1BMQVlfUFJPVE9DT0xfVU5LTk9XThABEhgKFERJU1BMQVlfUFJPVE9DT0xfWDExEAISHAoY'
    'RElTUExBWV9QUk9UT0NPTF9XQVlMQU5EEAM=');

@$core.Deprecated('Use desktopEnvironmentRequestDescriptor instead')
const DesktopEnvironmentRequest$json = {
  '1': 'DesktopEnvironmentRequest',
  '2': [
    {'1': 'desktop_env_name', '3': 1, '4': 1, '5': 9, '10': 'desktopEnvName'},
  ],
};

/// Descriptor for `DesktopEnvironmentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List desktopEnvironmentRequestDescriptor =
    $convert.base64Decode(
        'ChlEZXNrdG9wRW52aXJvbm1lbnRSZXF1ZXN0EigKEGRlc2t0b3BfZW52X25hbWUYASABKAlSDm'
        'Rlc2t0b3BFbnZOYW1l');

@$core.Deprecated('Use displayProtocolRequestDescriptor instead')
const DisplayProtocolRequest$json = {
  '1': 'DisplayProtocolRequest',
  '2': [
    {
      '1': 'protocol',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.telemetry.v1.DisplayProtocol',
      '10': 'protocol'
    },
  ],
};

/// Descriptor for `DisplayProtocolRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List displayProtocolRequestDescriptor =
    $convert.base64Decode(
        'ChZEaXNwbGF5UHJvdG9jb2xSZXF1ZXN0EjkKCHByb3RvY29sGAEgASgOMh0udGVsZW1ldHJ5Ln'
        'YxLkRpc3BsYXlQcm90b2NvbFIIcHJvdG9jb2w=');
