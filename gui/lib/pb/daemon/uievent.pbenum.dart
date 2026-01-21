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
class UIEvent_FormReference extends $pb.ProtobufEnum {
  static const UIEvent_FormReference FORM_REFERENCE_UNSPECIFIED =
      UIEvent_FormReference._(
          0, _omitEnumNames ? '' : 'FORM_REFERENCE_UNSPECIFIED');
  static const UIEvent_FormReference CLI =
      UIEvent_FormReference._(1, _omitEnumNames ? '' : 'CLI');
  static const UIEvent_FormReference TRAY =
      UIEvent_FormReference._(2, _omitEnumNames ? '' : 'TRAY');
  static const UIEvent_FormReference HOME_SCREEN =
      UIEvent_FormReference._(3, _omitEnumNames ? '' : 'HOME_SCREEN');

  static const $core.List<UIEvent_FormReference> values =
      <UIEvent_FormReference>[
    FORM_REFERENCE_UNSPECIFIED,
    CLI,
    TRAY,
    HOME_SCREEN,
  ];

  static final $core.List<UIEvent_FormReference?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static UIEvent_FormReference? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const UIEvent_FormReference._(super.value, super.name);
}

/// ItemName represents a user interface event item name
class UIEvent_ItemName extends $pb.ProtobufEnum {
  static const UIEvent_ItemName ITEM_NAME_UNSPECIFIED =
      UIEvent_ItemName._(0, _omitEnumNames ? '' : 'ITEM_NAME_UNSPECIFIED');
  static const UIEvent_ItemName CONNECT =
      UIEvent_ItemName._(1, _omitEnumNames ? '' : 'CONNECT');
  static const UIEvent_ItemName CONNECT_RECENTS =
      UIEvent_ItemName._(2, _omitEnumNames ? '' : 'CONNECT_RECENTS');
  static const UIEvent_ItemName DISCONNECT =
      UIEvent_ItemName._(3, _omitEnumNames ? '' : 'DISCONNECT');
  static const UIEvent_ItemName LOGIN =
      UIEvent_ItemName._(4, _omitEnumNames ? '' : 'LOGIN');
  static const UIEvent_ItemName LOGOUT =
      UIEvent_ItemName._(5, _omitEnumNames ? '' : 'LOGOUT');
  static const UIEvent_ItemName RATE_CONNECTION =
      UIEvent_ItemName._(6, _omitEnumNames ? '' : 'RATE_CONNECTION');
  static const UIEvent_ItemName MESHNET_INVITE_SEND =
      UIEvent_ItemName._(7, _omitEnumNames ? '' : 'MESHNET_INVITE_SEND');

  static const $core.List<UIEvent_ItemName> values = <UIEvent_ItemName>[
    ITEM_NAME_UNSPECIFIED,
    CONNECT,
    CONNECT_RECENTS,
    DISCONNECT,
    LOGIN,
    LOGOUT,
    RATE_CONNECTION,
    MESHNET_INVITE_SEND,
  ];

  static final $core.List<UIEvent_ItemName?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 7);
  static UIEvent_ItemName? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const UIEvent_ItemName._(super.value, super.name);
}

/// ItemType represents the type of user interface event
class UIEvent_ItemType extends $pb.ProtobufEnum {
  static const UIEvent_ItemType ITEM_TYPE_UNSPECIFIED =
      UIEvent_ItemType._(0, _omitEnumNames ? '' : 'ITEM_TYPE_UNSPECIFIED');
  static const UIEvent_ItemType CLICK =
      UIEvent_ItemType._(1, _omitEnumNames ? '' : 'CLICK');

  static const $core.List<UIEvent_ItemType> values = <UIEvent_ItemType>[
    ITEM_TYPE_UNSPECIFIED,
    CLICK,
  ];

  static final $core.List<UIEvent_ItemType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 1);
  static UIEvent_ItemType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const UIEvent_ItemType._(super.value, super.name);
}

/// ItemValue represents the value associated with a user interface event
class UIEvent_ItemValue extends $pb.ProtobufEnum {
  static const UIEvent_ItemValue ITEM_VALUE_UNSPECIFIED =
      UIEvent_ItemValue._(0, _omitEnumNames ? '' : 'ITEM_VALUE_UNSPECIFIED');
  static const UIEvent_ItemValue COUNTRY =
      UIEvent_ItemValue._(1, _omitEnumNames ? '' : 'COUNTRY');
  static const UIEvent_ItemValue CITY =
      UIEvent_ItemValue._(2, _omitEnumNames ? '' : 'CITY');
  static const UIEvent_ItemValue DIP =
      UIEvent_ItemValue._(3, _omitEnumNames ? '' : 'DIP');
  static const UIEvent_ItemValue MESHNET =
      UIEvent_ItemValue._(4, _omitEnumNames ? '' : 'MESHNET');
  static const UIEvent_ItemValue OBFUSCATED =
      UIEvent_ItemValue._(5, _omitEnumNames ? '' : 'OBFUSCATED');
  static const UIEvent_ItemValue ONION_OVER_VPN =
      UIEvent_ItemValue._(6, _omitEnumNames ? '' : 'ONION_OVER_VPN');
  static const UIEvent_ItemValue DOUBLE_VPN =
      UIEvent_ItemValue._(7, _omitEnumNames ? '' : 'DOUBLE_VPN');
  static const UIEvent_ItemValue P2P =
      UIEvent_ItemValue._(8, _omitEnumNames ? '' : 'P2P');

  static const $core.List<UIEvent_ItemValue> values = <UIEvent_ItemValue>[
    ITEM_VALUE_UNSPECIFIED,
    COUNTRY,
    CITY,
    DIP,
    MESHNET,
    OBFUSCATED,
    ONION_OVER_VPN,
    DOUBLE_VPN,
    P2P,
  ];

  static final $core.List<UIEvent_ItemValue?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 8);
  static UIEvent_ItemValue? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const UIEvent_ItemValue._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
