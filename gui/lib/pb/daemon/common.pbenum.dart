// This is a generated file - do not edit.
//
// Generated from common.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class DaemonApiVersion extends $pb.ProtobufEnum {
  static const DaemonApiVersion UNKNOWN_VERSION =
      DaemonApiVersion._(0, _omitEnumNames ? '' : 'UNKNOWN_VERSION');
  static const DaemonApiVersion CURRENT_VERSION =
      DaemonApiVersion._(4, _omitEnumNames ? '' : 'CURRENT_VERSION');

  static const $core.List<DaemonApiVersion> values = <DaemonApiVersion>[
    UNKNOWN_VERSION,
    CURRENT_VERSION,
  ];

  static final $core.Map<$core.int, DaemonApiVersion> _byValue =
      $pb.ProtobufEnum.initByValue(values);
  static DaemonApiVersion? valueOf($core.int value) => _byValue[value];

  const DaemonApiVersion._(super.value, super.name);
}

class TriState extends $pb.ProtobufEnum {
  static const TriState UNKNOWN =
      TriState._(0, _omitEnumNames ? '' : 'UNKNOWN');
  static const TriState DISABLED =
      TriState._(1, _omitEnumNames ? '' : 'DISABLED');
  static const TriState ENABLED =
      TriState._(2, _omitEnumNames ? '' : 'ENABLED');

  static const $core.List<TriState> values = <TriState>[
    UNKNOWN,
    DISABLED,
    ENABLED,
  ];

  static final $core.List<TriState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static TriState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const TriState._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
