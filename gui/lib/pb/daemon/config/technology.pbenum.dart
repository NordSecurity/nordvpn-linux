// This is a generated file - do not edit.
//
// Generated from technology.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class Technology extends $pb.ProtobufEnum {
  static const Technology UNKNOWN_TECHNOLOGY =
      Technology._(0, _omitEnumNames ? '' : 'UNKNOWN_TECHNOLOGY');
  static const Technology OPENVPN =
      Technology._(1, _omitEnumNames ? '' : 'OPENVPN');
  static const Technology NORDLYNX =
      Technology._(2, _omitEnumNames ? '' : 'NORDLYNX');
  static const Technology NORDWHISPER =
      Technology._(3, _omitEnumNames ? '' : 'NORDWHISPER');

  static const $core.List<Technology> values = <Technology>[
    UNKNOWN_TECHNOLOGY,
    OPENVPN,
    NORDLYNX,
    NORDWHISPER,
  ];

  static final $core.List<Technology?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static Technology? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const Technology._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
