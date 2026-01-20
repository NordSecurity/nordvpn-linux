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

/// UIFormReference identifies the UI form or screen from which an action
/// originated
class UIFormReference extends $pb.ProtobufEnum {
  static const UIFormReference UI_FORM_REFERENCE_UNSPECIFIED =
      UIFormReference._(
          0, _omitEnumNames ? '' : 'UI_FORM_REFERENCE_UNSPECIFIED');
  static const UIFormReference UI_FORM_REFERENCE_CLI =
      UIFormReference._(1, _omitEnumNames ? '' : 'UI_FORM_REFERENCE_CLI');
  static const UIFormReference UI_FORM_REFERENCE_TRAY =
      UIFormReference._(2, _omitEnumNames ? '' : 'UI_FORM_REFERENCE_TRAY');
  static const UIFormReference UI_FORM_REFERENCE_HOME_SCREEN =
      UIFormReference._(
          3, _omitEnumNames ? '' : 'UI_FORM_REFERENCE_HOME_SCREEN');

  static const $core.List<UIFormReference> values = <UIFormReference>[
    UI_FORM_REFERENCE_UNSPECIFIED,
    UI_FORM_REFERENCE_CLI,
    UI_FORM_REFERENCE_TRAY,
    UI_FORM_REFERENCE_HOME_SCREEN,
  ];

  static final $core.List<UIFormReference?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static UIFormReference? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const UIFormReference._(super.value, super.name);
}

/// UIItemName represents a user interface event item name
class UIItemName extends $pb.ProtobufEnum {
  static const UIItemName UI_ITEM_NAME_UNSPECIFIED =
      UIItemName._(0, _omitEnumNames ? '' : 'UI_ITEM_NAME_UNSPECIFIED');
  static const UIItemName UI_ITEM_NAME_CONNECT =
      UIItemName._(1, _omitEnumNames ? '' : 'UI_ITEM_NAME_CONNECT');
  static const UIItemName UI_ITEM_NAME_CONNECT_RECENTS =
      UIItemName._(2, _omitEnumNames ? '' : 'UI_ITEM_NAME_CONNECT_RECENTS');
  static const UIItemName UI_ITEM_NAME_DISCONNECT =
      UIItemName._(3, _omitEnumNames ? '' : 'UI_ITEM_NAME_DISCONNECT');
  static const UIItemName UI_ITEM_NAME_LOGIN =
      UIItemName._(4, _omitEnumNames ? '' : 'UI_ITEM_NAME_LOGIN');
  static const UIItemName UI_ITEM_NAME_LOGOUT =
      UIItemName._(5, _omitEnumNames ? '' : 'UI_ITEM_NAME_LOGOUT');
  static const UIItemName UI_ITEM_NAME_RATE_CONNECTION =
      UIItemName._(6, _omitEnumNames ? '' : 'UI_ITEM_NAME_RATE_CONNECTION');
  static const UIItemName UI_ITEM_NAME_MESHNET_INVITE_SEND =
      UIItemName._(7, _omitEnumNames ? '' : 'UI_ITEM_NAME_MESHNET_INVITE_SEND');

  static const $core.List<UIItemName> values = <UIItemName>[
    UI_ITEM_NAME_UNSPECIFIED,
    UI_ITEM_NAME_CONNECT,
    UI_ITEM_NAME_CONNECT_RECENTS,
    UI_ITEM_NAME_DISCONNECT,
    UI_ITEM_NAME_LOGIN,
    UI_ITEM_NAME_LOGOUT,
    UI_ITEM_NAME_RATE_CONNECTION,
    UI_ITEM_NAME_MESHNET_INVITE_SEND,
  ];

  static final $core.List<UIItemName?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 7);
  static UIItemName? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const UIItemName._(super.value, super.name);
}

