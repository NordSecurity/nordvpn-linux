// This is a generated file - do not edit.
//
// Generated from settings.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use settingsResponseDescriptor instead')
const SettingsResponse$json = {
  '1': 'SettingsResponse',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 3, '10': 'type'},
    {'1': 'data', '3': 2, '4': 1, '5': 11, '6': '.pb.Settings', '10': 'data'},
  ],
};

/// Descriptor for `SettingsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List settingsResponseDescriptor = $convert.base64Decode(
    'ChBTZXR0aW5nc1Jlc3BvbnNlEhIKBHR5cGUYASABKANSBHR5cGUSIAoEZGF0YRgCIAEoCzIMLn'
    'BiLlNldHRpbmdzUgRkYXRh');

@$core.Deprecated('Use settingsVPNProtocolsResponseDescriptor instead')
const SettingsVPNProtocolsResponse$json = {
  '1': 'SettingsVPNProtocolsResponse',
  '2': [
    {
      '1': 'vpn_protocols',
      '3': 1,
      '4': 3,
      '5': 14,
      '6': '.config.VPNProtocol',
      '10': 'vpnProtocols'
    },
  ],
};

/// Descriptor for `SettingsVPNProtocolsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List settingsVPNProtocolsResponseDescriptor =
    $convert.base64Decode(
        'ChxTZXR0aW5nc1ZQTlByb3RvY29sc1Jlc3BvbnNlEjgKDXZwbl9wcm90b2NvbHMYASADKA4yEy'
        '5jb25maWcuVlBOUHJvdG9jb2xSDHZwblByb3RvY29scw==');

@$core.Deprecated('Use autoconnectDataDescriptor instead')
const AutoconnectData$json = {
  '1': 'AutoconnectData',
  '2': [
    {'1': 'enabled', '3': 1, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'country', '3': 2, '4': 1, '5': 9, '10': 'country'},
    {'1': 'city', '3': 3, '4': 1, '5': 9, '10': 'city'},
    {
      '1': 'server_group',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.config.ServerGroup',
      '10': 'serverGroup'
    },
    {'1': 'country_code', '3': 5, '4': 1, '5': 9, '10': 'countryCode'},
  ],
};

/// Descriptor for `AutoconnectData`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List autoconnectDataDescriptor = $convert.base64Decode(
    'Cg9BdXRvY29ubmVjdERhdGESGAoHZW5hYmxlZBgBIAEoCFIHZW5hYmxlZBIYCgdjb3VudHJ5GA'
    'IgASgJUgdjb3VudHJ5EhIKBGNpdHkYAyABKAlSBGNpdHkSNgoMc2VydmVyX2dyb3VwGAQgASgO'
    'MhMuY29uZmlnLlNlcnZlckdyb3VwUgtzZXJ2ZXJHcm91cBIhCgxjb3VudHJ5X2NvZGUYBSABKA'
    'lSC2NvdW50cnlDb2Rl');

@$core.Deprecated('Use settingsDescriptor instead')
const Settings$json = {
  '1': 'Settings',
  '2': [
    {'1': 'firewall', '3': 2, '4': 1, '5': 8, '10': 'firewall'},
    {'1': 'kill_switch', '3': 3, '4': 1, '5': 8, '10': 'killSwitch'},
    {
      '1': 'auto_connect_data',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.pb.AutoconnectData',
      '10': 'autoConnectData'
    },
    {'1': 'meshnet', '3': 6, '4': 1, '5': 8, '10': 'meshnet'},
    {'1': 'routing', '3': 7, '4': 1, '5': 8, '10': 'routing'},
    {'1': 'fwmark', '3': 8, '4': 1, '5': 13, '10': 'fwmark'},
    {
      '1': 'analytics_consent',
      '3': 9,
      '4': 1,
      '5': 14,
      '6': '.consent.ConsentMode',
      '10': 'analyticsConsent'
    },
    {'1': 'dns', '3': 10, '4': 3, '5': 9, '10': 'dns'},
    {
      '1': 'real_time_protection',
      '3': 11,
      '4': 1,
      '5': 8,
      '10': 'realTimeProtection'
    },
    {'1': 'lan_discovery', '3': 13, '4': 1, '5': 8, '10': 'lanDiscovery'},
    {
      '1': 'allowlist',
      '3': 14,
      '4': 1,
      '5': 11,
      '6': '.pb.Allowlist',
      '10': 'allowlist'
    },
    {'1': 'postquantum_vpn', '3': 17, '4': 1, '5': 8, '10': 'postquantumVpn'},
    {
      '1': 'user_settings',
      '3': 18,
      '4': 1,
      '5': 11,
      '6': '.pb.UserSpecificSettings',
      '10': 'userSettings'
    },
    {'1': 'arp_ignore', '3': 19, '4': 1, '5': 8, '10': 'arpIgnore'},
    {'1': 'ech', '3': 20, '4': 1, '5': 8, '10': 'ech'},
    {
      '1': 'vpn_protocol',
      '3': 21,
      '4': 1,
      '5': 14,
      '6': '.config.VPNProtocol',
      '10': 'vpnProtocol'
    },
  ],
  '9': [
    {'1': 1, '2': 2},
    {'1': 12, '2': 13},
    {'1': 15, '2': 16},
    {'1': 16, '2': 17},
  ],
  '10': ['technology', 'protocol', 'obfuscate'],
};

