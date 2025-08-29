// This is a generated file - do not edit.
//
// Generated from set.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class SetErrorCode extends $pb.ProtobufEnum {
  static const SetErrorCode FAILURE =
      SetErrorCode._(0, _omitEnumNames ? '' : 'FAILURE');
  static const SetErrorCode CONFIG_ERROR =
      SetErrorCode._(1, _omitEnumNames ? '' : 'CONFIG_ERROR');
  static const SetErrorCode ALREADY_SET =
      SetErrorCode._(2, _omitEnumNames ? '' : 'ALREADY_SET');

  static const $core.List<SetErrorCode> values = <SetErrorCode>[
    FAILURE,
    CONFIG_ERROR,
    ALREADY_SET,
  ];

  static final $core.List<SetErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static SetErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const SetErrorCode._(super.value, super.name);
}

class SetThreatProtectionLiteStatus extends $pb.ProtobufEnum {
  static const SetThreatProtectionLiteStatus TPL_CONFIGURED =
      SetThreatProtectionLiteStatus._(
          0, _omitEnumNames ? '' : 'TPL_CONFIGURED');
  static const SetThreatProtectionLiteStatus TPL_CONFIGURED_DNS_RESET =
      SetThreatProtectionLiteStatus._(
          1, _omitEnumNames ? '' : 'TPL_CONFIGURED_DNS_RESET');

  static const $core.List<SetThreatProtectionLiteStatus> values =
      <SetThreatProtectionLiteStatus>[
    TPL_CONFIGURED,
    TPL_CONFIGURED_DNS_RESET,
  ];

  static final $core.List<SetThreatProtectionLiteStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 1);
  static SetThreatProtectionLiteStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const SetThreatProtectionLiteStatus._(super.value, super.name);
}

class SetDNSStatus extends $pb.ProtobufEnum {
  static const SetDNSStatus DNS_CONFIGURED =
      SetDNSStatus._(0, _omitEnumNames ? '' : 'DNS_CONFIGURED');
  static const SetDNSStatus DNS_CONFIGURED_TPL_RESET =
      SetDNSStatus._(1, _omitEnumNames ? '' : 'DNS_CONFIGURED_TPL_RESET');
  static const SetDNSStatus INVALID_DNS_ADDRESS =
      SetDNSStatus._(2, _omitEnumNames ? '' : 'INVALID_DNS_ADDRESS');
  static const SetDNSStatus TOO_MANY_VALUES =
      SetDNSStatus._(3, _omitEnumNames ? '' : 'TOO_MANY_VALUES');

  static const $core.List<SetDNSStatus> values = <SetDNSStatus>[
    DNS_CONFIGURED,
    DNS_CONFIGURED_TPL_RESET,
    INVALID_DNS_ADDRESS,
    TOO_MANY_VALUES,
  ];

  static final $core.List<SetDNSStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static SetDNSStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const SetDNSStatus._(super.value, super.name);
}

class SetProtocolStatus extends $pb.ProtobufEnum {
  static const SetProtocolStatus PROTOCOL_CONFIGURED =
      SetProtocolStatus._(0, _omitEnumNames ? '' : 'PROTOCOL_CONFIGURED');
  static const SetProtocolStatus PROTOCOL_CONFIGURED_VPN_ON =
      SetProtocolStatus._(
          1, _omitEnumNames ? '' : 'PROTOCOL_CONFIGURED_VPN_ON');
  static const SetProtocolStatus INVALID_TECHNOLOGY =
      SetProtocolStatus._(2, _omitEnumNames ? '' : 'INVALID_TECHNOLOGY');

  static const $core.List<SetProtocolStatus> values = <SetProtocolStatus>[
    PROTOCOL_CONFIGURED,
    PROTOCOL_CONFIGURED_VPN_ON,
    INVALID_TECHNOLOGY,
  ];

  static final $core.List<SetProtocolStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static SetProtocolStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const SetProtocolStatus._(super.value, super.name);
}

class SetLANDiscoveryStatus extends $pb.ProtobufEnum {
  static const SetLANDiscoveryStatus DISCOVERY_CONFIGURED =
      SetLANDiscoveryStatus._(0, _omitEnumNames ? '' : 'DISCOVERY_CONFIGURED');
  static const SetLANDiscoveryStatus DISCOVERY_CONFIGURED_ALLOWLIST_RESET =
      SetLANDiscoveryStatus._(
          1, _omitEnumNames ? '' : 'DISCOVERY_CONFIGURED_ALLOWLIST_RESET');

  static const $core.List<SetLANDiscoveryStatus> values =
      <SetLANDiscoveryStatus>[
    DISCOVERY_CONFIGURED,
    DISCOVERY_CONFIGURED_ALLOWLIST_RESET,
  ];

  static final $core.List<SetLANDiscoveryStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 1);
  static SetLANDiscoveryStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const SetLANDiscoveryStatus._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
