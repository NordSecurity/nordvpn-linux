// This is a generated file - do not edit.
//
// Generated from state.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class AppStateError extends $pb.ProtobufEnum {
  static const AppStateError FAILED_TO_GET_UID =
      AppStateError._(0, _omitEnumNames ? '' : 'FAILED_TO_GET_UID');

  static const $core.List<AppStateError> values = <AppStateError>[
    FAILED_TO_GET_UID,
  ];

  static final $core.List<AppStateError?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 0);
  static AppStateError? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AppStateError._(super.value, super.name);
}

class LoginEventType extends $pb.ProtobufEnum {
  static const LoginEventType LOGIN =
      LoginEventType._(0, _omitEnumNames ? '' : 'LOGIN');
  static const LoginEventType LOGOUT =
      LoginEventType._(1, _omitEnumNames ? '' : 'LOGOUT');

  static const $core.List<LoginEventType> values = <LoginEventType>[
    LOGIN,
    LOGOUT,
  ];

  static final $core.List<LoginEventType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 1);
  static LoginEventType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const LoginEventType._(super.value, super.name);
}

class UpdateEvent extends $pb.ProtobufEnum {
  static const UpdateEvent SERVERS_LIST_UPDATE =
      UpdateEvent._(0, _omitEnumNames ? '' : 'SERVERS_LIST_UPDATE');

  static const $core.List<UpdateEvent> values = <UpdateEvent>[
    SERVERS_LIST_UPDATE,
  ];

  static final $core.List<UpdateEvent?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 0);
  static UpdateEvent? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const UpdateEvent._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
