// This is a generated file - do not edit.
//
// Generated from peer.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// PeerStatus defines the current connection status with the peer
class PeerStatus extends $pb.ProtobufEnum {
  static const PeerStatus DISCONNECTED =
      PeerStatus._(0, _omitEnumNames ? '' : 'DISCONNECTED');
  static const PeerStatus CONNECTED =
      PeerStatus._(1, _omitEnumNames ? '' : 'CONNECTED');

  static const $core.List<PeerStatus> values = <PeerStatus>[
    DISCONNECTED,
    CONNECTED,
  ];

  static final $core.List<PeerStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 1);
  static PeerStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const PeerStatus._(super.value, super.name);
}

/// UpdatePeerErrorCode defines an error code on updating a peer within
/// the meshnet
class UpdatePeerErrorCode extends $pb.ProtobufEnum {
  static const UpdatePeerErrorCode PEER_NOT_FOUND =
      UpdatePeerErrorCode._(0, _omitEnumNames ? '' : 'PEER_NOT_FOUND');

  static const $core.List<UpdatePeerErrorCode> values = <UpdatePeerErrorCode>[
    PEER_NOT_FOUND,
  ];

  static final $core.List<UpdatePeerErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 0);
  static UpdatePeerErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const UpdatePeerErrorCode._(super.value, super.name);
}

/// ChangeNicknameErrorCode defines the errors that occur at meshnet nickname changes
class ChangeNicknameErrorCode extends $pb.ProtobufEnum {
  static const ChangeNicknameErrorCode SAME_NICKNAME =
      ChangeNicknameErrorCode._(0, _omitEnumNames ? '' : 'SAME_NICKNAME');
  static const ChangeNicknameErrorCode NICKNAME_ALREADY_EMPTY =
      ChangeNicknameErrorCode._(
          1, _omitEnumNames ? '' : 'NICKNAME_ALREADY_EMPTY');
  static const ChangeNicknameErrorCode DOMAIN_NAME_EXISTS =
      ChangeNicknameErrorCode._(2, _omitEnumNames ? '' : 'DOMAIN_NAME_EXISTS');
  static const ChangeNicknameErrorCode RATE_LIMIT_REACH =
      ChangeNicknameErrorCode._(3, _omitEnumNames ? '' : 'RATE_LIMIT_REACH');
  static const ChangeNicknameErrorCode NICKNAME_TOO_LONG =
      ChangeNicknameErrorCode._(4, _omitEnumNames ? '' : 'NICKNAME_TOO_LONG');
  static const ChangeNicknameErrorCode DUPLICATE_NICKNAME =
      ChangeNicknameErrorCode._(5, _omitEnumNames ? '' : 'DUPLICATE_NICKNAME');
  static const ChangeNicknameErrorCode CONTAINS_FORBIDDEN_WORD =
      ChangeNicknameErrorCode._(
          6, _omitEnumNames ? '' : 'CONTAINS_FORBIDDEN_WORD');
  static const ChangeNicknameErrorCode SUFFIX_OR_PREFIX_ARE_INVALID =
      ChangeNicknameErrorCode._(
          7, _omitEnumNames ? '' : 'SUFFIX_OR_PREFIX_ARE_INVALID');
  static const ChangeNicknameErrorCode NICKNAME_HAS_DOUBLE_HYPHENS =
      ChangeNicknameErrorCode._(
          8, _omitEnumNames ? '' : 'NICKNAME_HAS_DOUBLE_HYPHENS');
  static const ChangeNicknameErrorCode INVALID_CHARS =
      ChangeNicknameErrorCode._(9, _omitEnumNames ? '' : 'INVALID_CHARS');

  static const $core.List<ChangeNicknameErrorCode> values =
      <ChangeNicknameErrorCode>[
    SAME_NICKNAME,
    NICKNAME_ALREADY_EMPTY,
    DOMAIN_NAME_EXISTS,
    RATE_LIMIT_REACH,
    NICKNAME_TOO_LONG,
    DUPLICATE_NICKNAME,
    CONTAINS_FORBIDDEN_WORD,
    SUFFIX_OR_PREFIX_ARE_INVALID,
    NICKNAME_HAS_DOUBLE_HYPHENS,
    INVALID_CHARS,
  ];

  static final $core.List<ChangeNicknameErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 9);
  static ChangeNicknameErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ChangeNicknameErrorCode._(super.value, super.name);
}

