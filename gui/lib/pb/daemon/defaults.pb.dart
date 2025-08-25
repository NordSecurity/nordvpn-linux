// This is a generated file - do not edit.
//
// Generated from defaults.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

class SetDefaultsRequest extends $pb.GeneratedMessage {
  factory SetDefaultsRequest({
    $core.bool? noLogout,
  }) {
    final result = create();
    if (noLogout != null) result.noLogout = noLogout;
    return result;
  }

  SetDefaultsRequest._();

  factory SetDefaultsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetDefaultsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetDefaultsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'noLogout')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetDefaultsRequest clone() => SetDefaultsRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetDefaultsRequest copyWith(void Function(SetDefaultsRequest) updates) =>
      super.copyWith((message) => updates(message as SetDefaultsRequest))
          as SetDefaultsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetDefaultsRequest create() => SetDefaultsRequest._();
  @$core.override
  SetDefaultsRequest createEmptyInstance() => create();
  static $pb.PbList<SetDefaultsRequest> createRepeated() =>
      $pb.PbList<SetDefaultsRequest>();
  @$core.pragma('dart2js:noInline')
  static SetDefaultsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetDefaultsRequest>(create);
  static SetDefaultsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get noLogout => $_getBF(0);
  @$pb.TagNumber(1)
  set noLogout($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasNoLogout() => $_has(0);
  @$pb.TagNumber(1)
  void clearNoLogout() => $_clearField(1);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
