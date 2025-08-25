// This is a generated file - do not edit.
//
// Generated from analytics_consent.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class ConsentMode extends $pb.ProtobufEnum {
  static const ConsentMode UNDEFINED =
      ConsentMode._(0, _omitEnumNames ? '' : 'UNDEFINED');
  static const ConsentMode GRANTED =
      ConsentMode._(1, _omitEnumNames ? '' : 'GRANTED');
  static const ConsentMode DENIED =
      ConsentMode._(2, _omitEnumNames ? '' : 'DENIED');

  static const $core.List<ConsentMode> values = <ConsentMode>[
    UNDEFINED,
    GRANTED,
    DENIED,
  ];

  static final $core.List<ConsentMode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static ConsentMode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ConsentMode._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
