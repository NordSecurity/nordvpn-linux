// This is a generated file - do not edit.
//
// Generated from protocol.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class Protocol extends $pb.ProtobufEnum {
  static const Protocol UNKNOWN_PROTOCOL =
      Protocol._(0, _omitEnumNames ? '' : 'UNKNOWN_PROTOCOL');
  static const Protocol UDP = Protocol._(1, _omitEnumNames ? '' : 'UDP');
  static const Protocol TCP = Protocol._(2, _omitEnumNames ? '' : 'TCP');
  static const Protocol Webtunnel =
      Protocol._(3, _omitEnumNames ? '' : 'Webtunnel');

  static const $core.List<Protocol> values = <Protocol>[
    UNKNOWN_PROTOCOL,
    UDP,
    TCP,
    Webtunnel,
  ];

  static final $core.List<Protocol?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static Protocol? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const Protocol._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
