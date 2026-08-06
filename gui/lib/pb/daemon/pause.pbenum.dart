// This is a generated file - do not edit.
//
// Generated from pause.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class PauseInverval extends $pb.ProtobufEnum {
  static const PauseInverval PAUSE_5_MIN =
      PauseInverval._(0, _omitEnumNames ? '' : 'PAUSE_5_MIN');
  static const PauseInverval PAUSE_15_MIN =
      PauseInverval._(1, _omitEnumNames ? '' : 'PAUSE_15_MIN');
  static const PauseInverval PAUSE_30_MIN =
      PauseInverval._(2, _omitEnumNames ? '' : 'PAUSE_30_MIN');
  static const PauseInverval PAUSE_1_HOUR =
      PauseInverval._(3, _omitEnumNames ? '' : 'PAUSE_1_HOUR');
  static const PauseInverval PAUSE_24_HOURS =
      PauseInverval._(4, _omitEnumNames ? '' : 'PAUSE_24_HOURS');

  static const $core.List<PauseInverval> values = <PauseInverval>[
    PAUSE_5_MIN,
    PAUSE_15_MIN,
    PAUSE_30_MIN,
    PAUSE_1_HOUR,
    PAUSE_24_HOURS,
  ];

  static final $core.List<PauseInverval?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static PauseInverval? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const PauseInverval._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
