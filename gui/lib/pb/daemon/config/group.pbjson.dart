// This is a generated file - do not edit.
//
// Generated from group.proto.

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

@$core.Deprecated('Use serverGroupDescriptor instead')
const ServerGroup$json = {
  '1': 'ServerGroup',
  '2': [
    {'1': 'UNDEFINED', '2': 0},
    {'1': 'DOUBLE_VPN', '2': 1},
    {'1': 'ONION_OVER_VPN', '2': 3},
    {'1': 'DEDICATED_IP', '2': 9},
    {'1': 'STANDARD_VPN_SERVERS', '2': 11},
    {'1': 'OVPN_OBFUSCATED', '2': 17},
    {'1': 'DEDICATED_SERVER', '2': 99},
    {'1': 'NW_OBFUSCATED', '2': 100},
  ],
  '4': [
    {'1': 5, '2': 5},
    {'1': 7, '2': 7},
    {'1': 13, '2': 13},
    {'1': 15, '2': 15},
    {'1': 19, '2': 19},
    {'1': 21, '2': 21},
    {'1': 23, '2': 23},
    {'1': 25, '2': 25},
  ],
  '5': [
    'ULTRA_FAST_TV',
    'ANTI_DDOS',
    'NETFLIX_USA',
    'P2P',
    'EUROPE',
    'THE_AMERICAS',
    'ASIA_PACIFIC',
    'AFRICA_THE_MIDDLE_EAST_AND_INDIA'
  ],
};

/// Descriptor for `ServerGroup`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List serverGroupDescriptor = $convert.base64Decode(
    'CgtTZXJ2ZXJHcm91cBINCglVTkRFRklORUQQABIOCgpET1VCTEVfVlBOEAESEgoOT05JT05fT1'
    'ZFUl9WUE4QAxIQCgxERURJQ0FURURfSVAQCRIYChRTVEFOREFSRF9WUE5fU0VSVkVSUxALEhMK'
    'D09WUE5fT0JGVVNDQVRFRBAREhQKEERFRElDQVRFRF9TRVJWRVIQYxIRCg1OV19PQkZVU0NBVE'
    'VEEGQiBAgFEAUiBAgHEAciBAgNEA0iBAgPEA8iBAgTEBMiBAgVEBUiBAgXEBciBAgZEBkqDVVM'
    'VFJBX0ZBU1RfVFYqCUFOVElfRERPUyoLTkVURkxJWF9VU0EqA1AyUCoGRVVST1BFKgxUSEVfQU'
    '1FUklDQVMqDEFTSUFfUEFDSUZJQyogQUZSSUNBX1RIRV9NSURETEVfRUFTVF9BTkRfSU5ESUE=');
