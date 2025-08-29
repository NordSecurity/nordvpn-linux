// This is a generated file - do not edit.
//
// Generated from transfer.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class Direction extends $pb.ProtobufEnum {
  static const Direction UNKNOWN_DIRECTION =
      Direction._(0, _omitEnumNames ? '' : 'UNKNOWN_DIRECTION');
  static const Direction INCOMING =
      Direction._(1, _omitEnumNames ? '' : 'INCOMING');
  static const Direction OUTGOING =
      Direction._(2, _omitEnumNames ? '' : 'OUTGOING');

  static const $core.List<Direction> values = <Direction>[
    UNKNOWN_DIRECTION,
    INCOMING,
    OUTGOING,
  ];

  static final $core.List<Direction?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static Direction? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const Direction._(super.value, super.name);
}

class Status extends $pb.ProtobufEnum {
  /// Libdrop statuses for finished transfers
  static const Status SUCCESS = Status._(0, _omitEnumNames ? '' : 'SUCCESS');
  static const Status CANCELED = Status._(1, _omitEnumNames ? '' : 'CANCELED');
  static const Status BAD_PATH = Status._(2, _omitEnumNames ? '' : 'BAD_PATH');
  static const Status BAD_FILE = Status._(3, _omitEnumNames ? '' : 'BAD_FILE');
  static const Status TRANSPORT =
      Status._(4, _omitEnumNames ? '' : 'TRANSPORT');
  static const Status BAD_STATUS =
      Status._(5, _omitEnumNames ? '' : 'BAD_STATUS');
  static const Status SERVICE_STOP =
      Status._(6, _omitEnumNames ? '' : 'SERVICE_STOP');
  static const Status BAD_TRANSFER =
      Status._(7, _omitEnumNames ? '' : 'BAD_TRANSFER');
  static const Status BAD_TRANSFER_STATE =
      Status._(8, _omitEnumNames ? '' : 'BAD_TRANSFER_STATE');
  static const Status BAD_FILE_ID =
      Status._(9, _omitEnumNames ? '' : 'BAD_FILE_ID');
  static const Status BAD_SYSTEM_TIME =
      Status._(10, _omitEnumNames ? '' : 'BAD_SYSTEM_TIME');
  static const Status TRUNCATED_FILE =
      Status._(11, _omitEnumNames ? '' : 'TRUNCATED_FILE');
  static const Status EVENT_SEND =
      Status._(12, _omitEnumNames ? '' : 'EVENT_SEND');
  static const Status BAD_UUID = Status._(13, _omitEnumNames ? '' : 'BAD_UUID');
  static const Status CHANNEL_CLOSED =
      Status._(14, _omitEnumNames ? '' : 'CHANNEL_CLOSED');
  static const Status IO = Status._(15, _omitEnumNames ? '' : 'IO');
  static const Status DATA_SEND =
      Status._(16, _omitEnumNames ? '' : 'DATA_SEND');
  static const Status DIRECTORY_NOT_EXPECTED =
      Status._(17, _omitEnumNames ? '' : 'DIRECTORY_NOT_EXPECTED');
  static const Status EMPTY_TRANSFER =
      Status._(18, _omitEnumNames ? '' : 'EMPTY_TRANSFER');
  static const Status TRANSFER_CLOSED_BY_PEER =
      Status._(19, _omitEnumNames ? '' : 'TRANSFER_CLOSED_BY_PEER');
  static const Status TRANSFER_LIMITS_EXCEEDED =
      Status._(20, _omitEnumNames ? '' : 'TRANSFER_LIMITS_EXCEEDED');
  static const Status MISMATCHED_SIZE =
      Status._(21, _omitEnumNames ? '' : 'MISMATCHED_SIZE');
  static const Status UNEXPECTED_DATA =
      Status._(22, _omitEnumNames ? '' : 'UNEXPECTED_DATA');
  static const Status INVALID_ARGUMENT =
      Status._(23, _omitEnumNames ? '' : 'INVALID_ARGUMENT');
  static const Status TRANSFER_TIMEOUT =
      Status._(24, _omitEnumNames ? '' : 'TRANSFER_TIMEOUT');
  static const Status WS_SERVER =
      Status._(25, _omitEnumNames ? '' : 'WS_SERVER');
  static const Status WS_CLIENT =
      Status._(26, _omitEnumNames ? '' : 'WS_CLIENT');

  /// UNUSED = 27;
  static const Status FILE_MODIFIED =
      Status._(28, _omitEnumNames ? '' : 'FILE_MODIFIED');
  static const Status FILENAME_TOO_LONG =
      Status._(29, _omitEnumNames ? '' : 'FILENAME_TOO_LONG');
  static const Status AUTHENTICATION_FAILED =
      Status._(30, _omitEnumNames ? '' : 'AUTHENTICATION_FAILED');
  static const Status FILE_CHECKSUM_MISMATCH =
      Status._(33, _omitEnumNames ? '' : 'FILE_CHECKSUM_MISMATCH');
  static const Status FILE_REJECTED =
      Status._(34, _omitEnumNames ? '' : 'FILE_REJECTED');

  /// Internally defined statuses for unfinished transfers
  static const Status REQUESTED =
      Status._(100, _omitEnumNames ? '' : 'REQUESTED');
  static const Status ONGOING = Status._(101, _omitEnumNames ? '' : 'ONGOING');
  static const Status FINISHED_WITH_ERRORS =
      Status._(102, _omitEnumNames ? '' : 'FINISHED_WITH_ERRORS');
  static const Status ACCEPT_FAILURE =
      Status._(103, _omitEnumNames ? '' : 'ACCEPT_FAILURE');
  static const Status CANCELED_BY_PEER =
      Status._(104, _omitEnumNames ? '' : 'CANCELED_BY_PEER');
  static const Status INTERRUPTED =
      Status._(105, _omitEnumNames ? '' : 'INTERRUPTED');
  static const Status PAUSED = Status._(106, _omitEnumNames ? '' : 'PAUSED');
  static const Status PENDING = Status._(107, _omitEnumNames ? '' : 'PENDING');

  static const $core.List<Status> values = <Status>[
    SUCCESS,
    CANCELED,
    BAD_PATH,
    BAD_FILE,
    TRANSPORT,
    BAD_STATUS,
    SERVICE_STOP,
    BAD_TRANSFER,
    BAD_TRANSFER_STATE,
    BAD_FILE_ID,
    BAD_SYSTEM_TIME,
    TRUNCATED_FILE,
    EVENT_SEND,
    BAD_UUID,
    CHANNEL_CLOSED,
    IO,
    DATA_SEND,
    DIRECTORY_NOT_EXPECTED,
    EMPTY_TRANSFER,
    TRANSFER_CLOSED_BY_PEER,
    TRANSFER_LIMITS_EXCEEDED,
    MISMATCHED_SIZE,
    UNEXPECTED_DATA,
    INVALID_ARGUMENT,
    TRANSFER_TIMEOUT,
    WS_SERVER,
    WS_CLIENT,
    FILE_MODIFIED,
    FILENAME_TOO_LONG,
    AUTHENTICATION_FAILED,
    FILE_CHECKSUM_MISMATCH,
    FILE_REJECTED,
    REQUESTED,
    ONGOING,
    FINISHED_WITH_ERRORS,
    ACCEPT_FAILURE,
    CANCELED_BY_PEER,
    INTERRUPTED,
    PAUSED,
    PENDING,
  ];

  static final $core.Map<$core.int, Status> _byValue =
      $pb.ProtobufEnum.initByValue(values);
  static Status? valueOf($core.int value) => _byValue[value];

  const Status._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
