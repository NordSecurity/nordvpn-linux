// This is a generated file - do not edit.
//
// Generated from common.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class TriState extends $pb.ProtobufEnum {
  static const TriState UNKNOWN =
      TriState._(0, _omitEnumNames ? '' : 'UNKNOWN');
  static const TriState DISABLED =
      TriState._(1, _omitEnumNames ? '' : 'DISABLED');
  static const TriState ENABLED =
      TriState._(2, _omitEnumNames ? '' : 'ENABLED');

  static const $core.List<TriState> values = <TriState>[
    UNKNOWN,
    DISABLED,
    ENABLED,
  ];

  static final $core.List<TriState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static TriState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const TriState._(super.value, super.name);
}

class ClientID extends $pb.ProtobufEnum {
  static const ClientID UNKNOWN_CLIENT =
      ClientID._(0, _omitEnumNames ? '' : 'UNKNOWN_CLIENT');
  static const ClientID CLI = ClientID._(1, _omitEnumNames ? '' : 'CLI');
  static const ClientID GUI = ClientID._(2, _omitEnumNames ? '' : 'GUI');
  static const ClientID TRAY = ClientID._(3, _omitEnumNames ? '' : 'TRAY');

  static const $core.List<ClientID> values = <ClientID>[
    UNKNOWN_CLIENT,
    CLI,
    GUI,
    TRAY,
  ];

  static final $core.List<ClientID?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static ClientID? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ClientID._(super.value, super.name);
}

class DiagnosticsErrorCode extends $pb.ProtobufEnum {
  static const DiagnosticsErrorCode DIAGNOSTICS_ERROR_CODE_UNSPECIFIED =
      DiagnosticsErrorCode._(
          0, _omitEnumNames ? '' : 'DIAGNOSTICS_ERROR_CODE_UNSPECIFIED');
  static const DiagnosticsErrorCode DIAGNOSTICS_ERROR_CODE_INTERNAL =
      DiagnosticsErrorCode._(
          1, _omitEnumNames ? '' : 'DIAGNOSTICS_ERROR_CODE_INTERNAL');
  static const DiagnosticsErrorCode
      DIAGNOSTICS_ERROR_CODE_FAILED_TO_CREATE_ZIP = DiagnosticsErrorCode._(2,
          _omitEnumNames ? '' : 'DIAGNOSTICS_ERROR_CODE_FAILED_TO_CREATE_ZIP');
  static const DiagnosticsErrorCode DIAGNOSTICS_ERROR_CODE_CHOWN_FAILED =
      DiagnosticsErrorCode._(
          3, _omitEnumNames ? '' : 'DIAGNOSTICS_ERROR_CODE_CHOWN_FAILED');
  static const DiagnosticsErrorCode DIAGNOSTICS_ERROR_CODE_ZIP_TOO_LARGE =
      DiagnosticsErrorCode._(
          4, _omitEnumNames ? '' : 'DIAGNOSTICS_ERROR_CODE_ZIP_TOO_LARGE');
  static const DiagnosticsErrorCode DIAGNOSTICS_ERROR_CODE_COLLECTION_FAILED =
      DiagnosticsErrorCode._(
          5, _omitEnumNames ? '' : 'DIAGNOSTICS_ERROR_CODE_COLLECTION_FAILED');
  static const DiagnosticsErrorCode DIAGNOSTICS_ERROR_CODE_FAILED_TO_CLOSE_ZIP =
      DiagnosticsErrorCode._(6,
          _omitEnumNames ? '' : 'DIAGNOSTICS_ERROR_CODE_FAILED_TO_CLOSE_ZIP');
  static const DiagnosticsErrorCode
      DIAGNOSTICS_ERROR_CODE_NO_DAEMON_LOG_SOURCE = DiagnosticsErrorCode._(7,
          _omitEnumNames ? '' : 'DIAGNOSTICS_ERROR_CODE_NO_DAEMON_LOG_SOURCE');

  static const $core.List<DiagnosticsErrorCode> values = <DiagnosticsErrorCode>[
    DIAGNOSTICS_ERROR_CODE_UNSPECIFIED,
    DIAGNOSTICS_ERROR_CODE_INTERNAL,
    DIAGNOSTICS_ERROR_CODE_FAILED_TO_CREATE_ZIP,
    DIAGNOSTICS_ERROR_CODE_CHOWN_FAILED,
    DIAGNOSTICS_ERROR_CODE_ZIP_TOO_LARGE,
    DIAGNOSTICS_ERROR_CODE_COLLECTION_FAILED,
    DIAGNOSTICS_ERROR_CODE_FAILED_TO_CLOSE_ZIP,
    DIAGNOSTICS_ERROR_CODE_NO_DAEMON_LOG_SOURCE,
  ];

  static final $core.List<DiagnosticsErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 7);
  static DiagnosticsErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const DiagnosticsErrorCode._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
