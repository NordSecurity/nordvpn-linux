// This is a generated file - do not edit.
//
// Generated from settings.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

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
  ],
};

/// Descriptor for `AutoconnectData`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List autoconnectDataDescriptor = $convert.base64Decode(
    'Cg9BdXRvY29ubmVjdERhdGESGAoHZW5hYmxlZBgBIAEoCFIHZW5hYmxlZBIYCgdjb3VudHJ5GA'
    'IgASgJUgdjb3VudHJ5EhIKBGNpdHkYAyABKAlSBGNpdHkSNgoMc2VydmVyX2dyb3VwGAQgASgO'
    'MhMuY29uZmlnLlNlcnZlckdyb3VwUgtzZXJ2ZXJHcm91cA==');

@$core.Deprecated('Use settingsDescriptor instead')
const Settings$json = {
  '1': 'Settings',
  '2': [
    {
      '1': 'technology',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.config.Technology',
      '10': 'technology'
    },
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
      '1': 'threat_protection_lite',
      '3': 11,
      '4': 1,
      '5': 8,
      '10': 'threatProtectionLite'
    },
    {
      '1': 'protocol',
      '3': 12,
      '4': 1,
      '5': 14,
      '6': '.config.Protocol',
      '10': 'protocol'
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
    {'1': 'obfuscate', '3': 15, '4': 1, '5': 8, '10': 'obfuscate'},
    {'1': 'virtualLocation', '3': 16, '4': 1, '5': 8, '10': 'virtualLocation'},
    {'1': 'postquantum_vpn', '3': 17, '4': 1, '5': 8, '10': 'postquantumVpn'},
    {
      '1': 'user_settings',
      '3': 18,
      '4': 1,
      '5': 11,
      '6': '.pb.UserSpecificSettings',
      '10': 'userSettings'
    },
  ],
};

/// Descriptor for `Settings`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List settingsDescriptor = $convert.base64Decode(
    'CghTZXR0aW5ncxIyCgp0ZWNobm9sb2d5GAEgASgOMhIuY29uZmlnLlRlY2hub2xvZ3lSCnRlY2'
    'hub2xvZ3kSGgoIZmlyZXdhbGwYAiABKAhSCGZpcmV3YWxsEh8KC2tpbGxfc3dpdGNoGAMgASgI'
    'UgpraWxsU3dpdGNoEj8KEWF1dG9fY29ubmVjdF9kYXRhGAQgASgLMhMucGIuQXV0b2Nvbm5lY3'
    'REYXRhUg9hdXRvQ29ubmVjdERhdGESGAoHbWVzaG5ldBgGIAEoCFIHbWVzaG5ldBIYCgdyb3V0'
    'aW5nGAcgASgIUgdyb3V0aW5nEhYKBmZ3bWFyaxgIIAEoDVIGZndtYXJrEkEKEWFuYWx5dGljc1'
    '9jb25zZW50GAkgASgOMhQuY29uc2VudC5Db25zZW50TW9kZVIQYW5hbHl0aWNzQ29uc2VudBIQ'
    'CgNkbnMYCiADKAlSA2RucxI0ChZ0aHJlYXRfcHJvdGVjdGlvbl9saXRlGAsgASgIUhR0aHJlYX'
    'RQcm90ZWN0aW9uTGl0ZRIsCghwcm90b2NvbBgMIAEoDjIQLmNvbmZpZy5Qcm90b2NvbFIIcHJv'
    'dG9jb2wSIwoNbGFuX2Rpc2NvdmVyeRgNIAEoCFIMbGFuRGlzY292ZXJ5EisKCWFsbG93bGlzdB'
    'gOIAEoCzINLnBiLkFsbG93bGlzdFIJYWxsb3dsaXN0EhwKCW9iZnVzY2F0ZRgPIAEoCFIJb2Jm'
    'dXNjYXRlEigKD3ZpcnR1YWxMb2NhdGlvbhgQIAEoCFIPdmlydHVhbExvY2F0aW9uEicKD3Bvc3'
    'RxdWFudHVtX3ZwbhgRIAEoCFIOcG9zdHF1YW50dW1WcG4SPQoNdXNlcl9zZXR0aW5ncxgSIAEo'
    'CzIYLnBiLlVzZXJTcGVjaWZpY1NldHRpbmdzUgx1c2VyU2V0dGluZ3M=');

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