/// AllowRoutingErrorCode defines an error code which is specific to
/// allow routing
class AllowRoutingErrorCode extends $pb.ProtobufEnum {
  static const AllowRoutingErrorCode ROUTING_ALREADY_ALLOWED =
      AllowRoutingErrorCode._(
          0, _omitEnumNames ? '' : 'ROUTING_ALREADY_ALLOWED');

  static const $core.List<AllowRoutingErrorCode> values =
      <AllowRoutingErrorCode>[
    ROUTING_ALREADY_ALLOWED,
  ];

  static final $core.List<AllowRoutingErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 0);
  static AllowRoutingErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AllowRoutingErrorCode._(super.value, super.name);
}

/// DenyRoutingErrorCode defines an error code which is specific to
/// deny routing
class DenyRoutingErrorCode extends $pb.ProtobufEnum {
  static const DenyRoutingErrorCode ROUTING_ALREADY_DENIED =
      DenyRoutingErrorCode._(0, _omitEnumNames ? '' : 'ROUTING_ALREADY_DENIED');

  static const $core.List<DenyRoutingErrorCode> values = <DenyRoutingErrorCode>[
    ROUTING_ALREADY_DENIED,
  ];

  static final $core.List<DenyRoutingErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 0);
  static DenyRoutingErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const DenyRoutingErrorCode._(super.value, super.name);
}

/// AllowIncomingErrorCode defines an error code which is specific to
/// allow incoming traffic
class AllowIncomingErrorCode extends $pb.ProtobufEnum {
  static const AllowIncomingErrorCode INCOMING_ALREADY_ALLOWED =
      AllowIncomingErrorCode._(
          0, _omitEnumNames ? '' : 'INCOMING_ALREADY_ALLOWED');

  static const $core.List<AllowIncomingErrorCode> values =
      <AllowIncomingErrorCode>[
    INCOMING_ALREADY_ALLOWED,
  ];

  static final $core.List<AllowIncomingErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 0);
  static AllowIncomingErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AllowIncomingErrorCode._(super.value, super.name);
}

/// DenyIncomingErrorCode defines an error code which is specific to
/// deny incoming traffic
class DenyIncomingErrorCode extends $pb.ProtobufEnum {
  static const DenyIncomingErrorCode INCOMING_ALREADY_DENIED =
      DenyIncomingErrorCode._(
          0, _omitEnumNames ? '' : 'INCOMING_ALREADY_DENIED');

  static const $core.List<DenyIncomingErrorCode> values =
      <DenyIncomingErrorCode>[
    INCOMING_ALREADY_DENIED,
  ];

  static final $core.List<DenyIncomingErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 0);
  static DenyIncomingErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const DenyIncomingErrorCode._(super.value, super.name);
}

/// AllowLocalNetworkErrorCode defines an error code which is specific to
/// allow local network traffic
class AllowLocalNetworkErrorCode extends $pb.ProtobufEnum {
  static const AllowLocalNetworkErrorCode LOCAL_NETWORK_ALREADY_ALLOWED =
      AllowLocalNetworkErrorCode._(
          0, _omitEnumNames ? '' : 'LOCAL_NETWORK_ALREADY_ALLOWED');

  static const $core.List<AllowLocalNetworkErrorCode> values =
      <AllowLocalNetworkErrorCode>[
    LOCAL_NETWORK_ALREADY_ALLOWED,
  ];

  static final $core.List<AllowLocalNetworkErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 0);
  static AllowLocalNetworkErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AllowLocalNetworkErrorCode._(super.value, super.name);
}

/// DenyLocalNetworkErrorCode defines an error code which is specific to
/// deny local network traffic
class DenyLocalNetworkErrorCode extends $pb.ProtobufEnum {
  static const DenyLocalNetworkErrorCode LOCAL_NETWORK_ALREADY_DENIED =
      DenyLocalNetworkErrorCode._(
          0, _omitEnumNames ? '' : 'LOCAL_NETWORK_ALREADY_DENIED');

  static const $core.List<DenyLocalNetworkErrorCode> values =
      <DenyLocalNetworkErrorCode>[
    LOCAL_NETWORK_ALREADY_DENIED,
  ];

  static final $core.List<DenyLocalNetworkErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 0);
  static DenyLocalNetworkErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const DenyLocalNetworkErrorCode._(super.value, super.name);
}

