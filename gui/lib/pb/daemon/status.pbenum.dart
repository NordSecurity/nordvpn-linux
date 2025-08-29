// This is a generated file - do not edit.
//
// Generated from status.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class ConnectionSource extends $pb.ProtobufEnum {
  static const ConnectionSource UNKNOWN_SOURCE =
      ConnectionSource._(0, _omitEnumNames ? '' : 'UNKNOWN_SOURCE');
  static const ConnectionSource MANUAL =
      ConnectionSource._(1, _omitEnumNames ? '' : 'MANUAL');
  static const ConnectionSource AUTO =
      ConnectionSource._(2, _omitEnumNames ? '' : 'AUTO');

  static const $core.List<ConnectionSource> values = <ConnectionSource>[
    UNKNOWN_SOURCE,
    MANUAL,
    AUTO,
  ];

  static final $core.List<ConnectionSource?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static ConnectionSource? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ConnectionSource._(super.value, super.name);
}

class ConnectionState extends $pb.ProtobufEnum {
  static const ConnectionState UNKNOWN_STATE =
      ConnectionState._(0, _omitEnumNames ? '' : 'UNKNOWN_STATE');
  static const ConnectionState DISCONNECTED =
      ConnectionState._(1, _omitEnumNames ? '' : 'DISCONNECTED');
  static const ConnectionState CONNECTING =
      ConnectionState._(2, _omitEnumNames ? '' : 'CONNECTING');
  static const ConnectionState CONNECTED =
      ConnectionState._(3, _omitEnumNames ? '' : 'CONNECTED');

  static const $core.List<ConnectionState> values = <ConnectionState>[
    UNKNOWN_STATE,
    DISCONNECTED,
    CONNECTING,
    CONNECTED,
  ];

  static final $core.List<ConnectionState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static ConnectionState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ConnectionState._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
