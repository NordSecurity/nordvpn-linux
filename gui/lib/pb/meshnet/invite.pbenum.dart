// This is a generated file - do not edit.
//
// Generated from invite.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// RespondToInviteErrorCode defines the error of meshnet service
/// response to the invitation response
class RespondToInviteErrorCode extends $pb.ProtobufEnum {
  /// UNKNOWN defines that the exact error was not determined
  static const RespondToInviteErrorCode UNKNOWN =
      RespondToInviteErrorCode._(0, _omitEnumNames ? '' : 'UNKNOWN');

  /// NO_SUCH_INVITATION defines that the request was not handled
  /// successfully
  static const RespondToInviteErrorCode NO_SUCH_INVITATION =
      RespondToInviteErrorCode._(1, _omitEnumNames ? '' : 'NO_SUCH_INVITATION');

  /// DEVICE_COUNT defines that no more devices can be added
  static const RespondToInviteErrorCode DEVICE_COUNT =
      RespondToInviteErrorCode._(2, _omitEnumNames ? '' : 'DEVICE_COUNT');

  static const $core.List<RespondToInviteErrorCode> values =
      <RespondToInviteErrorCode>[
    UNKNOWN,
    NO_SUCH_INVITATION,
    DEVICE_COUNT,
  ];

  static final $core.List<RespondToInviteErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static RespondToInviteErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const RespondToInviteErrorCode._(super.value, super.name);
}

/// InviteResponseCode defines a response code specific to the
/// invitation send action
class InviteResponseErrorCode extends $pb.ProtobufEnum {
  /// ALREADY_EXISTS defines that the invitation to the specified
  /// email already exists
  static const InviteResponseErrorCode ALREADY_EXISTS =
      InviteResponseErrorCode._(0, _omitEnumNames ? '' : 'ALREADY_EXISTS');

  /// INVALID_EMAIL defines that the given email is invalid,
  /// therefore, cannot receive an invitation
  static const InviteResponseErrorCode INVALID_EMAIL =
      InviteResponseErrorCode._(1, _omitEnumNames ? '' : 'INVALID_EMAIL');

  /// SAME_ACCOUNT_EMAIL defines that the given email is for the same account,
  /// cannot send invite to myself
  static const InviteResponseErrorCode SAME_ACCOUNT_EMAIL =
      InviteResponseErrorCode._(2, _omitEnumNames ? '' : 'SAME_ACCOUNT_EMAIL');

  /// LIMIT_REACHED defines that the weekly invitation limit (20)
  /// has been reached
  static const InviteResponseErrorCode LIMIT_REACHED =
      InviteResponseErrorCode._(3, _omitEnumNames ? '' : 'LIMIT_REACHED');

  /// PEER_COUNT defines that no more devices can be invited
  static const InviteResponseErrorCode PEER_COUNT =
      InviteResponseErrorCode._(4, _omitEnumNames ? '' : 'PEER_COUNT');

  static const $core.List<InviteResponseErrorCode> values =
      <InviteResponseErrorCode>[
    ALREADY_EXISTS,
    INVALID_EMAIL,
    SAME_ACCOUNT_EMAIL,
    LIMIT_REACHED,
    PEER_COUNT,
  ];

  static final $core.List<InviteResponseErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static InviteResponseErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const InviteResponseErrorCode._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
