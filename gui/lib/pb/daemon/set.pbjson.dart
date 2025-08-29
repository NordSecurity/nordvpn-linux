// This is a generated file - do not edit.
//
// Generated from set.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use setErrorCodeDescriptor instead')
const SetErrorCode$json = {
  '1': 'SetErrorCode',
  '2': [
    {'1': 'FAILURE', '2': 0},
    {'1': 'CONFIG_ERROR', '2': 1},
    {'1': 'ALREADY_SET', '2': 2},
  ],
};

/// Descriptor for `SetErrorCode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List setErrorCodeDescriptor = $convert.base64Decode(
    'CgxTZXRFcnJvckNvZGUSCwoHRkFJTFVSRRAAEhAKDENPTkZJR19FUlJPUhABEg8KC0FMUkVBRF'
    'lfU0VUEAI=');

@$core.Deprecated('Use setThreatProtectionLiteStatusDescriptor instead')
const SetThreatProtectionLiteStatus$json = {
  '1': 'SetThreatProtectionLiteStatus',
  '2': [
    {'1': 'TPL_CONFIGURED', '2': 0},
    {'1': 'TPL_CONFIGURED_DNS_RESET', '2': 1},
  ],
};

/// Descriptor for `SetThreatProtectionLiteStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List setThreatProtectionLiteStatusDescriptor =
    $convert.base64Decode(
        'Ch1TZXRUaHJlYXRQcm90ZWN0aW9uTGl0ZVN0YXR1cxISCg5UUExfQ09ORklHVVJFRBAAEhwKGF'
        'RQTF9DT05GSUdVUkVEX0ROU19SRVNFVBAB');

@$core.Deprecated('Use setDNSStatusDescriptor instead')
const SetDNSStatus$json = {
  '1': 'SetDNSStatus',
  '2': [
    {'1': 'DNS_CONFIGURED', '2': 0},
    {'1': 'DNS_CONFIGURED_TPL_RESET', '2': 1},
    {'1': 'INVALID_DNS_ADDRESS', '2': 2},
    {'1': 'TOO_MANY_VALUES', '2': 3},
  ],
};

/// Descriptor for `SetDNSStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List setDNSStatusDescriptor = $convert.base64Decode(
    'CgxTZXRETlNTdGF0dXMSEgoORE5TX0NPTkZJR1VSRUQQABIcChhETlNfQ09ORklHVVJFRF9UUE'
    'xfUkVTRVQQARIXChNJTlZBTElEX0ROU19BRERSRVNTEAISEwoPVE9PX01BTllfVkFMVUVTEAM=');

@$core.Deprecated('Use setProtocolStatusDescriptor instead')
const SetProtocolStatus$json = {
  '1': 'SetProtocolStatus',
  '2': [
    {'1': 'PROTOCOL_CONFIGURED', '2': 0},
    {'1': 'PROTOCOL_CONFIGURED_VPN_ON', '2': 1},
    {'1': 'INVALID_TECHNOLOGY', '2': 2},
  ],
};

/// Descriptor for `SetProtocolStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List setProtocolStatusDescriptor = $convert.base64Decode(
    'ChFTZXRQcm90b2NvbFN0YXR1cxIXChNQUk9UT0NPTF9DT05GSUdVUkVEEAASHgoaUFJPVE9DT0'
    'xfQ09ORklHVVJFRF9WUE5fT04QARIWChJJTlZBTElEX1RFQ0hOT0xPR1kQAg==');

@$core.Deprecated('Use setLANDiscoveryStatusDescriptor instead')
const SetLANDiscoveryStatus$json = {
  '1': 'SetLANDiscoveryStatus',
  '2': [
    {'1': 'DISCOVERY_CONFIGURED', '2': 0},
    {'1': 'DISCOVERY_CONFIGURED_ALLOWLIST_RESET', '2': 1},
  ],
};

