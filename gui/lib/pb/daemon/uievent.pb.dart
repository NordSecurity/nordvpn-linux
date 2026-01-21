// This is a generated file - do not edit.
//
// Generated from uievent.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'uievent.pbenum.dart';

/// UIEvent contains nested enums for UI event tracking
class UIEvent extends $pb.GeneratedMessage {
  factory UIEvent() => create();

  UIEvent._();

  factory UIEvent.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UIEvent.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UIEvent',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UIEvent clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UIEvent copyWith(void Function(UIEvent) updates) =>
      super.copyWith((message) => updates(message as UIEvent)) as UIEvent;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UIEvent create() => UIEvent._();
  @$core.override
  UIEvent createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static UIEvent getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UIEvent>(create);
  static UIEvent? _defaultInstance;
}

const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
