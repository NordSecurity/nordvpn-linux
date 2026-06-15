// This is a generated file - do not edit.
//
// Generated from uievent.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use uIEventDescriptor instead')
const UIEvent$json = {
  '1': 'UIEvent',
  '4': [
    UIEvent_FormReference$json,
    UIEvent_ItemName$json,
    UIEvent_ItemType$json,
    UIEvent_ItemValue$json
  ],
};

@$core.Deprecated('Use uIEventDescriptor instead')
const UIEvent_FormReference$json = {
  '1': 'FormReference',
  '2': [
    {'1': 'FORM_REFERENCE_UNSPECIFIED', '2': 0},
    {'1': 'CLI', '2': 1},
    {'1': 'TRAY', '2': 2},
    {'1': 'HOME_SCREEN', '2': 3},
    {'1': 'GUI', '2': 4},
    {'1': 'CONNECTION_INFO', '2': 5},
  ],
};

@$core.Deprecated('Use uIEventDescriptor instead')
const UIEvent_ItemName$json = {
  '1': 'ItemName',
  '2': [
    {'1': 'ITEM_NAME_UNSPECIFIED', '2': 0},
    {'1': 'CONNECT', '2': 1},
    {'1': 'CONNECT_RECENTS', '2': 2},
    {'1': 'DISCONNECT', '2': 3},
    {'1': 'LOGIN', '2': 4},
    {'1': 'LOGOUT', '2': 5},
    {'1': 'RATE_CONNECTION', '2': 6},
    {'1': 'MESHNET_INVITE_SEND', '2': 7},
    {'1': 'LOGIN_TOKEN', '2': 8},
    {'1': 'PAUSE', '2': 9},
    {'1': 'RECONNECT', '2': 10},
    {'1': 'CHANGE_SETTINGS', '2': 11},
    {'1': 'GET_HELP', '2': 12},
  ],
};

@$core.Deprecated('Use uIEventDescriptor instead')
const UIEvent_ItemType$json = {
  '1': 'ItemType',
  '2': [
    {'1': 'ITEM_TYPE_UNSPECIFIED', '2': 0},
    {'1': 'CLICK', '2': 1},
  ],
};

@$core.Deprecated('Use uIEventDescriptor instead')
const UIEvent_ItemValue$json = {
  '1': 'ItemValue',
  '2': [
    {'1': 'ITEM_VALUE_UNSPECIFIED', '2': 0},
    {'1': 'COUNTRY', '2': 1},
    {'1': 'CITY', '2': 2},
    {'1': 'DIP', '2': 3},
    {'1': 'MESHNET', '2': 4},
    {'1': 'OBFUSCATED', '2': 5},
    {'1': 'ONION_OVER_VPN', '2': 6},
    {'1': 'DOUBLE_VPN', '2': 7},
    {'1': 'P2P', '2': 8},
    {'1': 'PAUSE_5_MIN', '2': 9},
    {'1': 'PAUSE_15_MIN', '2': 10},
    {'1': 'PAUSE_30_MIN', '2': 11},
    {'1': 'PAUSE_1_HOUR', '2': 12},
    {'1': 'PAUSE_24_HOURS', '2': 13},
    {'1': 'PAUSE_DISCONNECT', '2': 14},
    {'1': 'DEDICATED_SERVER', '2': 15},
  ],
};

/// Descriptor for `UIEvent`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List uIEventDescriptor = $convert.base64Decode(
    'CgdVSUV2ZW50InEKDUZvcm1SZWZlcmVuY2USHgoaRk9STV9SRUZFUkVOQ0VfVU5TUEVDSUZJRU'
    'QQABIHCgNDTEkQARIICgRUUkFZEAISDwoLSE9NRV9TQ1JFRU4QAxIHCgNHVUkQBBITCg9DT05O'
    'RUNUSU9OX0lORk8QBSLqAQoISXRlbU5hbWUSGQoVSVRFTV9OQU1FX1VOU1BFQ0lGSUVEEAASCw'
    'oHQ09OTkVDVBABEhMKD0NPTk5FQ1RfUkVDRU5UUxACEg4KCkRJU0NPTk5FQ1QQAxIJCgVMT0dJ'
    'ThAEEgoKBkxPR09VVBAFEhMKD1JBVEVfQ09OTkVDVElPThAGEhcKE01FU0hORVRfSU5WSVRFX1'
    'NFTkQQBxIPCgtMT0dJTl9UT0tFThAIEgkKBVBBVVNFEAkSDQoJUkVDT05ORUNUEAoSEwoPQ0hB'
    'TkdFX1NFVFRJTkdTEAsSDAoIR0VUX0hFTFAQDCIwCghJdGVtVHlwZRIZChVJVEVNX1RZUEVfVU'
    '5TUEVDSUZJRUQQABIJCgVDTElDSxABIpgCCglJdGVtVmFsdWUSGgoWSVRFTV9WQUxVRV9VTlNQ'
    'RUNJRklFRBAAEgsKB0NPVU5UUlkQARIICgRDSVRZEAISBwoDRElQEAMSCwoHTUVTSE5FVBAEEg'
    '4KCk9CRlVTQ0FURUQQBRISCg5PTklPTl9PVkVSX1ZQThAGEg4KCkRPVUJMRV9WUE4QBxIHCgNQ'
    'MlAQCBIPCgtQQVVTRV81X01JThAJEhAKDFBBVVNFXzE1X01JThAKEhAKDFBBVVNFXzMwX01JTh'
    'ALEhAKDFBBVVNFXzFfSE9VUhAMEhIKDlBBVVNFXzI0X0hPVVJTEA0SFAoQUEFVU0VfRElTQ09O'
    'TkVDVBAOEhQKEERFRElDQVRFRF9TRVJWRVIQDw==');