/// Descriptor for `Settings`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List settingsDescriptor = $convert.base64Decode(
    'CghTZXR0aW5ncxIaCghmaXJld2FsbBgCIAEoCFIIZmlyZXdhbGwSHwoLa2lsbF9zd2l0Y2gYAy'
    'ABKAhSCmtpbGxTd2l0Y2gSPwoRYXV0b19jb25uZWN0X2RhdGEYBCABKAsyEy5wYi5BdXRvY29u'
    'bmVjdERhdGFSD2F1dG9Db25uZWN0RGF0YRIYCgdtZXNobmV0GAYgASgIUgdtZXNobmV0EhgKB3'
    'JvdXRpbmcYByABKAhSB3JvdXRpbmcSFgoGZndtYXJrGAggASgNUgZmd21hcmsSQQoRYW5hbHl0'
    'aWNzX2NvbnNlbnQYCSABKA4yFC5jb25zZW50LkNvbnNlbnRNb2RlUhBhbmFseXRpY3NDb25zZW'
    '50EhAKA2RucxgKIAMoCVIDZG5zEjAKFHJlYWxfdGltZV9wcm90ZWN0aW9uGAsgASgIUhJyZWFs'
    'VGltZVByb3RlY3Rpb24SIwoNbGFuX2Rpc2NvdmVyeRgNIAEoCFIMbGFuRGlzY292ZXJ5EisKCW'
    'FsbG93bGlzdBgOIAEoCzINLnBiLkFsbG93bGlzdFIJYWxsb3dsaXN0EicKD3Bvc3RxdWFudHVt'
    'X3ZwbhgRIAEoCFIOcG9zdHF1YW50dW1WcG4SPQoNdXNlcl9zZXR0aW5ncxgSIAEoCzIYLnBiLl'
    'VzZXJTcGVjaWZpY1NldHRpbmdzUgx1c2VyU2V0dGluZ3MSHQoKYXJwX2lnbm9yZRgTIAEoCFIJ'
    'YXJwSWdub3JlEhAKA2VjaBgUIAEoCFIDZWNoEjYKDHZwbl9wcm90b2NvbBgVIAEoDjITLmNvbm'
    'ZpZy5WUE5Qcm90b2NvbFILdnBuUHJvdG9jb2xKBAgBEAJKBAgMEA1KBAgPEBBKBAgQEBFSCnRl'
    'Y2hub2xvZ3lSCHByb3RvY29sUglvYmZ1c2NhdGU=');

@$core.Deprecated('Use userSpecificSettingsDescriptor instead')
const UserSpecificSettings$json = {
  '1': 'UserSpecificSettings',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 3, '10': 'uid'},
    {'1': 'notify', '3': 2, '4': 1, '5': 8, '10': 'notify'},
    {'1': 'tray', '3': 3, '4': 1, '5': 8, '10': 'tray'},
  ],
};

/// Descriptor for `UserSpecificSettings`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List userSpecificSettingsDescriptor = $convert.base64Decode(
    'ChRVc2VyU3BlY2lmaWNTZXR0aW5ncxIQCgN1aWQYASABKANSA3VpZBIWCgZub3RpZnkYAiABKA'
    'hSBm5vdGlmeRISCgR0cmF5GAMgASgIUgR0cmF5');
