// This is a generated file - do not edit.
//
// Generated from transfer.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use directionDescriptor instead')
const Direction$json = {
  '1': 'Direction',
  '2': [
    {'1': 'UNKNOWN_DIRECTION', '2': 0},
    {'1': 'INCOMING', '2': 1},
    {'1': 'OUTGOING', '2': 2},
  ],
};

/// Descriptor for `Direction`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List directionDescriptor = $convert.base64Decode(
    'CglEaXJlY3Rpb24SFQoRVU5LTk9XTl9ESVJFQ1RJT04QABIMCghJTkNPTUlORxABEgwKCE9VVE'
    'dPSU5HEAI=');

@$core.Deprecated('Use statusDescriptor instead')
const Status$json = {
  '1': 'Status',
  '2': [
    {'1': 'SUCCESS', '2': 0},
    {'1': 'CANCELED', '2': 1},
    {'1': 'BAD_PATH', '2': 2},
    {'1': 'BAD_FILE', '2': 3},
    {'1': 'TRANSPORT', '2': 4},
    {'1': 'BAD_STATUS', '2': 5},
    {'1': 'SERVICE_STOP', '2': 6},
    {'1': 'BAD_TRANSFER', '2': 7},
    {'1': 'BAD_TRANSFER_STATE', '2': 8},
    {'1': 'BAD_FILE_ID', '2': 9},
    {'1': 'BAD_SYSTEM_TIME', '2': 10},
    {'1': 'TRUNCATED_FILE', '2': 11},
    {'1': 'EVENT_SEND', '2': 12},
    {'1': 'BAD_UUID', '2': 13},
    {'1': 'CHANNEL_CLOSED', '2': 14},
    {'1': 'IO', '2': 15},
    {'1': 'DATA_SEND', '2': 16},
    {'1': 'DIRECTORY_NOT_EXPECTED', '2': 17},
    {'1': 'EMPTY_TRANSFER', '2': 18},
    {'1': 'TRANSFER_CLOSED_BY_PEER', '2': 19},
    {'1': 'TRANSFER_LIMITS_EXCEEDED', '2': 20},
    {'1': 'MISMATCHED_SIZE', '2': 21},
    {'1': 'UNEXPECTED_DATA', '2': 22},
    {'1': 'INVALID_ARGUMENT', '2': 23},
    {'1': 'TRANSFER_TIMEOUT', '2': 24},
    {'1': 'WS_SERVER', '2': 25},
    {'1': 'WS_CLIENT', '2': 26},
    {'1': 'FILE_MODIFIED', '2': 28},
    {'1': 'FILENAME_TOO_LONG', '2': 29},
    {'1': 'AUTHENTICATION_FAILED', '2': 30},
    {'1': 'FILE_CHECKSUM_MISMATCH', '2': 33},
    {'1': 'FILE_REJECTED', '2': 34},
    {'1': 'REQUESTED', '2': 100},
    {'1': 'ONGOING', '2': 101},
    {'1': 'FINISHED_WITH_ERRORS', '2': 102},
    {'1': 'ACCEPT_FAILURE', '2': 103},
    {'1': 'CANCELED_BY_PEER', '2': 104},
    {'1': 'INTERRUPTED', '2': 105},
    {'1': 'PAUSED', '2': 106},
    {'1': 'PENDING', '2': 107},
  ],
};

/// Descriptor for `Status`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List statusDescriptor = $convert.base64Decode(
    'CgZTdGF0dXMSCwoHU1VDQ0VTUxAAEgwKCENBTkNFTEVEEAESDAoIQkFEX1BBVEgQAhIMCghCQU'
    'RfRklMRRADEg0KCVRSQU5TUE9SVBAEEg4KCkJBRF9TVEFUVVMQBRIQCgxTRVJWSUNFX1NUT1AQ'
    'BhIQCgxCQURfVFJBTlNGRVIQBxIWChJCQURfVFJBTlNGRVJfU1RBVEUQCBIPCgtCQURfRklMRV'
    '9JRBAJEhMKD0JBRF9TWVNURU1fVElNRRAKEhIKDlRSVU5DQVRFRF9GSUxFEAsSDgoKRVZFTlRf'
    'U0VORBAMEgwKCEJBRF9VVUlEEA0SEgoOQ0hBTk5FTF9DTE9TRUQQDhIGCgJJTxAPEg0KCURBVE'
    'FfU0VORBAQEhoKFkRJUkVDVE9SWV9OT1RfRVhQRUNURUQQERISCg5FTVBUWV9UUkFOU0ZFUhAS'
    'EhsKF1RSQU5TRkVSX0NMT1NFRF9CWV9QRUVSEBMSHAoYVFJBTlNGRVJfTElNSVRTX0VYQ0VFRE'
    'VEEBQSEwoPTUlTTUFUQ0hFRF9TSVpFEBUSEwoPVU5FWFBFQ1RFRF9EQVRBEBYSFAoQSU5WQUxJ'
    'RF9BUkdVTUVOVBAXEhQKEFRSQU5TRkVSX1RJTUVPVVQQGBINCglXU19TRVJWRVIQGRINCglXU1'
    '9DTElFTlQQGhIRCg1GSUxFX01PRElGSUVEEBwSFQoRRklMRU5BTUVfVE9PX0xPTkcQHRIZChVB'
    'VVRIRU5USUNBVElPTl9GQUlMRUQQHhIaChZGSUxFX0NIRUNLU1VNX01JU01BVENIECESEQoNRk'
    'lMRV9SRUpFQ1RFRBAiEg0KCVJFUVVFU1RFRBBkEgsKB09OR09JTkcQZRIYChRGSU5JU0hFRF9X'
    'SVRIX0VSUk9SUxBmEhIKDkFDQ0VQVF9GQUlMVVJFEGcSFAoQQ0FOQ0VMRURfQllfUEVFUhBoEg'
    '8KC0lOVEVSUlVQVEVEEGkSCgoGUEFVU0VEEGoSCwoHUEVORElORxBr');

