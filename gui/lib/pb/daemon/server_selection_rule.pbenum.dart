// This is a generated file - do not edit.
//
// Generated from server_selection_rule.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class ServerSelectionRule extends $pb.ProtobufEnum {
  static const ServerSelectionRule NONE =
      ServerSelectionRule._(0, _omitEnumNames ? '' : 'NONE');
  static const ServerSelectionRule RECOMMENDED =
      ServerSelectionRule._(1, _omitEnumNames ? '' : 'RECOMMENDED');
  static const ServerSelectionRule CITY =
      ServerSelectionRule._(2, _omitEnumNames ? '' : 'CITY');
  static const ServerSelectionRule COUNTRY =
      ServerSelectionRule._(3, _omitEnumNames ? '' : 'COUNTRY');
  static const ServerSelectionRule SPECIFIC_SERVER =
      ServerSelectionRule._(4, _omitEnumNames ? '' : 'SPECIFIC_SERVER');
  static const ServerSelectionRule GROUP =
      ServerSelectionRule._(5, _omitEnumNames ? '' : 'GROUP');
  static const ServerSelectionRule COUNTRY_WITH_GROUP =
      ServerSelectionRule._(6, _omitEnumNames ? '' : 'COUNTRY_WITH_GROUP');
  static const ServerSelectionRule SPECIFIC_SERVER_WITH_GROUP =
      ServerSelectionRule._(
          7, _omitEnumNames ? '' : 'SPECIFIC_SERVER_WITH_GROUP');

  static const $core.List<ServerSelectionRule> values = <ServerSelectionRule>[
    NONE,
    RECOMMENDED,
    CITY,
    COUNTRY,
    SPECIFIC_SERVER,
    GROUP,
    COUNTRY_WITH_GROUP,
    SPECIFIC_SERVER_WITH_GROUP,
  ];

  static final $core.List<ServerSelectionRule?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 7);
  static ServerSelectionRule? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ServerSelectionRule._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
