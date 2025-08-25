// This is a generated file - do not edit.
//
// Generated from service_response.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// ServiceErrorCode defines a set of error codes which handling
/// does not depend on any specific command used.
class ServiceErrorCode extends $pb.ProtobufEnum {
  static const ServiceErrorCode NOT_LOGGED_IN =
      ServiceErrorCode._(0, _omitEnumNames ? '' : 'NOT_LOGGED_IN');
  static const ServiceErrorCode API_FAILURE =
      ServiceErrorCode._(1, _omitEnumNames ? '' : 'API_FAILURE');
  static const ServiceErrorCode CONFIG_FAILURE =
      ServiceErrorCode._(2, _omitEnumNames ? '' : 'CONFIG_FAILURE');

  static const $core.List<ServiceErrorCode> values = <ServiceErrorCode>[
    NOT_LOGGED_IN,
    API_FAILURE,
    CONFIG_FAILURE,
  ];

  static final $core.List<ServiceErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static ServiceErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ServiceErrorCode._(super.value, super.name);
}

/// MeshnetErrorCode defines a set of meshnet specific error codes.
class MeshnetErrorCode extends $pb.ProtobufEnum {
  static const MeshnetErrorCode NOT_REGISTERED =
      MeshnetErrorCode._(0, _omitEnumNames ? '' : 'NOT_REGISTERED');
  static const MeshnetErrorCode LIB_FAILURE =
      MeshnetErrorCode._(1, _omitEnumNames ? '' : 'LIB_FAILURE');
  static const MeshnetErrorCode ALREADY_ENABLED =
      MeshnetErrorCode._(3, _omitEnumNames ? '' : 'ALREADY_ENABLED');
  static const MeshnetErrorCode ALREADY_DISABLED =
      MeshnetErrorCode._(4, _omitEnumNames ? '' : 'ALREADY_DISABLED');
  static const MeshnetErrorCode NOT_ENABLED =
      MeshnetErrorCode._(5, _omitEnumNames ? '' : 'NOT_ENABLED');
  static const MeshnetErrorCode TECH_FAILURE =
      MeshnetErrorCode._(6, _omitEnumNames ? '' : 'TECH_FAILURE');
  static const MeshnetErrorCode TUNNEL_CLOSED =
      MeshnetErrorCode._(7, _omitEnumNames ? '' : 'TUNNEL_CLOSED');
  static const MeshnetErrorCode CONFLICT_WITH_PQ =
      MeshnetErrorCode._(8, _omitEnumNames ? '' : 'CONFLICT_WITH_PQ');
  static const MeshnetErrorCode CONFLICT_WITH_PQ_SERVER =
      MeshnetErrorCode._(9, _omitEnumNames ? '' : 'CONFLICT_WITH_PQ_SERVER');

  static const $core.List<MeshnetErrorCode> values = <MeshnetErrorCode>[
    NOT_REGISTERED,
    LIB_FAILURE,
    ALREADY_ENABLED,
    ALREADY_DISABLED,
    NOT_ENABLED,
    TECH_FAILURE,
    TUNNEL_CLOSED,
    CONFLICT_WITH_PQ,
    CONFLICT_WITH_PQ_SERVER,
  ];

  static final $core.List<MeshnetErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 9);
  static MeshnetErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const MeshnetErrorCode._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