/// Descriptor for `SetLANDiscoveryStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List setLANDiscoveryStatusDescriptor = $convert.base64Decode(
    'ChVTZXRMQU5EaXNjb3ZlcnlTdGF0dXMSGAoURElTQ09WRVJZX0NPTkZJR1VSRUQQABIoCiRESV'
    'NDT1ZFUllfQ09ORklHVVJFRF9BTExPV0xJU1RfUkVTRVQQAQ==');

@$core.Deprecated('Use setAutoconnectRequestDescriptor instead')
const SetAutoconnectRequest$json = {
  '1': 'SetAutoconnectRequest',
  '2': [
    {'1': 'enabled', '3': 1, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'server_tag', '3': 2, '4': 1, '5': 9, '10': 'serverTag'},
    {'1': 'server_group', '3': 3, '4': 1, '5': 9, '10': 'serverGroup'},
  ],
};

/// Descriptor for `SetAutoconnectRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setAutoconnectRequestDescriptor = $convert.base64Decode(
    'ChVTZXRBdXRvY29ubmVjdFJlcXVlc3QSGAoHZW5hYmxlZBgBIAEoCFIHZW5hYmxlZBIdCgpzZX'
    'J2ZXJfdGFnGAIgASgJUglzZXJ2ZXJUYWcSIQoMc2VydmVyX2dyb3VwGAMgASgJUgtzZXJ2ZXJH'
    'cm91cA==');

@$core.Deprecated('Use setGenericRequestDescriptor instead')
const SetGenericRequest$json = {
  '1': 'SetGenericRequest',
  '2': [
    {'1': 'enabled', '3': 1, '4': 1, '5': 8, '10': 'enabled'},
  ],
};

/// Descriptor for `SetGenericRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setGenericRequestDescriptor = $convert.base64Decode(
    'ChFTZXRHZW5lcmljUmVxdWVzdBIYCgdlbmFibGVkGAEgASgIUgdlbmFibGVk');

@$core.Deprecated('Use setUint32RequestDescriptor instead')
const SetUint32Request$json = {
  '1': 'SetUint32Request',
  '2': [
    {'1': 'value', '3': 1, '4': 1, '5': 13, '10': 'value'},
  ],
};

/// Descriptor for `SetUint32Request`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setUint32RequestDescriptor = $convert
    .base64Decode('ChBTZXRVaW50MzJSZXF1ZXN0EhQKBXZhbHVlGAEgASgNUgV2YWx1ZQ==');

@$core.Deprecated('Use setThreatProtectionLiteRequestDescriptor instead')
const SetThreatProtectionLiteRequest$json = {
  '1': 'SetThreatProtectionLiteRequest',
  '2': [
    {
      '1': 'threat_protection_lite',
      '3': 1,
      '4': 1,
      '5': 8,
      '10': 'threatProtectionLite'
    },
  ],
};

/// Descriptor for `SetThreatProtectionLiteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setThreatProtectionLiteRequestDescriptor =
    $convert.base64Decode(
        'Ch5TZXRUaHJlYXRQcm90ZWN0aW9uTGl0ZVJlcXVlc3QSNAoWdGhyZWF0X3Byb3RlY3Rpb25fbG'
        'l0ZRgBIAEoCFIUdGhyZWF0UHJvdGVjdGlvbkxpdGU=');

