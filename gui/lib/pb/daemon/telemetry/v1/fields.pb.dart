// This is a generated file - do not edit.
//
// Generated from fields.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'fields.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'fields.pbenum.dart';

class DesktopEnvironmentRequest extends $pb.GeneratedMessage {
  factory DesktopEnvironmentRequest({
    $core.String? desktopEnvName,
  }) {
    final result = create();
    if (desktopEnvName != null) result.desktopEnvName = desktopEnvName;
    return result;
  }

  DesktopEnvironmentRequest._();

  factory DesktopEnvironmentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DesktopEnvironmentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DesktopEnvironmentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'telemetry.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'desktopEnvName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DesktopEnvironmentRequest clone() =>
      DesktopEnvironmentRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DesktopEnvironmentRequest copyWith(
          void Function(DesktopEnvironmentRequest) updates) =>
      super.copyWith((message) => updates(message as DesktopEnvironmentRequest))
          as DesktopEnvironmentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DesktopEnvironmentRequest create() => DesktopEnvironmentRequest._();
  @$core.override
  DesktopEnvironmentRequest createEmptyInstance() => create();
  static $pb.PbList<DesktopEnvironmentRequest> createRepeated() =>
      $pb.PbList<DesktopEnvironmentRequest>();
  @$core.pragma('dart2js:noInline')
  static DesktopEnvironmentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DesktopEnvironmentRequest>(create);
  static DesktopEnvironmentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get desktopEnvName => $_getSZ(0);
  @$pb.TagNumber(1)
  set desktopEnvName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDesktopEnvName() => $_has(0);
  @$pb.TagNumber(1)
  void clearDesktopEnvName() => $_clearField(1);
}

class DisplayProtocolRequest extends $pb.GeneratedMessage {
  factory DisplayProtocolRequest({
    DisplayProtocol? protocol,
  }) {
    final result = create();
    if (protocol != null) result.protocol = protocol;
    return result;
  }

  DisplayProtocolRequest._();

  factory DisplayProtocolRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DisplayProtocolRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DisplayProtocolRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'telemetry.v1'),
      createEmptyInstance: create)
    ..e<DisplayProtocol>(
        1, _omitFieldNames ? '' : 'protocol', $pb.PbFieldType.OE,
        defaultOrMaker: DisplayProtocol.DISPLAY_PROTOCOL_UNSPECIFIED,
        valueOf: DisplayProtocol.valueOf,
        enumValues: DisplayProtocol.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisplayProtocolRequest clone() =>
      DisplayProtocolRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisplayProtocolRequest copyWith(
          void Function(DisplayProtocolRequest) updates) =>
      super.copyWith((message) => updates(message as DisplayProtocolRequest))
          as DisplayProtocolRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DisplayProtocolRequest create() => DisplayProtocolRequest._();
  @$core.override
  DisplayProtocolRequest createEmptyInstance() => create();
  static $pb.PbList<DisplayProtocolRequest> createRepeated() =>
      $pb.PbList<DisplayProtocolRequest>();
  @$core.pragma('dart2js:noInline')
  static DisplayProtocolRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DisplayProtocolRequest>(create);
  static DisplayProtocolRequest? _defaultInstance;

  @$pb.TagNumber(1)
  DisplayProtocol get protocol => $_getN(0);
  @$pb.TagNumber(1)
  set protocol(DisplayProtocol value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProtocol() => $_has(0);
  @$pb.TagNumber(1)
  void clearProtocol() => $_clearField(1);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