class AllowFileshareErrorCode extends $pb.ProtobufEnum {
  static const AllowFileshareErrorCode SEND_ALREADY_ALLOWED =
      AllowFileshareErrorCode._(
          0, _omitEnumNames ? '' : 'SEND_ALREADY_ALLOWED');

  static const $core.List<AllowFileshareErrorCode> values =
      <AllowFileshareErrorCode>[
    SEND_ALREADY_ALLOWED,
  ];

  static final $core.List<AllowFileshareErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 0);
  static AllowFileshareErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AllowFileshareErrorCode._(super.value, super.name);
}

class DenyFileshareErrorCode extends $pb.ProtobufEnum {
  static const DenyFileshareErrorCode SEND_ALREADY_DENIED =
      DenyFileshareErrorCode._(0, _omitEnumNames ? '' : 'SEND_ALREADY_DENIED');

  static const $core.List<DenyFileshareErrorCode> values =
      <DenyFileshareErrorCode>[
    SEND_ALREADY_DENIED,
  ];

  static final $core.List<DenyFileshareErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 0);
  static DenyFileshareErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const DenyFileshareErrorCode._(super.value, super.name);
}

class EnableAutomaticFileshareErrorCode extends $pb.ProtobufEnum {
  static const EnableAutomaticFileshareErrorCode
      AUTOMATIC_FILESHARE_ALREADY_ENABLED = EnableAutomaticFileshareErrorCode._(
          0, _omitEnumNames ? '' : 'AUTOMATIC_FILESHARE_ALREADY_ENABLED');

  static const $core.List<EnableAutomaticFileshareErrorCode> values =
      <EnableAutomaticFileshareErrorCode>[
    AUTOMATIC_FILESHARE_ALREADY_ENABLED,
  ];

  static final $core.List<EnableAutomaticFileshareErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 0);
  static EnableAutomaticFileshareErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const EnableAutomaticFileshareErrorCode._(super.value, super.name);
}

class DisableAutomaticFileshareErrorCode extends $pb.ProtobufEnum {
  static const DisableAutomaticFileshareErrorCode
      AUTOMATIC_FILESHARE_ALREADY_DISABLED =
      DisableAutomaticFileshareErrorCode._(
          0, _omitEnumNames ? '' : 'AUTOMATIC_FILESHARE_ALREADY_DISABLED');

  static const $core.List<DisableAutomaticFileshareErrorCode> values =
      <DisableAutomaticFileshareErrorCode>[
    AUTOMATIC_FILESHARE_ALREADY_DISABLED,
  ];

  static final $core.List<DisableAutomaticFileshareErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 0);
  static DisableAutomaticFileshareErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const DisableAutomaticFileshareErrorCode._(super.value, super.name);
}

class ConnectErrorCode extends $pb.ProtobufEnum {
  static const ConnectErrorCode PEER_DOES_NOT_ALLOW_ROUTING =
      ConnectErrorCode._(
          0, _omitEnumNames ? '' : 'PEER_DOES_NOT_ALLOW_ROUTING');
  static const ConnectErrorCode ALREADY_CONNECTED =
      ConnectErrorCode._(1, _omitEnumNames ? '' : 'ALREADY_CONNECTED');
  static const ConnectErrorCode CONNECT_FAILED =
      ConnectErrorCode._(2, _omitEnumNames ? '' : 'CONNECT_FAILED');
  static const ConnectErrorCode PEER_NO_IP =
      ConnectErrorCode._(3, _omitEnumNames ? '' : 'PEER_NO_IP');
  static const ConnectErrorCode ALREADY_CONNECTING =
      ConnectErrorCode._(4, _omitEnumNames ? '' : 'ALREADY_CONNECTING');
  static const ConnectErrorCode CANCELED =
      ConnectErrorCode._(5, _omitEnumNames ? '' : 'CANCELED');

  static const $core.List<ConnectErrorCode> values = <ConnectErrorCode>[
    PEER_DOES_NOT_ALLOW_ROUTING,
    ALREADY_CONNECTED,
    CONNECT_FAILED,
    PEER_NO_IP,
    ALREADY_CONNECTING,
    CANCELED,
  ];

  static final $core.List<ConnectErrorCode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static ConnectErrorCode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ConnectErrorCode._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