@$core.Deprecated('Use transferDescriptor instead')
const Transfer$json = {
  '1': 'Transfer',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {
      '1': 'direction',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.filesharepb.Direction',
      '10': 'direction'
    },
    {'1': 'peer', '3': 3, '4': 1, '5': 9, '10': 'peer'},
    {
      '1': 'status',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.filesharepb.Status',
      '10': 'status'
    },
    {
      '1': 'created',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'created'
    },
    {
      '1': 'files',
      '3': 6,
      '4': 3,
      '5': 11,
      '6': '.filesharepb.File',
      '10': 'files'
    },
    {'1': 'path', '3': 7, '4': 1, '5': 9, '10': 'path'},
    {'1': 'total_size', '3': 8, '4': 1, '5': 4, '10': 'totalSize'},
    {
      '1': 'total_transferred',
      '3': 9,
      '4': 1,
      '5': 4,
      '10': 'totalTransferred'
    },
  ],
};

/// Descriptor for `Transfer`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List transferDescriptor = $convert.base64Decode(
    'CghUcmFuc2ZlchIOCgJpZBgBIAEoCVICaWQSNAoJZGlyZWN0aW9uGAIgASgOMhYuZmlsZXNoYX'
    'JlcGIuRGlyZWN0aW9uUglkaXJlY3Rpb24SEgoEcGVlchgDIAEoCVIEcGVlchIrCgZzdGF0dXMY'
    'BCABKA4yEy5maWxlc2hhcmVwYi5TdGF0dXNSBnN0YXR1cxI0CgdjcmVhdGVkGAUgASgLMhouZ2'
    '9vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIHY3JlYXRlZBInCgVmaWxlcxgGIAMoCzIRLmZpbGVz'
    'aGFyZXBiLkZpbGVSBWZpbGVzEhIKBHBhdGgYByABKAlSBHBhdGgSHQoKdG90YWxfc2l6ZRgIIA'
    'EoBFIJdG90YWxTaXplEisKEXRvdGFsX3RyYW5zZmVycmVkGAkgASgEUhB0b3RhbFRyYW5zZmVy'
    'cmVk');

@$core.Deprecated('Use fileDescriptor instead')
const File$json = {
  '1': 'File',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'path', '3': 6, '4': 1, '5': 9, '10': 'path'},
    {'1': 'fullPath', '3': 7, '4': 1, '5': 9, '10': 'fullPath'},
    {'1': 'size', '3': 2, '4': 1, '5': 4, '10': 'size'},
    {'1': 'transferred', '3': 3, '4': 1, '5': 4, '10': 'transferred'},
    {
      '1': 'status',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.filesharepb.Status',
      '10': 'status'
    },
    {
      '1': 'children',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.filesharepb.File.ChildrenEntry',
      '10': 'children'
    },
  ],
  '3': [File_ChildrenEntry$json],
};

@$core.Deprecated('Use fileDescriptor instead')
const File_ChildrenEntry$json = {
  '1': 'ChildrenEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {
      '1': 'value',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.filesharepb.File',
      '10': 'value'
    },
  ],
  '7': {'7': true},
};

/// Descriptor for `File`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List fileDescriptor = $convert.base64Decode(
    'CgRGaWxlEg4KAmlkGAEgASgJUgJpZBISCgRwYXRoGAYgASgJUgRwYXRoEhoKCGZ1bGxQYXRoGA'
    'cgASgJUghmdWxsUGF0aBISCgRzaXplGAIgASgEUgRzaXplEiAKC3RyYW5zZmVycmVkGAMgASgE'
    'Ugt0cmFuc2ZlcnJlZBIrCgZzdGF0dXMYBCABKA4yEy5maWxlc2hhcmVwYi5TdGF0dXNSBnN0YX'
    'R1cxI7CghjaGlsZHJlbhgFIAMoCzIfLmZpbGVzaGFyZXBiLkZpbGUuQ2hpbGRyZW5FbnRyeVII'
    'Y2hpbGRyZW4aTgoNQ2hpbGRyZW5FbnRyeRIQCgNrZXkYASABKAlSA2tleRInCgV2YWx1ZRgCIA'
    'EoCzIRLmZpbGVzaGFyZXBiLkZpbGVSBXZhbHVlOgI4AQ==');
