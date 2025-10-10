// This is a generated file - do not edit.
//
// Generated from server_selection_rule.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use serverSelectionRuleDescriptor instead')
const ServerSelectionRule$json = {
  '1': 'ServerSelectionRule',
  '2': [
    {'1': 'NONE', '2': 0},
    {'1': 'RECOMMENDED', '2': 1},
    {'1': 'CITY', '2': 2},
    {'1': 'COUNTRY', '2': 3},
    {'1': 'SPECIFIC_SERVER', '2': 4},
    {'1': 'GROUP', '2': 5},
    {'1': 'COUNTRY_WITH_GROUP', '2': 6},
    {'1': 'SPECIFIC_SERVER_WITH_GROUP', '2': 7},
  ],
};

/// Descriptor for `ServerSelectionRule`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List serverSelectionRuleDescriptor = $convert.base64Decode(
    'ChNTZXJ2ZXJTZWxlY3Rpb25SdWxlEggKBE5PTkUQABIPCgtSRUNPTU1FTkRFRBABEggKBENJVF'
    'kQAhILCgdDT1VOVFJZEAMSEwoPU1BFQ0lGSUNfU0VSVkVSEAQSCQoFR1JPVVAQBRIWChJDT1VO'
    'VFJZX1dJVEhfR1JPVVAQBhIeChpTUEVDSUZJQ19TRVJWRVJfV0lUSF9HUk9VUBAH');
