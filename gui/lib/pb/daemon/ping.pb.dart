// This is a generated file - do not edit.
//
// Generated from ping.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

class PingResponse extends $pb.GeneratedMessage {
  factory PingResponse({
    $fixnum.Int64? type,
    $fixnum.Int64? major,
    $fixnum.Int64? minor,
    $fixnum.Int64? patch,
    $core.String? metadata,
  }) {
    final result = create();
    if (type != null) result.type = type;
    if (major != null) result.major = major;
    if (minor != null) result.minor = minor;
    if (patch != null) result.patch = patch;
    if (metadata != null) result.metadata = metadata;
    return result;
  }

  PingResponse._();

  factory PingResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory PingResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PingResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aInt64(1, _omitFieldNames ? '' : 'type')
    ..aInt64(2, _omitFieldNames ? '' : 'major')
    ..aInt64(3, _omitFieldNames ? '' : 'minor')
    ..aInt64(4, _omitFieldNames ? '' : 'patch')
    ..aOS(5, _omitFieldNames ? '' : 'metadata')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PingResponse clone() => PingResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PingResponse copyWith(void Function(PingResponse) updates) =>
      super.copyWith((message) => updates(message as PingResponse))
          as PingResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PingResponse create() => PingResponse._();
  @$core.override
  PingResponse createEmptyInstance() => create();
  static $pb.PbList<PingResponse> createRepeated() =>
      $pb.PbList<PingResponse>();
  @$core.pragma('dart2js:noInline')
  static PingResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PingResponse>(create);
  static PingResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get type => $_getI64(0);
  @$pb.TagNumber(1)
  set type($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get major => $_getI64(1);
  @$pb.TagNumber(2)
  set major($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMajor() => $_has(1);
  @$pb.TagNumber(2)
  void clearMajor() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get minor => $_getI64(2);
  @$pb.TagNumber(3)
  set minor($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMinor() => $_has(2);
  @$pb.TagNumber(3)
  void clearMinor() => $_clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get patch => $_getI64(3);
  @$pb.TagNumber(4)
  set patch($fixnum.Int64 value) => $_setInt64(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPatch() => $_has(3);
  @$pb.TagNumber(4)
  void clearPatch() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get metadata => $_getSZ(4);
  @$pb.TagNumber(5)
  set metadata($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasMetadata() => $_has(4);
  @$pb.TagNumber(5)
  void clearMetadata() => $_clearField(5);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
