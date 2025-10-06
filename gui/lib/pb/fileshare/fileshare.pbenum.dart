// This is a generated file - do not edit.
//
// Generated from fileshare.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// ServiceErrorCode defines a set of error codes whose handling
/// does not depend on any specific command used.
class ServiceErrorCode extends $pb.ProtobufEnum {
  static const ServiceErrorCode MESH_NOT_ENABLED =
      ServiceErrorCode._(0, _omitEnumNames ? '' : 'MESH_NOT_ENABLED');
  static const ServiceErrorCode INTERNAL_FAILURE =
      ServiceErrorCode._(1, _omitEnumNames ? '' : 'INTERNAL_FAILURE');

  static const $core.List<ServiceErrorCode> values = <ServiceErrorCode>[
    MESH_NOT_ENABLED,
    INTERNAL_FAILURE,
  ];

  static final $core.List<ServiceErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 1);
  static ServiceErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ServiceErrorCode._(super.value, super.name);
}

/// FileshareErrorCode defines a set of fileshare specific error codes.
class FileshareErrorCode extends $pb.ProtobufEnum {
  static const FileshareErrorCode LIB_FAILURE =
      FileshareErrorCode._(0, _omitEnumNames ? '' : 'LIB_FAILURE');
  static const FileshareErrorCode TRANSFER_NOT_FOUND =
      FileshareErrorCode._(1, _omitEnumNames ? '' : 'TRANSFER_NOT_FOUND');
  static const FileshareErrorCode INVALID_PEER =
      FileshareErrorCode._(2, _omitEnumNames ? '' : 'INVALID_PEER');
  static const FileshareErrorCode FILE_NOT_FOUND =
      FileshareErrorCode._(3, _omitEnumNames ? '' : 'FILE_NOT_FOUND');
  static const FileshareErrorCode ACCEPT_ALL_FILES_FAILED =
      FileshareErrorCode._(5, _omitEnumNames ? '' : 'ACCEPT_ALL_FILES_FAILED');
  static const FileshareErrorCode ACCEPT_OUTGOING =
      FileshareErrorCode._(6, _omitEnumNames ? '' : 'ACCEPT_OUTGOING');
  static const FileshareErrorCode ALREADY_ACCEPTED =
      FileshareErrorCode._(7, _omitEnumNames ? '' : 'ALREADY_ACCEPTED');
  static const FileshareErrorCode FILE_INVALIDATED =
      FileshareErrorCode._(8, _omitEnumNames ? '' : 'FILE_INVALIDATED');
  static const FileshareErrorCode TRANSFER_INVALIDATED =
      FileshareErrorCode._(9, _omitEnumNames ? '' : 'TRANSFER_INVALIDATED');
  static const FileshareErrorCode TOO_MANY_FILES =
      FileshareErrorCode._(10, _omitEnumNames ? '' : 'TOO_MANY_FILES');
  static const FileshareErrorCode DIRECTORY_TOO_DEEP =
      FileshareErrorCode._(11, _omitEnumNames ? '' : 'DIRECTORY_TOO_DEEP');
  static const FileshareErrorCode SENDING_NOT_ALLOWED =
      FileshareErrorCode._(12, _omitEnumNames ? '' : 'SENDING_NOT_ALLOWED');
  static const FileshareErrorCode PEER_DISCONNECTED =
      FileshareErrorCode._(13, _omitEnumNames ? '' : 'PEER_DISCONNECTED');
  static const FileshareErrorCode FILE_NOT_IN_PROGRESS =
      FileshareErrorCode._(14, _omitEnumNames ? '' : 'FILE_NOT_IN_PROGRESS');
  static const FileshareErrorCode TRANSFER_NOT_CREATED =
      FileshareErrorCode._(15, _omitEnumNames ? '' : 'TRANSFER_NOT_CREATED');
  static const FileshareErrorCode NOT_ENOUGH_SPACE =
      FileshareErrorCode._(16, _omitEnumNames ? '' : 'NOT_ENOUGH_SPACE');
  static const FileshareErrorCode ACCEPT_DIR_NOT_FOUND =
      FileshareErrorCode._(17, _omitEnumNames ? '' : 'ACCEPT_DIR_NOT_FOUND');
  static const FileshareErrorCode ACCEPT_DIR_IS_A_SYMLINK =
      FileshareErrorCode._(18, _omitEnumNames ? '' : 'ACCEPT_DIR_IS_A_SYMLINK');
  static const FileshareErrorCode ACCEPT_DIR_IS_NOT_A_DIRECTORY =
      FileshareErrorCode._(
          19, _omitEnumNames ? '' : 'ACCEPT_DIR_IS_NOT_A_DIRECTORY');
  static const FileshareErrorCode NO_FILES =
      FileshareErrorCode._(20, _omitEnumNames ? '' : 'NO_FILES');
  static const FileshareErrorCode ACCEPT_DIR_NO_PERMISSIONS =
      FileshareErrorCode._(
          21, _omitEnumNames ? '' : 'ACCEPT_DIR_NO_PERMISSIONS');
  static const FileshareErrorCode PURGE_FAILURE =
      FileshareErrorCode._(22, _omitEnumNames ? '' : 'PURGE_FAILURE');

  static const $core.List<FileshareErrorCode> values = <FileshareErrorCode>[
    LIB_FAILURE,
    TRANSFER_NOT_FOUND,
    INVALID_PEER,
    FILE_NOT_FOUND,
    ACCEPT_ALL_FILES_FAILED,
    ACCEPT_OUTGOING,
    ALREADY_ACCEPTED,
    FILE_INVALIDATED,
    TRANSFER_INVALIDATED,
    TOO_MANY_FILES,
    DIRECTORY_TOO_DEEP,
    SENDING_NOT_ALLOWED,
    PEER_DISCONNECTED,
    FILE_NOT_IN_PROGRESS,
    TRANSFER_NOT_CREATED,
    NOT_ENOUGH_SPACE,
    ACCEPT_DIR_NOT_FOUND,
    ACCEPT_DIR_IS_A_SYMLINK,
    ACCEPT_DIR_IS_NOT_A_DIRECTORY,
    NO_FILES,
    ACCEPT_DIR_NO_PERMISSIONS,
    PURGE_FAILURE,
  ];

  static final $core.List<FileshareErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 22);
  static FileshareErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const FileshareErrorCode._(super.value, super.name);
}

class SetNotificationsStatus extends $pb.ProtobufEnum {
  static const SetNotificationsStatus SET_SUCCESS =
      SetNotificationsStatus._(0, _omitEnumNames ? '' : 'SET_SUCCESS');
  static const SetNotificationsStatus NOTHING_TO_DO =
      SetNotificationsStatus._(1, _omitEnumNames ? '' : 'NOTHING_TO_DO');
  static const SetNotificationsStatus SET_FAILURE =
      SetNotificationsStatus._(2, _omitEnumNames ? '' : 'SET_FAILURE');

  static const $core.List<SetNotificationsStatus> values =
      <SetNotificationsStatus>[
    SET_SUCCESS,
    NOTHING_TO_DO,
    SET_FAILURE,
  ];

  static final $core.List<SetNotificationsStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static SetNotificationsStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const SetNotificationsStatus._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