@$core.Deprecated('Use setThreatProtectionLiteResponseDescriptor instead')
const SetThreatProtectionLiteResponse$json = {
  '1': 'SetThreatProtectionLiteResponse',
  '2': [
    {
      '1': 'error_code',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.pb.SetErrorCode',
      '9': 0,
      '10': 'errorCode'
    },
    {
      '1': 'set_threat_protection_lite_status',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.pb.SetThreatProtectionLiteStatus',
      '9': 0,
      '10': 'setThreatProtectionLiteStatus'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `SetThreatProtectionLiteResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setThreatProtectionLiteResponseDescriptor =
    $convert.base64Decode(
        'Ch9TZXRUaHJlYXRQcm90ZWN0aW9uTGl0ZVJlc3BvbnNlEjEKCmVycm9yX2NvZGUYASABKA4yEC'
        '5wYi5TZXRFcnJvckNvZGVIAFIJZXJyb3JDb2RlEm0KIXNldF90aHJlYXRfcHJvdGVjdGlvbl9s'
        'aXRlX3N0YXR1cxgCIAEoDjIhLnBiLlNldFRocmVhdFByb3RlY3Rpb25MaXRlU3RhdHVzSABSHX'
        'NldFRocmVhdFByb3RlY3Rpb25MaXRlU3RhdHVzQgoKCHJlc3BvbnNl');

@$core.Deprecated('Use setDNSRequestDescriptor instead')
const SetDNSRequest$json = {
  '1': 'SetDNSRequest',
  '2': [
    {'1': 'dns', '3': 2, '4': 3, '5': 9, '10': 'dns'},
    {
      '1': 'threat_protection_lite',
      '3': 3,
      '4': 1,
      '5': 8,
      '10': 'threatProtectionLite'
    },
  ],
};

/// Descriptor for `SetDNSRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setDNSRequestDescriptor = $convert.base64Decode(
    'Cg1TZXRETlNSZXF1ZXN0EhAKA2RucxgCIAMoCVIDZG5zEjQKFnRocmVhdF9wcm90ZWN0aW9uX2'
    'xpdGUYAyABKAhSFHRocmVhdFByb3RlY3Rpb25MaXRl');

@$core.Deprecated('Use setDNSResponseDescriptor instead')
const SetDNSResponse$json = {
  '1': 'SetDNSResponse',
  '2': [
    {
      '1': 'error_code',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.pb.SetErrorCode',
      '9': 0,
      '10': 'errorCode'
    },
    {
      '1': 'set_dns_status',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.pb.SetDNSStatus',
      '9': 0,
      '10': 'setDnsStatus'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `SetDNSResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setDNSResponseDescriptor = $convert.base64Decode(
    'Cg5TZXRETlNSZXNwb25zZRIxCgplcnJvcl9jb2RlGAIgASgOMhAucGIuU2V0RXJyb3JDb2RlSA'
    'BSCWVycm9yQ29kZRI4Cg5zZXRfZG5zX3N0YXR1cxgDIAEoDjIQLnBiLlNldEROU1N0YXR1c0gA'
    'UgxzZXREbnNTdGF0dXNCCgoIcmVzcG9uc2U=');

@$core.Deprecated('Use setKillSwitchRequestDescriptor instead')
const SetKillSwitchRequest$json = {
  '1': 'SetKillSwitchRequest',
  '2': [
    {'1': 'kill_switch', '3': 2, '4': 1, '5': 8, '10': 'killSwitch'},
  ],
};

/// Descriptor for `SetKillSwitchRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setKillSwitchRequestDescriptor = $convert.base64Decode(
    'ChRTZXRLaWxsU3dpdGNoUmVxdWVzdBIfCgtraWxsX3N3aXRjaBgCIAEoCFIKa2lsbFN3aXRjaA'
    '==');

@$core.Deprecated('Use setNotifyRequestDescriptor instead')
const SetNotifyRequest$json = {
  '1': 'SetNotifyRequest',
  '2': [
    {'1': 'notify', '3': 3, '4': 1, '5': 8, '10': 'notify'},
  ],
};

/// Descriptor for `SetNotifyRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setNotifyRequestDescriptor = $convert
    .base64Decode('ChBTZXROb3RpZnlSZXF1ZXN0EhYKBm5vdGlmeRgDIAEoCFIGbm90aWZ5');

@$core.Deprecated('Use setTrayRequestDescriptor instead')
const SetTrayRequest$json = {
  '1': 'SetTrayRequest',
  '2': [
    {'1': 'uid', '3': 2, '4': 1, '5': 3, '10': 'uid'},
    {'1': 'tray', '3': 3, '4': 1, '5': 8, '10': 'tray'},
  ],
};

/// Descriptor for `SetTrayRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setTrayRequestDescriptor = $convert.base64Decode(
    'Cg5TZXRUcmF5UmVxdWVzdBIQCgN1aWQYAiABKANSA3VpZBISCgR0cmF5GAMgASgIUgR0cmF5');

@$core.Deprecated('Use setProtocolRequestDescriptor instead')
const SetProtocolRequest$json = {
  '1': 'SetProtocolRequest',
  '2': [
    {
      '1': 'protocol',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.config.Protocol',
      '10': 'protocol'
    },
  ],
};

/// Descriptor for `SetProtocolRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setProtocolRequestDescriptor = $convert.base64Decode(
    'ChJTZXRQcm90b2NvbFJlcXVlc3QSLAoIcHJvdG9jb2wYAiABKA4yEC5jb25maWcuUHJvdG9jb2'
    'xSCHByb3RvY29s');

@$core.Deprecated('Use setProtocolResponseDescriptor instead')
const SetProtocolResponse$json = {
  '1': 'SetProtocolResponse',
  '2': [
    {
      '1': 'error_code',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.pb.SetErrorCode',
      '9': 0,
      '10': 'errorCode'
    },
    {
      '1': 'set_protocol_status',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.pb.SetProtocolStatus',
      '9': 0,
      '10': 'setProtocolStatus'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `SetProtocolResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setProtocolResponseDescriptor = $convert.base64Decode(
    'ChNTZXRQcm90b2NvbFJlc3BvbnNlEjEKCmVycm9yX2NvZGUYASABKA4yEC5wYi5TZXRFcnJvck'
    'NvZGVIAFIJZXJyb3JDb2RlEkcKE3NldF9wcm90b2NvbF9zdGF0dXMYAiABKA4yFS5wYi5TZXRQ'
    'cm90b2NvbFN0YXR1c0gAUhFzZXRQcm90b2NvbFN0YXR1c0IKCghyZXNwb25zZQ==');

@$core.Deprecated('Use setTechnologyRequestDescriptor instead')
const SetTechnologyRequest$json = {
  '1': 'SetTechnologyRequest',
  '2': [
    {
      '1': 'technology',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.config.Technology',
      '10': 'technology'
    },
  ],
};

/// Descriptor for `SetTechnologyRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setTechnologyRequestDescriptor = $convert.base64Decode(
    'ChRTZXRUZWNobm9sb2d5UmVxdWVzdBIyCgp0ZWNobm9sb2d5GAIgASgOMhIuY29uZmlnLlRlY2'
    'hub2xvZ3lSCnRlY2hub2xvZ3k=');

@$core.Deprecated('Use portRangeDescriptor instead')
const PortRange$json = {
  '1': 'PortRange',
  '2': [
    {'1': 'start_port', '3': 1, '4': 1, '5': 3, '10': 'startPort'},
    {'1': 'end_port', '3': 2, '4': 1, '5': 3, '10': 'endPort'},
  ],
};

/// Descriptor for `PortRange`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List portRangeDescriptor = $convert.base64Decode(
    'CglQb3J0UmFuZ2USHQoKc3RhcnRfcG9ydBgBIAEoA1IJc3RhcnRQb3J0EhkKCGVuZF9wb3J0GA'
    'IgASgDUgdlbmRQb3J0');

@$core.Deprecated('Use setAllowlistSubnetRequestDescriptor instead')
const SetAllowlistSubnetRequest$json = {
  '1': 'SetAllowlistSubnetRequest',
  '2': [
    {'1': 'subnet', '3': 1, '4': 1, '5': 9, '10': 'subnet'},
  ],
};

/// Descriptor for `SetAllowlistSubnetRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setAllowlistSubnetRequestDescriptor =
    $convert.base64Decode(
        'ChlTZXRBbGxvd2xpc3RTdWJuZXRSZXF1ZXN0EhYKBnN1Ym5ldBgBIAEoCVIGc3VibmV0');

@$core.Deprecated('Use setAllowlistPortsRequestDescriptor instead')
const SetAllowlistPortsRequest$json = {
  '1': 'SetAllowlistPortsRequest',
  '2': [
    {'1': 'is_udp', '3': 1, '4': 1, '5': 8, '10': 'isUdp'},
    {'1': 'is_tcp', '3': 2, '4': 1, '5': 8, '10': 'isTcp'},
    {
      '1': 'port_range',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.pb.PortRange',
      '10': 'portRange'
    },
  ],
};

/// Descriptor for `SetAllowlistPortsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setAllowlistPortsRequestDescriptor = $convert.base64Decode(
    'ChhTZXRBbGxvd2xpc3RQb3J0c1JlcXVlc3QSFQoGaXNfdWRwGAEgASgIUgVpc1VkcBIVCgZpc1'
    '90Y3AYAiABKAhSBWlzVGNwEiwKCnBvcnRfcmFuZ2UYAyABKAsyDS5wYi5Qb3J0UmFuZ2VSCXBv'
    'cnRSYW5nZQ==');

@$core.Deprecated('Use setAllowlistRequestDescriptor instead')
const SetAllowlistRequest$json = {
  '1': 'SetAllowlistRequest',
  '2': [
    {
      '1': 'set_allowlist_subnet_request',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.pb.SetAllowlistSubnetRequest',
      '9': 0,
      '10': 'setAllowlistSubnetRequest'
    },
    {
      '1': 'set_allowlist_ports_request',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.pb.SetAllowlistPortsRequest',
      '9': 0,
      '10': 'setAllowlistPortsRequest'
    },
  ],
  '8': [
    {'1': 'request'},
  ],
};

/// Descriptor for `SetAllowlistRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setAllowlistRequestDescriptor = $convert.base64Decode(
    'ChNTZXRBbGxvd2xpc3RSZXF1ZXN0EmAKHHNldF9hbGxvd2xpc3Rfc3VibmV0X3JlcXVlc3QYAS'
    'ABKAsyHS5wYi5TZXRBbGxvd2xpc3RTdWJuZXRSZXF1ZXN0SABSGXNldEFsbG93bGlzdFN1Ym5l'
    'dFJlcXVlc3QSXQobc2V0X2FsbG93bGlzdF9wb3J0c19yZXF1ZXN0GAIgASgLMhwucGIuU2V0QW'
    'xsb3dsaXN0UG9ydHNSZXF1ZXN0SABSGHNldEFsbG93bGlzdFBvcnRzUmVxdWVzdEIJCgdyZXF1'
    'ZXN0');

@$core.Deprecated('Use setLANDiscoveryRequestDescriptor instead')
const SetLANDiscoveryRequest$json = {
  '1': 'SetLANDiscoveryRequest',
  '2': [
    {'1': 'enabled', '3': 1, '4': 1, '5': 8, '10': 'enabled'},
  ],
};

/// Descriptor for `SetLANDiscoveryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setLANDiscoveryRequestDescriptor =
    $convert.base64Decode(
        'ChZTZXRMQU5EaXNjb3ZlcnlSZXF1ZXN0EhgKB2VuYWJsZWQYASABKAhSB2VuYWJsZWQ=');

@$core.Deprecated('Use setLANDiscoveryResponseDescriptor instead')
const SetLANDiscoveryResponse$json = {
  '1': 'SetLANDiscoveryResponse',
  '2': [
    {
      '1': 'error_code',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.pb.SetErrorCode',
      '9': 0,
      '10': 'errorCode'
    },
    {
      '1': 'set_lan_discovery_status',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.pb.SetLANDiscoveryStatus',
      '9': 0,
      '10': 'setLanDiscoveryStatus'
    },
  ],
  '8': [
    {'1': 'response'},
  ],
};

/// Descriptor for `SetLANDiscoveryResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setLANDiscoveryResponseDescriptor = $convert.base64Decode(
    'ChdTZXRMQU5EaXNjb3ZlcnlSZXNwb25zZRIxCgplcnJvcl9jb2RlGAEgASgOMhAucGIuU2V0RX'
    'Jyb3JDb2RlSABSCWVycm9yQ29kZRJUChhzZXRfbGFuX2Rpc2NvdmVyeV9zdGF0dXMYAiABKA4y'
    'GS5wYi5TZXRMQU5EaXNjb3ZlcnlTdGF0dXNIAFIVc2V0TGFuRGlzY292ZXJ5U3RhdHVzQgoKCH'
    'Jlc3BvbnNl');