/// UIItemType represents the type of user interface event
class UIItemType extends $pb.ProtobufEnum {
  static const UIItemType UI_ITEM_TYPE_UNSPECIFIED =
      UIItemType._(0, _omitEnumNames ? '' : 'UI_ITEM_TYPE_UNSPECIFIED');
  static const UIItemType UI_ITEM_TYPE_CLICK =
      UIItemType._(1, _omitEnumNames ? '' : 'UI_ITEM_TYPE_CLICK');
  static const UIItemType UI_ITEM_TYPE_SHOW =
      UIItemType._(2, _omitEnumNames ? '' : 'UI_ITEM_TYPE_SHOW');

  static const $core.List<UIItemType> values = <UIItemType>[
    UI_ITEM_TYPE_UNSPECIFIED,
    UI_ITEM_TYPE_CLICK,
    UI_ITEM_TYPE_SHOW,
  ];

  static final $core.List<UIItemType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static UIItemType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const UIItemType._(super.value, super.name);
}

/// UIItemValue represents the value associated with a user interface event
class UIItemValue extends $pb.ProtobufEnum {
  static const UIItemValue UI_ITEM_VALUE_CONNECTION_UNSPECIFIED = UIItemValue._(
      0, _omitEnumNames ? '' : 'UI_ITEM_VALUE_CONNECTION_UNSPECIFIED');
  static const UIItemValue UI_ITEM_VALUE_CONNECTION_COUNTRY = UIItemValue._(
      1, _omitEnumNames ? '' : 'UI_ITEM_VALUE_CONNECTION_COUNTRY');
  static const UIItemValue UI_ITEM_VALUE_CONNECTION_CITY =
      UIItemValue._(2, _omitEnumNames ? '' : 'UI_ITEM_VALUE_CONNECTION_CITY');
  static const UIItemValue UI_ITEM_VALUE_CONNECTION_DIP =
      UIItemValue._(3, _omitEnumNames ? '' : 'UI_ITEM_VALUE_CONNECTION_DIP');
  static const UIItemValue UI_ITEM_VALUE_CONNECTION_MESHNET = UIItemValue._(
      4, _omitEnumNames ? '' : 'UI_ITEM_VALUE_CONNECTION_MESHNET');
  static const UIItemValue UI_ITEM_VALUE_CONNECTION_OBFUSCATED = UIItemValue._(
      5, _omitEnumNames ? '' : 'UI_ITEM_VALUE_CONNECTION_OBFUSCATED');
  static const UIItemValue UI_ITEM_VALUE_CONNECTION_ONION_OVER_VPN =
      UIItemValue._(
          6, _omitEnumNames ? '' : 'UI_ITEM_VALUE_CONNECTION_ONION_OVER_VPN');
  static const UIItemValue UI_ITEM_VALUE_CONNECTION_DOUBLE_VPN = UIItemValue._(
      7, _omitEnumNames ? '' : 'UI_ITEM_VALUE_CONNECTION_DOUBLE_VPN');
  static const UIItemValue UI_ITEM_VALUE_CONNECTION_P2P =
      UIItemValue._(8, _omitEnumNames ? '' : 'UI_ITEM_VALUE_CONNECTION_P2P');

  static const $core.List<UIItemValue> values = <UIItemValue>[
    UI_ITEM_VALUE_CONNECTION_UNSPECIFIED,
    UI_ITEM_VALUE_CONNECTION_COUNTRY,
    UI_ITEM_VALUE_CONNECTION_CITY,
    UI_ITEM_VALUE_CONNECTION_DIP,
    UI_ITEM_VALUE_CONNECTION_MESHNET,
    UI_ITEM_VALUE_CONNECTION_OBFUSCATED,
    UI_ITEM_VALUE_CONNECTION_ONION_OVER_VPN,
    UI_ITEM_VALUE_CONNECTION_DOUBLE_VPN,
    UI_ITEM_VALUE_CONNECTION_P2P,
  ];

  static final $core.List<UIItemValue?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 8);
  static UIItemValue? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const UIItemValue._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
