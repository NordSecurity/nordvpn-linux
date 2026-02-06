// This is a generated file - do not edit.
//
// Generated from snapconf.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// ErrMissingConnections defines that some of the snap interfaces that are required for the gRPC
/// call to be executed are not connected.
class ErrMissingConnections extends $pb.GeneratedMessage {
  factory ErrMissingConnections({
    $core.Iterable<$core.String>? missingConnections,
  }) {
    final result = create();
    if (missingConnections != null)
      result.missingConnections.addAll(missingConnections);
    return result;
  }

  ErrMissingConnections._();

  factory ErrMissingConnections.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ErrMissingConnections.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ErrMissingConnections',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'snappb'),
      createEmptyInstance: create)
    ..pPS(1, _omitFieldNames ? '' : 'missingConnections')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ErrMissingConnections clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ErrMissingConnections copyWith(
          void Function(ErrMissingConnections) updates) =>
      super.copyWith((message) => updates(message as ErrMissingConnections))
          as ErrMissingConnections;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ErrMissingConnections create() => ErrMissingConnections._();
  @$core.override
  ErrMissingConnections createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ErrMissingConnections getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ErrMissingConnections>(create);
  static ErrMissingConnections? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get missingConnections => $_getList(0);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
