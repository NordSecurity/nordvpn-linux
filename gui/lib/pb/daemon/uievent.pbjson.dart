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

@$core.Deprecated('Use uIFormReferenceDescriptor instead')
const UIFormReference$json = {
  '1': 'UIFormReference',
  '2': [
    {'1': 'UI_FORM_REFERENCE_UNSPECIFIED', '2': 0},
    {'1': 'UI_FORM_REFERENCE_CLI', '2': 1},
    {'1': 'UI_FORM_REFERENCE_TRAY', '2': 2},
    {'1': 'UI_FORM_REFERENCE_HOME_SCREEN', '2': 3},
  ],
};

/// Descriptor for `UIFormReference`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List uIFormReferenceDescriptor = $convert.base64Decode(
    'Cg9VSUZvcm1SZWZlcmVuY2USIQodVUlfRk9STV9SRUZFUkVOQ0VfVU5TUEVDSUZJRUQQABIZCh'
    'VVSV9GT1JNX1JFRkVSRU5DRV9DTEkQARIaChZVSV9GT1JNX1JFRkVSRU5DRV9UUkFZEAISIQod'
    'VUlfRk9STV9SRUZFUkVOQ0VfSE9NRV9TQ1JFRU4QAw==');

@$core.Deprecated('Use uIItemNameDescriptor instead')
const UIItemName$json = {
  '1': 'UIItemName',
  '2': [
    {'1': 'UI_ITEM_NAME_UNSPECIFIED', '2': 0},
    {'1': 'UI_ITEM_NAME_CONNECT', '2': 1},
    {'1': 'UI_ITEM_NAME_CONNECT_RECENTS', '2': 2},
    {'1': 'UI_ITEM_NAME_DISCONNECT', '2': 3},
    {'1': 'UI_ITEM_NAME_LOGIN', '2': 4},
    {'1': 'UI_ITEM_NAME_LOGOUT', '2': 5},
    {'1': 'UI_ITEM_NAME_RATE_CONNECTION', '2': 6},
    {'1': 'UI_ITEM_NAME_MESHNET_INVITE_SEND', '2': 7},
  ],
};

/// Descriptor for `UIItemName`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List uIItemNameDescriptor = $convert.base64Decode(
    'CgpVSUl0ZW1OYW1lEhwKGFVJX0lURU1fTkFNRV9VTlNQRUNJRklFRBAAEhgKFFVJX0lURU1fTk'
    'FNRV9DT05ORUNUEAESIAocVUlfSVRFTV9OQU1FX0NPTk5FQ1RfUkVDRU5UUxACEhsKF1VJX0lU'
    'RU1fTkFNRV9ESVNDT05ORUNUEAMSFgoSVUlfSVRFTV9OQU1FX0xPR0lOEAQSFwoTVUlfSVRFTV'
    '9OQU1FX0xPR09VVBAFEiAKHFVJX0lURU1fTkFNRV9SQVRFX0NPTk5FQ1RJT04QBhIkCiBVSV9J'
    'VEVNX05BTUVfTUVTSE5FVF9JTlZJVEVfU0VORBAH');

@$core.Deprecated('Use uIItemTypeDescriptor instead')
const UIItemType$json = {
  '1': 'UIItemType',
  '2': [
    {'1': 'UI_ITEM_TYPE_UNSPECIFIED', '2': 0},
    {'1': 'UI_ITEM_TYPE_CLICK', '2': 1},
    {'1': 'UI_ITEM_TYPE_SHOW', '2': 2},
  ],
};

/// Descriptor for `UIItemType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List uIItemTypeDescriptor = $convert.base64Decode(
    'CgpVSUl0ZW1UeXBlEhwKGFVJX0lURU1fVFlQRV9VTlNQRUNJRklFRBAAEhYKElVJX0lURU1fVF'
    'lQRV9DTElDSxABEhUKEVVJX0lURU1fVFlQRV9TSE9XEAI=');

@$core.Deprecated('Use uIItemValueDescriptor instead')
const UIItemValue$json = {
  '1': 'UIItemValue',
  '2': [
    {'1': 'UI_ITEM_VALUE_CONNECTION_UNSPECIFIED', '2': 0},
    {'1': 'UI_ITEM_VALUE_CONNECTION_COUNTRY', '2': 1},
    {'1': 'UI_ITEM_VALUE_CONNECTION_CITY', '2': 2},
    {'1': 'UI_ITEM_VALUE_CONNECTION_DIP', '2': 3},
    {'1': 'UI_ITEM_VALUE_CONNECTION_MESHNET', '2': 4},
    {'1': 'UI_ITEM_VALUE_CONNECTION_OBFUSCATED', '2': 5},
    {'1': 'UI_ITEM_VALUE_CONNECTION_ONION_OVER_VPN', '2': 6},
    {'1': 'UI_ITEM_VALUE_CONNECTION_DOUBLE_VPN', '2': 7},
    {'1': 'UI_ITEM_VALUE_CONNECTION_P2P', '2': 8},
  ],
};

/// Descriptor for `UIItemValue`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List uIItemValueDescriptor = $convert.base64Decode(
    'CgtVSUl0ZW1WYWx1ZRIoCiRVSV9JVEVNX1ZBTFVFX0NPTk5FQ1RJT05fVU5TUEVDSUZJRUQQAB'
    'IkCiBVSV9JVEVNX1ZBTFVFX0NPTk5FQ1RJT05fQ09VTlRSWRABEiEKHVVJX0lURU1fVkFMVUVf'
    'Q09OTkVDVElPTl9DSVRZEAISIAocVUlfSVRFTV9WQUxVRV9DT05ORUNUSU9OX0RJUBADEiQKIF'
    'VJX0lURU1fVkFMVUVfQ09OTkVDVElPTl9NRVNITkVUEAQSJwojVUlfSVRFTV9WQUxVRV9DT05O'
    'RUNUSU9OX09CRlVTQ0FURUQQBRIrCidVSV9JVEVNX1ZBTFVFX0NPTk5FQ1RJT05fT05JT05fT1'
    'ZFUl9WUE4QBhInCiNVSV9JVEVNX1ZBTFVFX0NPTk5FQ1RJT05fRE9VQkxFX1ZQThAHEiAKHFVJ'
    'X0lURU1fVkFMVUVfQ09OTkVDVElPTl9QMlAQCA==');
