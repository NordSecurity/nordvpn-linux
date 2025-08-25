// This is a generated file - do not edit.
//
// Generated from servers.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class ServersError extends $pb.ProtobufEnum {
  static const ServersError NO_ERROR =
      ServersError._(0, _omitEnumNames ? '' : 'NO_ERROR');
  static const ServersError GET_CONFIG_ERROR =
      ServersError._(1, _omitEnumNames ? '' : 'GET_CONFIG_ERROR');
  static const ServersError FILTER_SERVERS_ERROR =
      ServersError._(2, _omitEnumNames ? '' : 'FILTER_SERVERS_ERROR');

  static const $core.List<ServersError> values = <ServersError>[
    NO_ERROR,
    GET_CONFIG_ERROR,
    FILTER_SERVERS_ERROR,
  ];

  static final $core.List<ServersError?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static ServersError? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ServersError._(super.value, super.name);
}

class Technology extends $pb.ProtobufEnum {
  static const Technology UNKNOWN_TECHNLOGY =
      Technology._(0, _omitEnumNames ? '' : 'UNKNOWN_TECHNLOGY');
  static const Technology NORDLYNX =
      Technology._(1, _omitEnumNames ? '' : 'NORDLYNX');
  static const Technology OPENVPN_TCP =
      Technology._(2, _omitEnumNames ? '' : 'OPENVPN_TCP');
  static const Technology OPENVPN_UDP =
      Technology._(3, _omitEnumNames ? '' : 'OPENVPN_UDP');
  static const Technology OBFUSCATED_OPENVPN_TCP =
      Technology._(4, _omitEnumNames ? '' : 'OBFUSCATED_OPENVPN_TCP');
  static const Technology OBFUSCATED_OPENVPN_UDP =
      Technology._(5, _omitEnumNames ? '' : 'OBFUSCATED_OPENVPN_UDP');

  static const $core.List<Technology> values = <Technology>[
    UNKNOWN_TECHNLOGY,
    NORDLYNX,
    OPENVPN_TCP,
    OPENVPN_UDP,
    OBFUSCATED_OPENVPN_TCP,
    OBFUSCATED_OPENVPN_UDP,
  ];

  static final $core.List<Technology?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static Technology? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const Technology._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
