// This is a generated file - do not edit.
//
// Generated from vpn_protocol.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class VPNProtocol extends $pb.ProtobufEnum {
  static const VPNProtocol VPN_PROTOCOL_UNSPECIFIED =
      VPNProtocol._(0, _omitEnumNames ? '' : 'VPN_PROTOCOL_UNSPECIFIED');
  static const VPNProtocol VPN_PROTOCOL_NORDLYNX =
      VPNProtocol._(1, _omitEnumNames ? '' : 'VPN_PROTOCOL_NORDLYNX');
  static const VPNProtocol VPN_PROTOCOL_OPENVPN_UDP =
      VPNProtocol._(2, _omitEnumNames ? '' : 'VPN_PROTOCOL_OPENVPN_UDP');
  static const VPNProtocol VPN_PROTOCOL_OPENVPN_TCP =
      VPNProtocol._(3, _omitEnumNames ? '' : 'VPN_PROTOCOL_OPENVPN_TCP');
  static const VPNProtocol VPN_PROTOCOL_NORDWHISPER =
      VPNProtocol._(4, _omitEnumNames ? '' : 'VPN_PROTOCOL_NORDWHISPER');

  static const $core.List<VPNProtocol> values = <VPNProtocol>[
    VPN_PROTOCOL_UNSPECIFIED,
    VPN_PROTOCOL_NORDLYNX,
    VPN_PROTOCOL_OPENVPN_UDP,
    VPN_PROTOCOL_OPENVPN_TCP,
    VPN_PROTOCOL_NORDWHISPER,
  ];

  static final $core.List<VPNProtocol?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static VPNProtocol? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const VPNProtocol._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
