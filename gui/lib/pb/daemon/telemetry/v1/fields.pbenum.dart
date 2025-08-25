// This is a generated file - do not edit.
//
// Generated from fields.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// Defines supported display protocols
class DisplayProtocol extends $pb.ProtobufEnum {
  static const DisplayProtocol DISPLAY_PROTOCOL_UNSPECIFIED = DisplayProtocol._(
      0, _omitEnumNames ? '' : 'DISPLAY_PROTOCOL_UNSPECIFIED');
  static const DisplayProtocol DISPLAY_PROTOCOL_UNKNOWN =
      DisplayProtocol._(1, _omitEnumNames ? '' : 'DISPLAY_PROTOCOL_UNKNOWN');
  static const DisplayProtocol DISPLAY_PROTOCOL_X11 =
      DisplayProtocol._(2, _omitEnumNames ? '' : 'DISPLAY_PROTOCOL_X11');
  static const DisplayProtocol DISPLAY_PROTOCOL_WAYLAND =
      DisplayProtocol._(3, _omitEnumNames ? '' : 'DISPLAY_PROTOCOL_WAYLAND');

  static const $core.List<DisplayProtocol> values = <DisplayProtocol>[
    DISPLAY_PROTOCOL_UNSPECIFIED,
    DISPLAY_PROTOCOL_UNKNOWN,
    DISPLAY_PROTOCOL_X11,
    DISPLAY_PROTOCOL_WAYLAND,
  ];

  static final $core.List<DisplayProtocol?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static DisplayProtocol? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const DisplayProtocol._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
