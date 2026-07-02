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

import 'uievent.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'uievent.pbenum.dart';

/// UIEvent contains nested enums for UI event tracking
class UIEvent extends $pb.GeneratedMessage {
  factory UIEvent({
    UIEvent_FormReference? formReference,
    UIEvent_ItemName? itemName,
    UIEvent_ItemType? itemType,
    UIEvent_ItemValue? itemValue,
  }) {
    final result = create();
    if (formReference != null) result.formReference = formReference;
    if (itemName != null) result.itemName = itemName;
    if (itemType != null) result.itemType = itemType;
    if (itemValue != null) result.itemValue = itemValue;
    return result;
  }

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
    ..aE<UIEvent_FormReference>(1, _omitFieldNames ? '' : 'formReference',
        enumValues: UIEvent_FormReference.values)
    ..aE<UIEvent_ItemName>(2, _omitFieldNames ? '' : 'itemName',
        enumValues: UIEvent_ItemName.values)
    ..aE<UIEvent_ItemType>(3, _omitFieldNames ? '' : 'itemType',
        enumValues: UIEvent_ItemType.values)
    ..aE<UIEvent_ItemValue>(4, _omitFieldNames ? '' : 'itemValue',
        enumValues: UIEvent_ItemValue.values)
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

  @$pb.TagNumber(1)
  UIEvent_FormReference get formReference => $_getN(0);
  @$pb.TagNumber(1)
  set formReference(UIEvent_FormReference value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasFormReference() => $_has(0);
  @$pb.TagNumber(1)
  void clearFormReference() => $_clearField(1);

  @$pb.TagNumber(2)
  UIEvent_ItemName get itemName => $_getN(1);
  @$pb.TagNumber(2)
  set itemName(UIEvent_ItemName value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasItemName() => $_has(1);
  @$pb.TagNumber(2)
  void clearItemName() => $_clearField(2);

  @$pb.TagNumber(3)
  UIEvent_ItemType get itemType => $_getN(2);
  @$pb.TagNumber(3)
  set itemType(UIEvent_ItemType value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasItemType() => $_has(2);
  @$pb.TagNumber(3)
  void clearItemType() => $_clearField(3);

  @$pb.TagNumber(4)
  UIEvent_ItemValue get itemValue => $_getN(3);
  @$pb.TagNumber(4)
  set itemValue(UIEvent_ItemValue value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasItemValue() => $_has(3);
  @$pb.TagNumber(4)
  void clearItemValue() => $_clearField(4);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
