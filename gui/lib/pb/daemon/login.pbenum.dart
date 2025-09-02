// This is a generated file - do not edit.
//
// Generated from login.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class LoginType extends $pb.ProtobufEnum {
  static const LoginType LoginType_UNKNOWN =
      LoginType._(0, _omitEnumNames ? '' : 'LoginType_UNKNOWN');
  static const LoginType LoginType_LOGIN =
      LoginType._(1, _omitEnumNames ? '' : 'LoginType_LOGIN');
  static const LoginType LoginType_SIGNUP =
      LoginType._(2, _omitEnumNames ? '' : 'LoginType_SIGNUP');

  static const $core.List<LoginType> values = <LoginType>[
    LoginType_UNKNOWN,
    LoginType_LOGIN,
    LoginType_SIGNUP,
  ];

  static final $core.List<LoginType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static LoginType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const LoginType._(super.value, super.name);
}

class LoginStatus extends $pb.ProtobufEnum {
  static const LoginStatus SUCCESS =
      LoginStatus._(0, _omitEnumNames ? '' : 'SUCCESS');
  static const LoginStatus UNKNOWN_OAUTH2_ERROR =
      LoginStatus._(1, _omitEnumNames ? '' : 'UNKNOWN_OAUTH2_ERROR');
  static const LoginStatus ALREADY_LOGGED_IN =
      LoginStatus._(2, _omitEnumNames ? '' : 'ALREADY_LOGGED_IN');
  static const LoginStatus NO_NET =
      LoginStatus._(3, _omitEnumNames ? '' : 'NO_NET');
  static const LoginStatus CONSENT_MISSING =
      LoginStatus._(4, _omitEnumNames ? '' : 'CONSENT_MISSING');

  static const $core.List<LoginStatus> values = <LoginStatus>[
    SUCCESS,
    UNKNOWN_OAUTH2_ERROR,
    ALREADY_LOGGED_IN,
    NO_NET,
    CONSENT_MISSING,
  ];

  static final $core.List<LoginStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static LoginStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const LoginStatus._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
