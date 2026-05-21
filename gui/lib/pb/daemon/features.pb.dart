// This is a generated file - do not edit.
//
// Generated from features.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

class FeatureToggles extends $pb.GeneratedMessage {
  factory FeatureToggles({
    $core.bool? meshnetEnabled,
    $core.bool? dedicatedServerEnabled,
  }) {
    final result = create();
    if (meshnetEnabled != null) result.meshnetEnabled = meshnetEnabled;
    if (dedicatedServerEnabled != null)
      result.dedicatedServerEnabled = dedicatedServerEnabled;
    return result;
  }

  FeatureToggles._();

  factory FeatureToggles.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory FeatureToggles.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'FeatureToggles',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'meshnetEnabled')
    ..aOB(2, _omitFieldNames ? '' : 'dedicatedServerEnabled')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FeatureToggles clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FeatureToggles copyWith(void Function(FeatureToggles) updates) =>
      super.copyWith((message) => updates(message as FeatureToggles))
          as FeatureToggles;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static FeatureToggles create() => FeatureToggles._();
  @$core.override
  FeatureToggles createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static FeatureToggles getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<FeatureToggles>(create);
  static FeatureToggles? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get meshnetEnabled => $_getBF(0);
  @$pb.TagNumber(1)
  set meshnetEnabled($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMeshnetEnabled() => $_has(0);
  @$pb.TagNumber(1)
  void clearMeshnetEnabled() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get dedicatedServerEnabled => $_getBF(1);
  @$pb.TagNumber(2)
  set dedicatedServerEnabled($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDedicatedServerEnabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearDedicatedServerEnabled() => $_clearField(2);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
