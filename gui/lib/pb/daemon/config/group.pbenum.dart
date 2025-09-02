// This is a generated file - do not edit.
//
// Generated from group.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class ServerGroup extends $pb.ProtobufEnum {
  static const ServerGroup UNDEFINED =
      ServerGroup._(0, _omitEnumNames ? '' : 'UNDEFINED');
  static const ServerGroup DOUBLE_VPN =
      ServerGroup._(1, _omitEnumNames ? '' : 'DOUBLE_VPN');
  static const ServerGroup ONION_OVER_VPN =
      ServerGroup._(3, _omitEnumNames ? '' : 'ONION_OVER_VPN');
  static const ServerGroup ULTRA_FAST_TV =
      ServerGroup._(5, _omitEnumNames ? '' : 'ULTRA_FAST_TV');
  static const ServerGroup ANTI_DDOS =
      ServerGroup._(7, _omitEnumNames ? '' : 'ANTI_DDOS');
  static const ServerGroup DEDICATED_IP =
      ServerGroup._(9, _omitEnumNames ? '' : 'DEDICATED_IP');
  static const ServerGroup STANDARD_VPN_SERVERS =
      ServerGroup._(11, _omitEnumNames ? '' : 'STANDARD_VPN_SERVERS');
  static const ServerGroup NETFLIX_USA =
      ServerGroup._(13, _omitEnumNames ? '' : 'NETFLIX_USA');
  static const ServerGroup P2P = ServerGroup._(15, _omitEnumNames ? '' : 'P2P');
  static const ServerGroup OBFUSCATED =
      ServerGroup._(17, _omitEnumNames ? '' : 'OBFUSCATED');
  static const ServerGroup EUROPE =
      ServerGroup._(19, _omitEnumNames ? '' : 'EUROPE');
  static const ServerGroup THE_AMERICAS =
      ServerGroup._(21, _omitEnumNames ? '' : 'THE_AMERICAS');
  static const ServerGroup ASIA_PACIFIC =
      ServerGroup._(23, _omitEnumNames ? '' : 'ASIA_PACIFIC');
  static const ServerGroup AFRICA_THE_MIDDLE_EAST_AND_INDIA = ServerGroup._(
      25, _omitEnumNames ? '' : 'AFRICA_THE_MIDDLE_EAST_AND_INDIA');

  static const $core.List<ServerGroup> values = <ServerGroup>[
    UNDEFINED,
    DOUBLE_VPN,
    ONION_OVER_VPN,
    ULTRA_FAST_TV,
    ANTI_DDOS,
    DEDICATED_IP,
    STANDARD_VPN_SERVERS,
    NETFLIX_USA,
    P2P,
    OBFUSCATED,
    EUROPE,
    THE_AMERICAS,
    ASIA_PACIFIC,
    AFRICA_THE_MIDDLE_EAST_AND_INDIA,
  ];

  static final $core.Map<$core.int, ServerGroup> _byValue =
      $pb.ProtobufEnum.initByValue(values);
  static ServerGroup? valueOf($core.int value) => _byValue[value];

  const ServerGroup._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
