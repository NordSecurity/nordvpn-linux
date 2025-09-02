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

import 'empty.pb.dart' as $0;
import 'peer.pbenum.dart';
import 'service_response.pbenum.dart' as $1;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'peer.pbenum.dart';

enum GetPeersResponse_Response { peers, error, notSet }

/// GetPeersResponse defines
class GetPeersResponse extends $pb.GeneratedMessage {
  factory GetPeersResponse({
    PeerList? peers,
    Error? error,
  }) {
    final result = create();
    if (peers != null) result.peers = peers;
    if (error != null) result.error = error;
    return result;
  }

  GetPeersResponse._();

  factory GetPeersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetPeersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, GetPeersResponse_Response>
      _GetPeersResponse_ResponseByTag = {
    1: GetPeersResponse_Response.peers,
    4: GetPeersResponse_Response.error,
    0: GetPeersResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetPeersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 4])
    ..aOM<PeerList>(1, _omitFieldNames ? '' : 'peers',
        subBuilder: PeerList.create)
    ..aOM<Error>(4, _omitFieldNames ? '' : 'error', subBuilder: Error.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetPeersResponse clone() => GetPeersResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetPeersResponse copyWith(void Function(GetPeersResponse) updates) =>
      super.copyWith((message) => updates(message as GetPeersResponse))
          as GetPeersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetPeersResponse create() => GetPeersResponse._();
  @$core.override
  GetPeersResponse createEmptyInstance() => create();
  static $pb.PbList<GetPeersResponse> createRepeated() =>
      $pb.PbList<GetPeersResponse>();
  @$core.pragma('dart2js:noInline')
  static GetPeersResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetPeersResponse>(create);
  static GetPeersResponse? _defaultInstance;

  GetPeersResponse_Response whichResponse() =>
      _GetPeersResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  PeerList get peers => $_getN(0);
  @$pb.TagNumber(1)
  set peers(PeerList value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasPeers() => $_has(0);
  @$pb.TagNumber(1)
  void clearPeers() => $_clearField(1);
  @$pb.TagNumber(1)
  PeerList ensurePeers() => $_ensure(0);

  @$pb.TagNumber(4)
  Error get error => $_getN(1);
  @$pb.TagNumber(4)
  set error(Error value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasError() => $_has(1);
  @$pb.TagNumber(4)
  void clearError() => $_clearField(4);
  @$pb.TagNumber(4)
  Error ensureError() => $_ensure(1);
}

/// PeerList defines a list of all the peers related to the device
class PeerList extends $pb.GeneratedMessage {
  factory PeerList({
    Peer? self,
    $core.Iterable<Peer>? local,
    $core.Iterable<Peer>? external,
  }) {
    final result = create();
    if (self != null) result.self = self;
    if (local != null) result.local.addAll(local);
    if (external != null) result.external.addAll(external);
    return result;
  }

  PeerList._();

  factory PeerList.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory PeerList.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PeerList',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..aOM<Peer>(1, _omitFieldNames ? '' : 'self', subBuilder: Peer.create)
    ..pc<Peer>(2, _omitFieldNames ? '' : 'local', $pb.PbFieldType.PM,
        subBuilder: Peer.create)
    ..pc<Peer>(3, _omitFieldNames ? '' : 'external', $pb.PbFieldType.PM,
        subBuilder: Peer.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PeerList clone() => PeerList()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PeerList copyWith(void Function(PeerList) updates) =>
      super.copyWith((message) => updates(message as PeerList)) as PeerList;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PeerList create() => PeerList._();
  @$core.override
  PeerList createEmptyInstance() => create();
  static $pb.PbList<PeerList> createRepeated() => $pb.PbList<PeerList>();
  @$core.pragma('dart2js:noInline')
  static PeerList getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PeerList>(create);
  static PeerList? _defaultInstance;

  @$pb.TagNumber(1)
  Peer get self => $_getN(0);
  @$pb.TagNumber(1)
  set self(Peer value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSelf() => $_has(0);
  @$pb.TagNumber(1)
  void clearSelf() => $_clearField(1);
  @$pb.TagNumber(1)
  Peer ensureSelf() => $_ensure(0);

  @$pb.TagNumber(2)
  $pb.PbList<Peer> get local => $_getList(1);

  @$pb.TagNumber(3)
  $pb.PbList<Peer> get external => $_getList(2);
}

/// Peer defines a single meshnet peer
class Peer extends $pb.GeneratedMessage {
  factory Peer({
    $core.String? identifier,
    $core.String? pubkey,
    $core.String? ip,
    $core.Iterable<$core.String>? endpoints,
    $core.String? os,
    $core.String? osVersion,
    $core.String? hostname,
    $core.String? distro,
    $core.String? email,
    $core.bool? isInboundAllowed,
    $core.bool? isRoutable,
    $core.bool? doIAllowInbound,
    $core.bool? doIAllowRouting,
    PeerStatus? status,
    $core.bool? isLocalNetworkAllowed,
    $core.bool? doIAllowLocalNetwork,
    $core.bool? isFileshareAllowed,
    $core.bool? doIAllowFileshare,
    $core.bool? alwaysAcceptFiles,
    $core.String? nickname,
  }) {
    final result = create();
    if (identifier != null) result.identifier = identifier;
    if (pubkey != null) result.pubkey = pubkey;
    if (ip != null) result.ip = ip;
    if (endpoints != null) result.endpoints.addAll(endpoints);
    if (os != null) result.os = os;
    if (osVersion != null) result.osVersion = osVersion;
    if (hostname != null) result.hostname = hostname;
    if (distro != null) result.distro = distro;
    if (email != null) result.email = email;
    if (isInboundAllowed != null) result.isInboundAllowed = isInboundAllowed;
    if (isRoutable != null) result.isRoutable = isRoutable;
    if (doIAllowInbound != null) result.doIAllowInbound = doIAllowInbound;
    if (doIAllowRouting != null) result.doIAllowRouting = doIAllowRouting;
    if (status != null) result.status = status;
    if (isLocalNetworkAllowed != null)
      result.isLocalNetworkAllowed = isLocalNetworkAllowed;
    if (doIAllowLocalNetwork != null)
      result.doIAllowLocalNetwork = doIAllowLocalNetwork;
    if (isFileshareAllowed != null)
      result.isFileshareAllowed = isFileshareAllowed;
    if (doIAllowFileshare != null) result.doIAllowFileshare = doIAllowFileshare;
    if (alwaysAcceptFiles != null) result.alwaysAcceptFiles = alwaysAcceptFiles;
    if (nickname != null) result.nickname = nickname;
    return result;
  }

  Peer._();

  factory Peer.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Peer.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Peer',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'identifier')
    ..aOS(2, _omitFieldNames ? '' : 'pubkey')
    ..aOS(3, _omitFieldNames ? '' : 'ip')
    ..pPS(4, _omitFieldNames ? '' : 'endpoints')
    ..aOS(5, _omitFieldNames ? '' : 'os')
    ..aOS(6, _omitFieldNames ? '' : 'osVersion')
    ..aOS(7, _omitFieldNames ? '' : 'hostname')
    ..aOS(8, _omitFieldNames ? '' : 'distro')
    ..aOS(9, _omitFieldNames ? '' : 'email')
    ..aOB(10, _omitFieldNames ? '' : 'isInboundAllowed')
    ..aOB(11, _omitFieldNames ? '' : 'isRoutable')
    ..aOB(12, _omitFieldNames ? '' : 'doIAllowInbound')
    ..aOB(13, _omitFieldNames ? '' : 'doIAllowRouting')
    ..e<PeerStatus>(14, _omitFieldNames ? '' : 'status', $pb.PbFieldType.OE,
        defaultOrMaker: PeerStatus.DISCONNECTED,
        valueOf: PeerStatus.valueOf,
        enumValues: PeerStatus.values)
    ..aOB(15, _omitFieldNames ? '' : 'isLocalNetworkAllowed')
    ..aOB(16, _omitFieldNames ? '' : 'doIAllowLocalNetwork')
    ..aOB(17, _omitFieldNames ? '' : 'isFileshareAllowed')
    ..aOB(18, _omitFieldNames ? '' : 'doIAllowFileshare')
    ..aOB(19, _omitFieldNames ? '' : 'alwaysAcceptFiles')
    ..aOS(20, _omitFieldNames ? '' : 'nickname')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Peer clone() => Peer()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Peer copyWith(void Function(Peer) updates) =>
      super.copyWith((message) => updates(message as Peer)) as Peer;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Peer create() => Peer._();
  @$core.override
  Peer createEmptyInstance() => create();
  static $pb.PbList<Peer> createRepeated() => $pb.PbList<Peer>();
  @$core.pragma('dart2js:noInline')
  static Peer getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Peer>(create);
  static Peer? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get identifier => $_getSZ(0);
  @$pb.TagNumber(1)
  set identifier($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasIdentifier() => $_has(0);
  @$pb.TagNumber(1)
  void clearIdentifier() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get pubkey => $_getSZ(1);
  @$pb.TagNumber(2)
  set pubkey($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPubkey() => $_has(1);
  @$pb.TagNumber(2)
  void clearPubkey() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get ip => $_getSZ(2);
  @$pb.TagNumber(3)
  set ip($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasIp() => $_has(2);
  @$pb.TagNumber(3)
  void clearIp() => $_clearField(3);

  @$pb.TagNumber(4)
  $pb.PbList<$core.String> get endpoints => $_getList(3);

  @$pb.TagNumber(5)
  $core.String get os => $_getSZ(4);
  @$pb.TagNumber(5)
  set os($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasOs() => $_has(4);
  @$pb.TagNumber(5)
  void clearOs() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get osVersion => $_getSZ(5);
  @$pb.TagNumber(6)
  set osVersion($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasOsVersion() => $_has(5);
  @$pb.TagNumber(6)
  void clearOsVersion() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get hostname => $_getSZ(6);
  @$pb.TagNumber(7)
  set hostname($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasHostname() => $_has(6);
  @$pb.TagNumber(7)
  void clearHostname() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get distro => $_getSZ(7);
  @$pb.TagNumber(8)
  set distro($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasDistro() => $_has(7);
  @$pb.TagNumber(8)
  void clearDistro() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get email => $_getSZ(8);
  @$pb.TagNumber(9)
  set email($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasEmail() => $_has(8);
  @$pb.TagNumber(9)
  void clearEmail() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.bool get isInboundAllowed => $_getBF(9);
  @$pb.TagNumber(10)
  set isInboundAllowed($core.bool value) => $_setBool(9, value);
  @$pb.TagNumber(10)
  $core.bool hasIsInboundAllowed() => $_has(9);
  @$pb.TagNumber(10)
  void clearIsInboundAllowed() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.bool get isRoutable => $_getBF(10);
  @$pb.TagNumber(11)
  set isRoutable($core.bool value) => $_setBool(10, value);
  @$pb.TagNumber(11)
  $core.bool hasIsRoutable() => $_has(10);
  @$pb.TagNumber(11)
  void clearIsRoutable() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.bool get doIAllowInbound => $_getBF(11);
  @$pb.TagNumber(12)
  set doIAllowInbound($core.bool value) => $_setBool(11, value);
  @$pb.TagNumber(12)
  $core.bool hasDoIAllowInbound() => $_has(11);
  @$pb.TagNumber(12)
  void clearDoIAllowInbound() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.bool get doIAllowRouting => $_getBF(12);
  @$pb.TagNumber(13)
  set doIAllowRouting($core.bool value) => $_setBool(12, value);
  @$pb.TagNumber(13)
  $core.bool hasDoIAllowRouting() => $_has(12);
  @$pb.TagNumber(13)
  void clearDoIAllowRouting() => $_clearField(13);

  @$pb.TagNumber(14)
  PeerStatus get status => $_getN(13);
  @$pb.TagNumber(14)
  set status(PeerStatus value) => $_setField(14, value);
  @$pb.TagNumber(14)
  $core.bool hasStatus() => $_has(13);
  @$pb.TagNumber(14)
  void clearStatus() => $_clearField(14);

  @$pb.TagNumber(15)
  $core.bool get isLocalNetworkAllowed => $_getBF(14);
  @$pb.TagNumber(15)
  set isLocalNetworkAllowed($core.bool value) => $_setBool(14, value);
  @$pb.TagNumber(15)
  $core.bool hasIsLocalNetworkAllowed() => $_has(14);
  @$pb.TagNumber(15)
  void clearIsLocalNetworkAllowed() => $_clearField(15);

  @$pb.TagNumber(16)
  $core.bool get doIAllowLocalNetwork => $_getBF(15);
  @$pb.TagNumber(16)
  set doIAllowLocalNetwork($core.bool value) => $_setBool(15, value);
  @$pb.TagNumber(16)
  $core.bool hasDoIAllowLocalNetwork() => $_has(15);
  @$pb.TagNumber(16)
  void clearDoIAllowLocalNetwork() => $_clearField(16);

  @$pb.TagNumber(17)
  $core.bool get isFileshareAllowed => $_getBF(16);
  @$pb.TagNumber(17)
  set isFileshareAllowed($core.bool value) => $_setBool(16, value);
  @$pb.TagNumber(17)
  $core.bool hasIsFileshareAllowed() => $_has(16);
  @$pb.TagNumber(17)
  void clearIsFileshareAllowed() => $_clearField(17);

  @$pb.TagNumber(18)
  $core.bool get doIAllowFileshare => $_getBF(17);
  @$pb.TagNumber(18)
  set doIAllowFileshare($core.bool value) => $_setBool(17, value);
  @$pb.TagNumber(18)
  $core.bool hasDoIAllowFileshare() => $_has(17);
  @$pb.TagNumber(18)
  void clearDoIAllowFileshare() => $_clearField(18);

  @$pb.TagNumber(19)
  $core.bool get alwaysAcceptFiles => $_getBF(18);
  @$pb.TagNumber(19)
  set alwaysAcceptFiles($core.bool value) => $_setBool(18, value);
  @$pb.TagNumber(19)
  $core.bool hasAlwaysAcceptFiles() => $_has(18);
  @$pb.TagNumber(19)
  void clearAlwaysAcceptFiles() => $_clearField(19);

  @$pb.TagNumber(20)
  $core.String get nickname => $_getSZ(19);
  @$pb.TagNumber(20)
  set nickname($core.String value) => $_setString(19, value);
  @$pb.TagNumber(20)
  $core.bool hasNickname() => $_has(19);
  @$pb.TagNumber(20)
  void clearNickname() => $_clearField(20);
}

/// UpdatePeerRequest defines a request to remove a peer from a meshnet
class UpdatePeerRequest extends $pb.GeneratedMessage {
  factory UpdatePeerRequest({
    $core.String? identifier,
  }) {
    final result = create();
    if (identifier != null) result.identifier = identifier;
    return result;
  }

  UpdatePeerRequest._();

  factory UpdatePeerRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UpdatePeerRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdatePeerRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'identifier')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdatePeerRequest clone() => UpdatePeerRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdatePeerRequest copyWith(void Function(UpdatePeerRequest) updates) =>
      super.copyWith((message) => updates(message as UpdatePeerRequest))
          as UpdatePeerRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdatePeerRequest create() => UpdatePeerRequest._();
  @$core.override
  UpdatePeerRequest createEmptyInstance() => create();
  static $pb.PbList<UpdatePeerRequest> createRepeated() =>
      $pb.PbList<UpdatePeerRequest>();
  @$core.pragma('dart2js:noInline')
  static UpdatePeerRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdatePeerRequest>(create);
  static UpdatePeerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get identifier => $_getSZ(0);
  @$pb.TagNumber(1)
  set identifier($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasIdentifier() => $_has(0);
  @$pb.TagNumber(1)
  void clearIdentifier() => $_clearField(1);
}

enum Error_Error { serviceErrorCode, meshnetErrorCode, notSet }

/// Error defines a generic meshnet error that could be returned by most meshnet endpoints
class Error extends $pb.GeneratedMessage {
  factory Error({
    $1.ServiceErrorCode? serviceErrorCode,
    $1.MeshnetErrorCode? meshnetErrorCode,
  }) {
    final result = create();
    if (serviceErrorCode != null) result.serviceErrorCode = serviceErrorCode;
    if (meshnetErrorCode != null) result.meshnetErrorCode = meshnetErrorCode;
    return result;
  }

  Error._();

  factory Error.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Error.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, Error_Error> _Error_ErrorByTag = {
    1: Error_Error.serviceErrorCode,
    2: Error_Error.meshnetErrorCode,
    0: Error_Error.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Error',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2])
    ..e<$1.ServiceErrorCode>(
        1, _omitFieldNames ? '' : 'serviceErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: $1.ServiceErrorCode.NOT_LOGGED_IN,
        valueOf: $1.ServiceErrorCode.valueOf,
        enumValues: $1.ServiceErrorCode.values)
    ..e<$1.MeshnetErrorCode>(
        2, _omitFieldNames ? '' : 'meshnetErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: $1.MeshnetErrorCode.NOT_REGISTERED,
        valueOf: $1.MeshnetErrorCode.valueOf,
        enumValues: $1.MeshnetErrorCode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Error clone() => Error()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Error copyWith(void Function(Error) updates) =>
      super.copyWith((message) => updates(message as Error)) as Error;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Error create() => Error._();
  @$core.override
  Error createEmptyInstance() => create();
  static $pb.PbList<Error> createRepeated() => $pb.PbList<Error>();
  @$core.pragma('dart2js:noInline')
  static Error getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Error>(create);
  static Error? _defaultInstance;

  Error_Error whichError() => _Error_ErrorByTag[$_whichOneof(0)]!;
  void clearError() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $1.ServiceErrorCode get serviceErrorCode => $_getN(0);
  @$pb.TagNumber(1)
  set serviceErrorCode($1.ServiceErrorCode value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasServiceErrorCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearServiceErrorCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $1.MeshnetErrorCode get meshnetErrorCode => $_getN(1);
  @$pb.TagNumber(2)
  set meshnetErrorCode($1.MeshnetErrorCode value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasMeshnetErrorCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearMeshnetErrorCode() => $_clearField(2);
}

enum UpdatePeerError_Error { generalError, updatePeerErrorCode, notSet }

/// UpdatePeerError can be either generic meshnet error or generic peer update error
class UpdatePeerError extends $pb.GeneratedMessage {
  factory UpdatePeerError({
    Error? generalError,
    UpdatePeerErrorCode? updatePeerErrorCode,
  }) {
    final result = create();
    if (generalError != null) result.generalError = generalError;
    if (updatePeerErrorCode != null)
      result.updatePeerErrorCode = updatePeerErrorCode;
    return result;
  }

  UpdatePeerError._();

  factory UpdatePeerError.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UpdatePeerError.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, UpdatePeerError_Error>
      _UpdatePeerError_ErrorByTag = {
    1: UpdatePeerError_Error.generalError,
    2: UpdatePeerError_Error.updatePeerErrorCode,
    0: UpdatePeerError_Error.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdatePeerError',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2])
    ..aOM<Error>(1, _omitFieldNames ? '' : 'generalError',
        subBuilder: Error.create)
    ..e<UpdatePeerErrorCode>(
        2, _omitFieldNames ? '' : 'updatePeerErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: UpdatePeerErrorCode.PEER_NOT_FOUND,
        valueOf: UpdatePeerErrorCode.valueOf,
        enumValues: UpdatePeerErrorCode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdatePeerError clone() => UpdatePeerError()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdatePeerError copyWith(void Function(UpdatePeerError) updates) =>
      super.copyWith((message) => updates(message as UpdatePeerError))
          as UpdatePeerError;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdatePeerError create() => UpdatePeerError._();
  @$core.override
  UpdatePeerError createEmptyInstance() => create();
  static $pb.PbList<UpdatePeerError> createRepeated() =>
      $pb.PbList<UpdatePeerError>();
  @$core.pragma('dart2js:noInline')
  static UpdatePeerError getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdatePeerError>(create);
  static UpdatePeerError? _defaultInstance;

  UpdatePeerError_Error whichError() =>
      _UpdatePeerError_ErrorByTag[$_whichOneof(0)]!;
  void clearError() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  Error get generalError => $_getN(0);
  @$pb.TagNumber(1)
  set generalError(Error value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasGeneralError() => $_has(0);
  @$pb.TagNumber(1)
  void clearGeneralError() => $_clearField(1);
  @$pb.TagNumber(1)
  Error ensureGeneralError() => $_ensure(0);

  @$pb.TagNumber(2)
  UpdatePeerErrorCode get updatePeerErrorCode => $_getN(1);
  @$pb.TagNumber(2)
  set updatePeerErrorCode(UpdatePeerErrorCode value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasUpdatePeerErrorCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearUpdatePeerErrorCode() => $_clearField(2);
}

enum RemovePeerResponse_Response { empty, updatePeerError, notSet }

/// RemovePeerResponse defines a peer removal response
class RemovePeerResponse extends $pb.GeneratedMessage {
  factory RemovePeerResponse({
    $0.Empty? empty,
    UpdatePeerError? updatePeerError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (updatePeerError != null) result.updatePeerError = updatePeerError;
    return result;
  }

  RemovePeerResponse._();

  factory RemovePeerResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RemovePeerResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, RemovePeerResponse_Response>
      _RemovePeerResponse_ResponseByTag = {
    1: RemovePeerResponse_Response.empty,
    5: RemovePeerResponse_Response.updatePeerError,
    0: RemovePeerResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RemovePeerResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 5])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..aOM<UpdatePeerError>(5, _omitFieldNames ? '' : 'updatePeerError',
        subBuilder: UpdatePeerError.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RemovePeerResponse clone() => RemovePeerResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RemovePeerResponse copyWith(void Function(RemovePeerResponse) updates) =>
      super.copyWith((message) => updates(message as RemovePeerResponse))
          as RemovePeerResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RemovePeerResponse create() => RemovePeerResponse._();
  @$core.override
  RemovePeerResponse createEmptyInstance() => create();
  static $pb.PbList<RemovePeerResponse> createRepeated() =>
      $pb.PbList<RemovePeerResponse>();
  @$core.pragma('dart2js:noInline')
  static RemovePeerResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RemovePeerResponse>(create);
  static RemovePeerResponse? _defaultInstance;

  RemovePeerResponse_Response whichResponse() =>
      _RemovePeerResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $0.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($0.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(5)
  UpdatePeerError get updatePeerError => $_getN(1);
  @$pb.TagNumber(5)
  set updatePeerError(UpdatePeerError value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasUpdatePeerError() => $_has(1);
  @$pb.TagNumber(5)
  void clearUpdatePeerError() => $_clearField(5);
  @$pb.TagNumber(5)
  UpdatePeerError ensureUpdatePeerError() => $_ensure(1);
}

/// ChangePeerNicknameRequest defines a request to change the nickname for a meshnet peer
class ChangePeerNicknameRequest extends $pb.GeneratedMessage {
  factory ChangePeerNicknameRequest({
    $core.String? identifier,
    $core.String? nickname,
  }) {
    final result = create();
    if (identifier != null) result.identifier = identifier;
    if (nickname != null) result.nickname = nickname;
    return result;
  }

  ChangePeerNicknameRequest._();

  factory ChangePeerNicknameRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ChangePeerNicknameRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChangePeerNicknameRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'identifier')
    ..aOS(2, _omitFieldNames ? '' : 'nickname')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangePeerNicknameRequest clone() =>
      ChangePeerNicknameRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangePeerNicknameRequest copyWith(
          void Function(ChangePeerNicknameRequest) updates) =>
      super.copyWith((message) => updates(message as ChangePeerNicknameRequest))
          as ChangePeerNicknameRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ChangePeerNicknameRequest create() => ChangePeerNicknameRequest._();
  @$core.override
  ChangePeerNicknameRequest createEmptyInstance() => create();
  static $pb.PbList<ChangePeerNicknameRequest> createRepeated() =>
      $pb.PbList<ChangePeerNicknameRequest>();
  @$core.pragma('dart2js:noInline')
  static ChangePeerNicknameRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ChangePeerNicknameRequest>(create);
  static ChangePeerNicknameRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get identifier => $_getSZ(0);
  @$pb.TagNumber(1)
  set identifier($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasIdentifier() => $_has(0);
  @$pb.TagNumber(1)
  void clearIdentifier() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get nickname => $_getSZ(1);
  @$pb.TagNumber(2)
  set nickname($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNickname() => $_has(1);
  @$pb.TagNumber(2)
  void clearNickname() => $_clearField(2);
}

/// ChangeMachineNicknameRequest defines a request to change the nickname for the current machine from meshnet
class ChangeMachineNicknameRequest extends $pb.GeneratedMessage {
  factory ChangeMachineNicknameRequest({
    $core.String? nickname,
  }) {
    final result = create();
    if (nickname != null) result.nickname = nickname;
    return result;
  }

  ChangeMachineNicknameRequest._();

  factory ChangeMachineNicknameRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ChangeMachineNicknameRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChangeMachineNicknameRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'nickname')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangeMachineNicknameRequest clone() =>
      ChangeMachineNicknameRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangeMachineNicknameRequest copyWith(
          void Function(ChangeMachineNicknameRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ChangeMachineNicknameRequest))
          as ChangeMachineNicknameRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ChangeMachineNicknameRequest create() =>
      ChangeMachineNicknameRequest._();
  @$core.override
  ChangeMachineNicknameRequest createEmptyInstance() => create();
  static $pb.PbList<ChangeMachineNicknameRequest> createRepeated() =>
      $pb.PbList<ChangeMachineNicknameRequest>();
  @$core.pragma('dart2js:noInline')
  static ChangeMachineNicknameRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ChangeMachineNicknameRequest>(create);
  static ChangeMachineNicknameRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get nickname => $_getSZ(0);
  @$pb.TagNumber(1)
  set nickname($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasNickname() => $_has(0);
  @$pb.TagNumber(1)
  void clearNickname() => $_clearField(1);
}

enum ChangeNicknameResponse_Response {
  empty,
  changeNicknameErrorCode,
  updatePeerError,
  notSet
}

/// ChangeNicknameResponse defines a response to change(set/remove) the nickname for a peer or for current machine
class ChangeNicknameResponse extends $pb.GeneratedMessage {
  factory ChangeNicknameResponse({
    $0.Empty? empty,
    ChangeNicknameErrorCode? changeNicknameErrorCode,
    UpdatePeerError? updatePeerError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (changeNicknameErrorCode != null)
      result.changeNicknameErrorCode = changeNicknameErrorCode;
    if (updatePeerError != null) result.updatePeerError = updatePeerError;
    return result;
  }

  ChangeNicknameResponse._();

  factory ChangeNicknameResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ChangeNicknameResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, ChangeNicknameResponse_Response>
      _ChangeNicknameResponse_ResponseByTag = {
    1: ChangeNicknameResponse_Response.empty,
    5: ChangeNicknameResponse_Response.changeNicknameErrorCode,
    6: ChangeNicknameResponse_Response.updatePeerError,
    0: ChangeNicknameResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChangeNicknameResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 5, 6])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<ChangeNicknameErrorCode>(
        5, _omitFieldNames ? '' : 'changeNicknameErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: ChangeNicknameErrorCode.SAME_NICKNAME,
        valueOf: ChangeNicknameErrorCode.valueOf,
        enumValues: ChangeNicknameErrorCode.values)
    ..aOM<UpdatePeerError>(6, _omitFieldNames ? '' : 'updatePeerError',
        subBuilder: UpdatePeerError.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangeNicknameResponse clone() =>
      ChangeNicknameResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangeNicknameResponse copyWith(
          void Function(ChangeNicknameResponse) updates) =>
      super.copyWith((message) => updates(message as ChangeNicknameResponse))
          as ChangeNicknameResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ChangeNicknameResponse create() => ChangeNicknameResponse._();
  @$core.override
  ChangeNicknameResponse createEmptyInstance() => create();
  static $pb.PbList<ChangeNicknameResponse> createRepeated() =>
      $pb.PbList<ChangeNicknameResponse>();
  @$core.pragma('dart2js:noInline')
  static ChangeNicknameResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ChangeNicknameResponse>(create);
  static ChangeNicknameResponse? _defaultInstance;

  ChangeNicknameResponse_Response whichResponse() =>
      _ChangeNicknameResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $0.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($0.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(5)
  ChangeNicknameErrorCode get changeNicknameErrorCode => $_getN(1);
  @$pb.TagNumber(5)
  set changeNicknameErrorCode(ChangeNicknameErrorCode value) =>
      $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasChangeNicknameErrorCode() => $_has(1);
  @$pb.TagNumber(5)
  void clearChangeNicknameErrorCode() => $_clearField(5);

  @$pb.TagNumber(6)
  UpdatePeerError get updatePeerError => $_getN(2);
  @$pb.TagNumber(6)
  set updatePeerError(UpdatePeerError value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdatePeerError() => $_has(2);
  @$pb.TagNumber(6)
  void clearUpdatePeerError() => $_clearField(6);
  @$pb.TagNumber(6)
  UpdatePeerError ensureUpdatePeerError() => $_ensure(2);
}

enum AllowRoutingResponse_Response {
  empty,
  allowRoutingErrorCode,
  updatePeerError,
  notSet
}

/// AllowRoutingResponse defines a response for allow routing request
class AllowRoutingResponse extends $pb.GeneratedMessage {
  factory AllowRoutingResponse({
    $0.Empty? empty,
    AllowRoutingErrorCode? allowRoutingErrorCode,
    UpdatePeerError? updatePeerError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (allowRoutingErrorCode != null)
      result.allowRoutingErrorCode = allowRoutingErrorCode;
    if (updatePeerError != null) result.updatePeerError = updatePeerError;
    return result;
  }

  AllowRoutingResponse._();

  factory AllowRoutingResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AllowRoutingResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, AllowRoutingResponse_Response>
      _AllowRoutingResponse_ResponseByTag = {
    1: AllowRoutingResponse_Response.empty,
    3: AllowRoutingResponse_Response.allowRoutingErrorCode,
    6: AllowRoutingResponse_Response.updatePeerError,
    0: AllowRoutingResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AllowRoutingResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 3, 6])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<AllowRoutingErrorCode>(
        3, _omitFieldNames ? '' : 'allowRoutingErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: AllowRoutingErrorCode.ROUTING_ALREADY_ALLOWED,
        valueOf: AllowRoutingErrorCode.valueOf,
        enumValues: AllowRoutingErrorCode.values)
    ..aOM<UpdatePeerError>(6, _omitFieldNames ? '' : 'updatePeerError',
        subBuilder: UpdatePeerError.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AllowRoutingResponse clone() =>
      AllowRoutingResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AllowRoutingResponse copyWith(void Function(AllowRoutingResponse) updates) =>
      super.copyWith((message) => updates(message as AllowRoutingResponse))
          as AllowRoutingResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AllowRoutingResponse create() => AllowRoutingResponse._();
  @$core.override
  AllowRoutingResponse createEmptyInstance() => create();
  static $pb.PbList<AllowRoutingResponse> createRepeated() =>
      $pb.PbList<AllowRoutingResponse>();
  @$core.pragma('dart2js:noInline')
  static AllowRoutingResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AllowRoutingResponse>(create);
  static AllowRoutingResponse? _defaultInstance;

  AllowRoutingResponse_Response whichResponse() =>
      _AllowRoutingResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $0.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($0.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(3)
  AllowRoutingErrorCode get allowRoutingErrorCode => $_getN(1);
  @$pb.TagNumber(3)
  set allowRoutingErrorCode(AllowRoutingErrorCode value) =>
      $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasAllowRoutingErrorCode() => $_has(1);
  @$pb.TagNumber(3)
  void clearAllowRoutingErrorCode() => $_clearField(3);

  @$pb.TagNumber(6)
  UpdatePeerError get updatePeerError => $_getN(2);
  @$pb.TagNumber(6)
  set updatePeerError(UpdatePeerError value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdatePeerError() => $_has(2);
  @$pb.TagNumber(6)
  void clearUpdatePeerError() => $_clearField(6);
  @$pb.TagNumber(6)
  UpdatePeerError ensureUpdatePeerError() => $_ensure(2);
}

enum DenyRoutingResponse_Response {
  empty,
  denyRoutingErrorCode,
  updatePeerError,
  notSet
}

/// DenyRoutingResponse defines a response for allow routing request
class DenyRoutingResponse extends $pb.GeneratedMessage {
  factory DenyRoutingResponse({
    $0.Empty? empty,
    DenyRoutingErrorCode? denyRoutingErrorCode,
    UpdatePeerError? updatePeerError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (denyRoutingErrorCode != null)
      result.denyRoutingErrorCode = denyRoutingErrorCode;
    if (updatePeerError != null) result.updatePeerError = updatePeerError;
    return result;
  }

  DenyRoutingResponse._();

  factory DenyRoutingResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DenyRoutingResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, DenyRoutingResponse_Response>
      _DenyRoutingResponse_ResponseByTag = {
    1: DenyRoutingResponse_Response.empty,
    3: DenyRoutingResponse_Response.denyRoutingErrorCode,
    6: DenyRoutingResponse_Response.updatePeerError,
    0: DenyRoutingResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DenyRoutingResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 3, 6])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<DenyRoutingErrorCode>(
        3, _omitFieldNames ? '' : 'denyRoutingErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: DenyRoutingErrorCode.ROUTING_ALREADY_DENIED,
        valueOf: DenyRoutingErrorCode.valueOf,
        enumValues: DenyRoutingErrorCode.values)
    ..aOM<UpdatePeerError>(6, _omitFieldNames ? '' : 'updatePeerError',
        subBuilder: UpdatePeerError.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DenyRoutingResponse clone() => DenyRoutingResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DenyRoutingResponse copyWith(void Function(DenyRoutingResponse) updates) =>
      super.copyWith((message) => updates(message as DenyRoutingResponse))
          as DenyRoutingResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DenyRoutingResponse create() => DenyRoutingResponse._();
  @$core.override
  DenyRoutingResponse createEmptyInstance() => create();
  static $pb.PbList<DenyRoutingResponse> createRepeated() =>
      $pb.PbList<DenyRoutingResponse>();
  @$core.pragma('dart2js:noInline')
  static DenyRoutingResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DenyRoutingResponse>(create);
  static DenyRoutingResponse? _defaultInstance;

  DenyRoutingResponse_Response whichResponse() =>
      _DenyRoutingResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $0.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($0.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(3)
  DenyRoutingErrorCode get denyRoutingErrorCode => $_getN(1);
  @$pb.TagNumber(3)
  set denyRoutingErrorCode(DenyRoutingErrorCode value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasDenyRoutingErrorCode() => $_has(1);
  @$pb.TagNumber(3)
  void clearDenyRoutingErrorCode() => $_clearField(3);

  @$pb.TagNumber(6)
  UpdatePeerError get updatePeerError => $_getN(2);
  @$pb.TagNumber(6)
  set updatePeerError(UpdatePeerError value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdatePeerError() => $_has(2);
  @$pb.TagNumber(6)
  void clearUpdatePeerError() => $_clearField(6);
  @$pb.TagNumber(6)
  UpdatePeerError ensureUpdatePeerError() => $_ensure(2);
}

enum AllowIncomingResponse_Response {
  empty,
  allowIncomingErrorCode,
  updatePeerError,
  notSet
}

/// AllowIncomingResponse defines a response for allow incoming
/// traffic request
class AllowIncomingResponse extends $pb.GeneratedMessage {
  factory AllowIncomingResponse({
    $0.Empty? empty,
    AllowIncomingErrorCode? allowIncomingErrorCode,
    UpdatePeerError? updatePeerError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (allowIncomingErrorCode != null)
      result.allowIncomingErrorCode = allowIncomingErrorCode;
    if (updatePeerError != null) result.updatePeerError = updatePeerError;
    return result;
  }

  AllowIncomingResponse._();

  factory AllowIncomingResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AllowIncomingResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, AllowIncomingResponse_Response>
      _AllowIncomingResponse_ResponseByTag = {
    1: AllowIncomingResponse_Response.empty,
    3: AllowIncomingResponse_Response.allowIncomingErrorCode,
    6: AllowIncomingResponse_Response.updatePeerError,
    0: AllowIncomingResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AllowIncomingResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 3, 6])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<AllowIncomingErrorCode>(
        3, _omitFieldNames ? '' : 'allowIncomingErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: AllowIncomingErrorCode.INCOMING_ALREADY_ALLOWED,
        valueOf: AllowIncomingErrorCode.valueOf,
        enumValues: AllowIncomingErrorCode.values)
    ..aOM<UpdatePeerError>(6, _omitFieldNames ? '' : 'updatePeerError',
        subBuilder: UpdatePeerError.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AllowIncomingResponse clone() =>
      AllowIncomingResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AllowIncomingResponse copyWith(
          void Function(AllowIncomingResponse) updates) =>
      super.copyWith((message) => updates(message as AllowIncomingResponse))
          as AllowIncomingResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AllowIncomingResponse create() => AllowIncomingResponse._();
  @$core.override
  AllowIncomingResponse createEmptyInstance() => create();
  static $pb.PbList<AllowIncomingResponse> createRepeated() =>
      $pb.PbList<AllowIncomingResponse>();
  @$core.pragma('dart2js:noInline')
  static AllowIncomingResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AllowIncomingResponse>(create);
  static AllowIncomingResponse? _defaultInstance;

  AllowIncomingResponse_Response whichResponse() =>
      _AllowIncomingResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $0.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($0.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(3)
  AllowIncomingErrorCode get allowIncomingErrorCode => $_getN(1);
  @$pb.TagNumber(3)
  set allowIncomingErrorCode(AllowIncomingErrorCode value) =>
      $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasAllowIncomingErrorCode() => $_has(1);
  @$pb.TagNumber(3)
  void clearAllowIncomingErrorCode() => $_clearField(3);

  @$pb.TagNumber(6)
  UpdatePeerError get updatePeerError => $_getN(2);
  @$pb.TagNumber(6)
  set updatePeerError(UpdatePeerError value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdatePeerError() => $_has(2);
  @$pb.TagNumber(6)
  void clearUpdatePeerError() => $_clearField(6);
  @$pb.TagNumber(6)
  UpdatePeerError ensureUpdatePeerError() => $_ensure(2);
}

enum DenyIncomingResponse_Response {
  empty,
  denyIncomingErrorCode,
  updatePeerError,
  notSet
}

/// DenyIncomingResponse defines a response for deny incoming
/// traffic request
class DenyIncomingResponse extends $pb.GeneratedMessage {
  factory DenyIncomingResponse({
    $0.Empty? empty,
    DenyIncomingErrorCode? denyIncomingErrorCode,
    UpdatePeerError? updatePeerError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (denyIncomingErrorCode != null)
      result.denyIncomingErrorCode = denyIncomingErrorCode;
    if (updatePeerError != null) result.updatePeerError = updatePeerError;
    return result;
  }

  DenyIncomingResponse._();

  factory DenyIncomingResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DenyIncomingResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, DenyIncomingResponse_Response>
      _DenyIncomingResponse_ResponseByTag = {
    1: DenyIncomingResponse_Response.empty,
    3: DenyIncomingResponse_Response.denyIncomingErrorCode,
    6: DenyIncomingResponse_Response.updatePeerError,
    0: DenyIncomingResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DenyIncomingResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 3, 6])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<DenyIncomingErrorCode>(
        3, _omitFieldNames ? '' : 'denyIncomingErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: DenyIncomingErrorCode.INCOMING_ALREADY_DENIED,
        valueOf: DenyIncomingErrorCode.valueOf,
        enumValues: DenyIncomingErrorCode.values)
    ..aOM<UpdatePeerError>(6, _omitFieldNames ? '' : 'updatePeerError',
        subBuilder: UpdatePeerError.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DenyIncomingResponse clone() =>
      DenyIncomingResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DenyIncomingResponse copyWith(void Function(DenyIncomingResponse) updates) =>
      super.copyWith((message) => updates(message as DenyIncomingResponse))
          as DenyIncomingResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DenyIncomingResponse create() => DenyIncomingResponse._();
  @$core.override
  DenyIncomingResponse createEmptyInstance() => create();
  static $pb.PbList<DenyIncomingResponse> createRepeated() =>
      $pb.PbList<DenyIncomingResponse>();
  @$core.pragma('dart2js:noInline')
  static DenyIncomingResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DenyIncomingResponse>(create);
  static DenyIncomingResponse? _defaultInstance;

  DenyIncomingResponse_Response whichResponse() =>
      _DenyIncomingResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $0.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($0.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(3)
  DenyIncomingErrorCode get denyIncomingErrorCode => $_getN(1);
  @$pb.TagNumber(3)
  set denyIncomingErrorCode(DenyIncomingErrorCode value) =>
      $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasDenyIncomingErrorCode() => $_has(1);
  @$pb.TagNumber(3)
  void clearDenyIncomingErrorCode() => $_clearField(3);

  @$pb.TagNumber(6)
  UpdatePeerError get updatePeerError => $_getN(2);
  @$pb.TagNumber(6)
  set updatePeerError(UpdatePeerError value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdatePeerError() => $_has(2);
  @$pb.TagNumber(6)
  void clearUpdatePeerError() => $_clearField(6);
  @$pb.TagNumber(6)
  UpdatePeerError ensureUpdatePeerError() => $_ensure(2);
}

enum AllowLocalNetworkResponse_Response {
  empty,
  allowLocalNetworkErrorCode,
  updatePeerError,
  notSet
}

/// AllowLocalNetworkResponse defines a response for allow local network
/// traffic request
class AllowLocalNetworkResponse extends $pb.GeneratedMessage {
  factory AllowLocalNetworkResponse({
    $0.Empty? empty,
    AllowLocalNetworkErrorCode? allowLocalNetworkErrorCode,
    UpdatePeerError? updatePeerError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (allowLocalNetworkErrorCode != null)
      result.allowLocalNetworkErrorCode = allowLocalNetworkErrorCode;
    if (updatePeerError != null) result.updatePeerError = updatePeerError;
    return result;
  }

  AllowLocalNetworkResponse._();

  factory AllowLocalNetworkResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AllowLocalNetworkResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, AllowLocalNetworkResponse_Response>
      _AllowLocalNetworkResponse_ResponseByTag = {
    1: AllowLocalNetworkResponse_Response.empty,
    3: AllowLocalNetworkResponse_Response.allowLocalNetworkErrorCode,
    6: AllowLocalNetworkResponse_Response.updatePeerError,
    0: AllowLocalNetworkResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AllowLocalNetworkResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 3, 6])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<AllowLocalNetworkErrorCode>(3,
        _omitFieldNames ? '' : 'allowLocalNetworkErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker:
            AllowLocalNetworkErrorCode.LOCAL_NETWORK_ALREADY_ALLOWED,
        valueOf: AllowLocalNetworkErrorCode.valueOf,
        enumValues: AllowLocalNetworkErrorCode.values)
    ..aOM<UpdatePeerError>(6, _omitFieldNames ? '' : 'updatePeerError',
        subBuilder: UpdatePeerError.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AllowLocalNetworkResponse clone() =>
      AllowLocalNetworkResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AllowLocalNetworkResponse copyWith(
          void Function(AllowLocalNetworkResponse) updates) =>
      super.copyWith((message) => updates(message as AllowLocalNetworkResponse))
          as AllowLocalNetworkResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AllowLocalNetworkResponse create() => AllowLocalNetworkResponse._();
  @$core.override
  AllowLocalNetworkResponse createEmptyInstance() => create();
  static $pb.PbList<AllowLocalNetworkResponse> createRepeated() =>
      $pb.PbList<AllowLocalNetworkResponse>();
  @$core.pragma('dart2js:noInline')
  static AllowLocalNetworkResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AllowLocalNetworkResponse>(create);
  static AllowLocalNetworkResponse? _defaultInstance;

  AllowLocalNetworkResponse_Response whichResponse() =>
      _AllowLocalNetworkResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $0.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($0.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(3)
  AllowLocalNetworkErrorCode get allowLocalNetworkErrorCode => $_getN(1);
  @$pb.TagNumber(3)
  set allowLocalNetworkErrorCode(AllowLocalNetworkErrorCode value) =>
      $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasAllowLocalNetworkErrorCode() => $_has(1);
  @$pb.TagNumber(3)
  void clearAllowLocalNetworkErrorCode() => $_clearField(3);

  @$pb.TagNumber(6)
  UpdatePeerError get updatePeerError => $_getN(2);
  @$pb.TagNumber(6)
  set updatePeerError(UpdatePeerError value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdatePeerError() => $_has(2);
  @$pb.TagNumber(6)
  void clearUpdatePeerError() => $_clearField(6);
  @$pb.TagNumber(6)
  UpdatePeerError ensureUpdatePeerError() => $_ensure(2);
}

enum DenyLocalNetworkResponse_Response {
  empty,
  denyLocalNetworkErrorCode,
  updatePeerError,
  notSet
}

/// DenyIncomingResponse defines a response for deny local network
/// traffic request
class DenyLocalNetworkResponse extends $pb.GeneratedMessage {
  factory DenyLocalNetworkResponse({
    $0.Empty? empty,
    DenyLocalNetworkErrorCode? denyLocalNetworkErrorCode,
    UpdatePeerError? updatePeerError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (denyLocalNetworkErrorCode != null)
      result.denyLocalNetworkErrorCode = denyLocalNetworkErrorCode;
    if (updatePeerError != null) result.updatePeerError = updatePeerError;
    return result;
  }

  DenyLocalNetworkResponse._();

  factory DenyLocalNetworkResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DenyLocalNetworkResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, DenyLocalNetworkResponse_Response>
      _DenyLocalNetworkResponse_ResponseByTag = {
    1: DenyLocalNetworkResponse_Response.empty,
    3: DenyLocalNetworkResponse_Response.denyLocalNetworkErrorCode,
    6: DenyLocalNetworkResponse_Response.updatePeerError,
    0: DenyLocalNetworkResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DenyLocalNetworkResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 3, 6])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<DenyLocalNetworkErrorCode>(3,
        _omitFieldNames ? '' : 'denyLocalNetworkErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: DenyLocalNetworkErrorCode.LOCAL_NETWORK_ALREADY_DENIED,
        valueOf: DenyLocalNetworkErrorCode.valueOf,
        enumValues: DenyLocalNetworkErrorCode.values)
    ..aOM<UpdatePeerError>(6, _omitFieldNames ? '' : 'updatePeerError',
        subBuilder: UpdatePeerError.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DenyLocalNetworkResponse clone() =>
      DenyLocalNetworkResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DenyLocalNetworkResponse copyWith(
          void Function(DenyLocalNetworkResponse) updates) =>
      super.copyWith((message) => updates(message as DenyLocalNetworkResponse))
          as DenyLocalNetworkResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DenyLocalNetworkResponse create() => DenyLocalNetworkResponse._();
  @$core.override
  DenyLocalNetworkResponse createEmptyInstance() => create();
  static $pb.PbList<DenyLocalNetworkResponse> createRepeated() =>
      $pb.PbList<DenyLocalNetworkResponse>();
  @$core.pragma('dart2js:noInline')
  static DenyLocalNetworkResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DenyLocalNetworkResponse>(create);
  static DenyLocalNetworkResponse? _defaultInstance;

  DenyLocalNetworkResponse_Response whichResponse() =>
      _DenyLocalNetworkResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $0.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($0.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(3)
  DenyLocalNetworkErrorCode get denyLocalNetworkErrorCode => $_getN(1);
  @$pb.TagNumber(3)
  set denyLocalNetworkErrorCode(DenyLocalNetworkErrorCode value) =>
      $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasDenyLocalNetworkErrorCode() => $_has(1);
  @$pb.TagNumber(3)
  void clearDenyLocalNetworkErrorCode() => $_clearField(3);

  @$pb.TagNumber(6)
  UpdatePeerError get updatePeerError => $_getN(2);
  @$pb.TagNumber(6)
  set updatePeerError(UpdatePeerError value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdatePeerError() => $_has(2);
  @$pb.TagNumber(6)
  void clearUpdatePeerError() => $_clearField(6);
  @$pb.TagNumber(6)
  UpdatePeerError ensureUpdatePeerError() => $_ensure(2);
}

enum AllowFileshareResponse_Response {
  empty,
  allowSendErrorCode,
  updatePeerError,
  notSet
}

/// AllowSendFileResponse defines a response for allow send file request
class AllowFileshareResponse extends $pb.GeneratedMessage {
  factory AllowFileshareResponse({
    $0.Empty? empty,
    AllowFileshareErrorCode? allowSendErrorCode,
    UpdatePeerError? updatePeerError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (allowSendErrorCode != null)
      result.allowSendErrorCode = allowSendErrorCode;
    if (updatePeerError != null) result.updatePeerError = updatePeerError;
    return result;
  }

  AllowFileshareResponse._();

  factory AllowFileshareResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AllowFileshareResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, AllowFileshareResponse_Response>
      _AllowFileshareResponse_ResponseByTag = {
    1: AllowFileshareResponse_Response.empty,
    3: AllowFileshareResponse_Response.allowSendErrorCode,
    6: AllowFileshareResponse_Response.updatePeerError,
    0: AllowFileshareResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AllowFileshareResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 3, 6])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<AllowFileshareErrorCode>(
        3, _omitFieldNames ? '' : 'allowSendErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: AllowFileshareErrorCode.SEND_ALREADY_ALLOWED,
        valueOf: AllowFileshareErrorCode.valueOf,
        enumValues: AllowFileshareErrorCode.values)
    ..aOM<UpdatePeerError>(6, _omitFieldNames ? '' : 'updatePeerError',
        subBuilder: UpdatePeerError.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AllowFileshareResponse clone() =>
      AllowFileshareResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AllowFileshareResponse copyWith(
          void Function(AllowFileshareResponse) updates) =>
      super.copyWith((message) => updates(message as AllowFileshareResponse))
          as AllowFileshareResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AllowFileshareResponse create() => AllowFileshareResponse._();
  @$core.override
  AllowFileshareResponse createEmptyInstance() => create();
  static $pb.PbList<AllowFileshareResponse> createRepeated() =>
      $pb.PbList<AllowFileshareResponse>();
  @$core.pragma('dart2js:noInline')
  static AllowFileshareResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AllowFileshareResponse>(create);
  static AllowFileshareResponse? _defaultInstance;

  AllowFileshareResponse_Response whichResponse() =>
      _AllowFileshareResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $0.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($0.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(3)
  AllowFileshareErrorCode get allowSendErrorCode => $_getN(1);
  @$pb.TagNumber(3)
  set allowSendErrorCode(AllowFileshareErrorCode value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasAllowSendErrorCode() => $_has(1);
  @$pb.TagNumber(3)
  void clearAllowSendErrorCode() => $_clearField(3);

  @$pb.TagNumber(6)
  UpdatePeerError get updatePeerError => $_getN(2);
  @$pb.TagNumber(6)
  set updatePeerError(UpdatePeerError value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdatePeerError() => $_has(2);
  @$pb.TagNumber(6)
  void clearUpdatePeerError() => $_clearField(6);
  @$pb.TagNumber(6)
  UpdatePeerError ensureUpdatePeerError() => $_ensure(2);
}

enum DenyFileshareResponse_Response {
  empty,
  denySendErrorCode,
  updatePeerError,
  notSet
}

/// DenySendFileResponse defines a response for deny send file request
class DenyFileshareResponse extends $pb.GeneratedMessage {
  factory DenyFileshareResponse({
    $0.Empty? empty,
    DenyFileshareErrorCode? denySendErrorCode,
    UpdatePeerError? updatePeerError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (denySendErrorCode != null) result.denySendErrorCode = denySendErrorCode;
    if (updatePeerError != null) result.updatePeerError = updatePeerError;
    return result;
  }

  DenyFileshareResponse._();

  factory DenyFileshareResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DenyFileshareResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, DenyFileshareResponse_Response>
      _DenyFileshareResponse_ResponseByTag = {
    1: DenyFileshareResponse_Response.empty,
    3: DenyFileshareResponse_Response.denySendErrorCode,
    6: DenyFileshareResponse_Response.updatePeerError,
    0: DenyFileshareResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DenyFileshareResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 3, 6])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<DenyFileshareErrorCode>(
        3, _omitFieldNames ? '' : 'denySendErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: DenyFileshareErrorCode.SEND_ALREADY_DENIED,
        valueOf: DenyFileshareErrorCode.valueOf,
        enumValues: DenyFileshareErrorCode.values)
    ..aOM<UpdatePeerError>(6, _omitFieldNames ? '' : 'updatePeerError',
        subBuilder: UpdatePeerError.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DenyFileshareResponse clone() =>
      DenyFileshareResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DenyFileshareResponse copyWith(
          void Function(DenyFileshareResponse) updates) =>
      super.copyWith((message) => updates(message as DenyFileshareResponse))
          as DenyFileshareResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DenyFileshareResponse create() => DenyFileshareResponse._();
  @$core.override
  DenyFileshareResponse createEmptyInstance() => create();
  static $pb.PbList<DenyFileshareResponse> createRepeated() =>
      $pb.PbList<DenyFileshareResponse>();
  @$core.pragma('dart2js:noInline')
  static DenyFileshareResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DenyFileshareResponse>(create);
  static DenyFileshareResponse? _defaultInstance;

  DenyFileshareResponse_Response whichResponse() =>
      _DenyFileshareResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $0.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($0.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(3)
  DenyFileshareErrorCode get denySendErrorCode => $_getN(1);
  @$pb.TagNumber(3)
  set denySendErrorCode(DenyFileshareErrorCode value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasDenySendErrorCode() => $_has(1);
  @$pb.TagNumber(3)
  void clearDenySendErrorCode() => $_clearField(3);

  @$pb.TagNumber(6)
  UpdatePeerError get updatePeerError => $_getN(2);
  @$pb.TagNumber(6)
  set updatePeerError(UpdatePeerError value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdatePeerError() => $_has(2);
  @$pb.TagNumber(6)
  void clearUpdatePeerError() => $_clearField(6);
  @$pb.TagNumber(6)
  UpdatePeerError ensureUpdatePeerError() => $_ensure(2);
}

enum EnableAutomaticFileshareResponse_Response {
  empty,
  enableAutomaticFileshareErrorCode,
  updatePeerError,
  notSet
}

/// AllowSendFileResponse defines a response for allow send file request
class EnableAutomaticFileshareResponse extends $pb.GeneratedMessage {
  factory EnableAutomaticFileshareResponse({
    $0.Empty? empty,
    EnableAutomaticFileshareErrorCode? enableAutomaticFileshareErrorCode,
    UpdatePeerError? updatePeerError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (enableAutomaticFileshareErrorCode != null)
      result.enableAutomaticFileshareErrorCode =
          enableAutomaticFileshareErrorCode;
    if (updatePeerError != null) result.updatePeerError = updatePeerError;
    return result;
  }

  EnableAutomaticFileshareResponse._();

  factory EnableAutomaticFileshareResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory EnableAutomaticFileshareResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, EnableAutomaticFileshareResponse_Response>
      _EnableAutomaticFileshareResponse_ResponseByTag = {
    1: EnableAutomaticFileshareResponse_Response.empty,
    3: EnableAutomaticFileshareResponse_Response
        .enableAutomaticFileshareErrorCode,
    6: EnableAutomaticFileshareResponse_Response.updatePeerError,
    0: EnableAutomaticFileshareResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EnableAutomaticFileshareResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 3, 6])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<EnableAutomaticFileshareErrorCode>(
        3,
        _omitFieldNames ? '' : 'enableAutomaticFileshareErrorCode',
        $pb.PbFieldType.OE,
        defaultOrMaker: EnableAutomaticFileshareErrorCode
            .AUTOMATIC_FILESHARE_ALREADY_ENABLED,
        valueOf: EnableAutomaticFileshareErrorCode.valueOf,
        enumValues: EnableAutomaticFileshareErrorCode.values)
    ..aOM<UpdatePeerError>(6, _omitFieldNames ? '' : 'updatePeerError',
        subBuilder: UpdatePeerError.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EnableAutomaticFileshareResponse clone() =>
      EnableAutomaticFileshareResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EnableAutomaticFileshareResponse copyWith(
          void Function(EnableAutomaticFileshareResponse) updates) =>
      super.copyWith(
              (message) => updates(message as EnableAutomaticFileshareResponse))
          as EnableAutomaticFileshareResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static EnableAutomaticFileshareResponse create() =>
      EnableAutomaticFileshareResponse._();
  @$core.override
  EnableAutomaticFileshareResponse createEmptyInstance() => create();
  static $pb.PbList<EnableAutomaticFileshareResponse> createRepeated() =>
      $pb.PbList<EnableAutomaticFileshareResponse>();
  @$core.pragma('dart2js:noInline')
  static EnableAutomaticFileshareResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<EnableAutomaticFileshareResponse>(
          create);
  static EnableAutomaticFileshareResponse? _defaultInstance;

  EnableAutomaticFileshareResponse_Response whichResponse() =>
      _EnableAutomaticFileshareResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $0.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($0.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(3)
  EnableAutomaticFileshareErrorCode get enableAutomaticFileshareErrorCode =>
      $_getN(1);
  @$pb.TagNumber(3)
  set enableAutomaticFileshareErrorCode(
          EnableAutomaticFileshareErrorCode value) =>
      $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasEnableAutomaticFileshareErrorCode() => $_has(1);
  @$pb.TagNumber(3)
  void clearEnableAutomaticFileshareErrorCode() => $_clearField(3);

  @$pb.TagNumber(6)
  UpdatePeerError get updatePeerError => $_getN(2);
  @$pb.TagNumber(6)
  set updatePeerError(UpdatePeerError value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdatePeerError() => $_has(2);
  @$pb.TagNumber(6)
  void clearUpdatePeerError() => $_clearField(6);
  @$pb.TagNumber(6)
  UpdatePeerError ensureUpdatePeerError() => $_ensure(2);
}

enum DisableAutomaticFileshareResponse_Response {
  empty,
  disableAutomaticFileshareErrorCode,
  updatePeerError,
  notSet
}

/// DenySendFileResponse defines a response for deny send file request
class DisableAutomaticFileshareResponse extends $pb.GeneratedMessage {
  factory DisableAutomaticFileshareResponse({
    $0.Empty? empty,
    DisableAutomaticFileshareErrorCode? disableAutomaticFileshareErrorCode,
    UpdatePeerError? updatePeerError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (disableAutomaticFileshareErrorCode != null)
      result.disableAutomaticFileshareErrorCode =
          disableAutomaticFileshareErrorCode;
    if (updatePeerError != null) result.updatePeerError = updatePeerError;
    return result;
  }

  DisableAutomaticFileshareResponse._();

  factory DisableAutomaticFileshareResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DisableAutomaticFileshareResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, DisableAutomaticFileshareResponse_Response>
      _DisableAutomaticFileshareResponse_ResponseByTag = {
    1: DisableAutomaticFileshareResponse_Response.empty,
    3: DisableAutomaticFileshareResponse_Response
        .disableAutomaticFileshareErrorCode,
    6: DisableAutomaticFileshareResponse_Response.updatePeerError,
    0: DisableAutomaticFileshareResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DisableAutomaticFileshareResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 3, 6])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<DisableAutomaticFileshareErrorCode>(
        3,
        _omitFieldNames ? '' : 'disableAutomaticFileshareErrorCode',
        $pb.PbFieldType.OE,
        defaultOrMaker: DisableAutomaticFileshareErrorCode
            .AUTOMATIC_FILESHARE_ALREADY_DISABLED,
        valueOf: DisableAutomaticFileshareErrorCode.valueOf,
        enumValues: DisableAutomaticFileshareErrorCode.values)
    ..aOM<UpdatePeerError>(6, _omitFieldNames ? '' : 'updatePeerError',
        subBuilder: UpdatePeerError.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisableAutomaticFileshareResponse clone() =>
      DisableAutomaticFileshareResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DisableAutomaticFileshareResponse copyWith(
          void Function(DisableAutomaticFileshareResponse) updates) =>
      super.copyWith((message) =>
              updates(message as DisableAutomaticFileshareResponse))
          as DisableAutomaticFileshareResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DisableAutomaticFileshareResponse create() =>
      DisableAutomaticFileshareResponse._();
  @$core.override
  DisableAutomaticFileshareResponse createEmptyInstance() => create();
  static $pb.PbList<DisableAutomaticFileshareResponse> createRepeated() =>
      $pb.PbList<DisableAutomaticFileshareResponse>();
  @$core.pragma('dart2js:noInline')
  static DisableAutomaticFileshareResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DisableAutomaticFileshareResponse>(
          create);
  static DisableAutomaticFileshareResponse? _defaultInstance;

  DisableAutomaticFileshareResponse_Response whichResponse() =>
      _DisableAutomaticFileshareResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $0.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($0.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(3)
  DisableAutomaticFileshareErrorCode get disableAutomaticFileshareErrorCode =>
      $_getN(1);
  @$pb.TagNumber(3)
  set disableAutomaticFileshareErrorCode(
          DisableAutomaticFileshareErrorCode value) =>
      $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasDisableAutomaticFileshareErrorCode() => $_has(1);
  @$pb.TagNumber(3)
  void clearDisableAutomaticFileshareErrorCode() => $_clearField(3);

  @$pb.TagNumber(6)
  UpdatePeerError get updatePeerError => $_getN(2);
  @$pb.TagNumber(6)
  set updatePeerError(UpdatePeerError value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdatePeerError() => $_has(2);
  @$pb.TagNumber(6)
  void clearUpdatePeerError() => $_clearField(6);
  @$pb.TagNumber(6)
  UpdatePeerError ensureUpdatePeerError() => $_ensure(2);
}

enum ConnectResponse_Response {
  empty,
  connectErrorCode,
  updatePeerError,
  notSet
}

class ConnectResponse extends $pb.GeneratedMessage {
  factory ConnectResponse({
    $0.Empty? empty,
    ConnectErrorCode? connectErrorCode,
    UpdatePeerError? updatePeerError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (connectErrorCode != null) result.connectErrorCode = connectErrorCode;
    if (updatePeerError != null) result.updatePeerError = updatePeerError;
    return result;
  }

  ConnectResponse._();

  factory ConnectResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ConnectResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, ConnectResponse_Response>
      _ConnectResponse_ResponseByTag = {
    1: ConnectResponse_Response.empty,
    3: ConnectResponse_Response.connectErrorCode,
    6: ConnectResponse_Response.updatePeerError,
    0: ConnectResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ConnectResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 3, 6])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<ConnectErrorCode>(
        3, _omitFieldNames ? '' : 'connectErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: ConnectErrorCode.PEER_DOES_NOT_ALLOW_ROUTING,
        valueOf: ConnectErrorCode.valueOf,
        enumValues: ConnectErrorCode.values)
    ..aOM<UpdatePeerError>(6, _omitFieldNames ? '' : 'updatePeerError',
        subBuilder: UpdatePeerError.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConnectResponse clone() => ConnectResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConnectResponse copyWith(void Function(ConnectResponse) updates) =>
      super.copyWith((message) => updates(message as ConnectResponse))
          as ConnectResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ConnectResponse create() => ConnectResponse._();
  @$core.override
  ConnectResponse createEmptyInstance() => create();
  static $pb.PbList<ConnectResponse> createRepeated() =>
      $pb.PbList<ConnectResponse>();
  @$core.pragma('dart2js:noInline')
  static ConnectResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ConnectResponse>(create);
  static ConnectResponse? _defaultInstance;

  ConnectResponse_Response whichResponse() =>
      _ConnectResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $0.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($0.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(3)
  ConnectErrorCode get connectErrorCode => $_getN(1);
  @$pb.TagNumber(3)
  set connectErrorCode(ConnectErrorCode value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasConnectErrorCode() => $_has(1);
  @$pb.TagNumber(3)
  void clearConnectErrorCode() => $_clearField(3);

  @$pb.TagNumber(6)
  UpdatePeerError get updatePeerError => $_getN(2);
  @$pb.TagNumber(6)
  set updatePeerError(UpdatePeerError value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdatePeerError() => $_has(2);
  @$pb.TagNumber(6)
  void clearUpdatePeerError() => $_clearField(6);
  @$pb.TagNumber(6)
  UpdatePeerError ensureUpdatePeerError() => $_ensure(2);
}

enum PrivateKeyResponse_Response {
  privateKey,
  serviceErrorCode,
  meshnetErrorCode,
  notSet
}

class PrivateKeyResponse extends $pb.GeneratedMessage {
  factory PrivateKeyResponse({
    $core.String? privateKey,
    $1.ServiceErrorCode? serviceErrorCode,
    $1.MeshnetErrorCode? meshnetErrorCode,
  }) {
    final result = create();
    if (privateKey != null) result.privateKey = privateKey;
    if (serviceErrorCode != null) result.serviceErrorCode = serviceErrorCode;
    if (meshnetErrorCode != null) result.meshnetErrorCode = meshnetErrorCode;
    return result;
  }

  PrivateKeyResponse._();

  factory PrivateKeyResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory PrivateKeyResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, PrivateKeyResponse_Response>
      _PrivateKeyResponse_ResponseByTag = {
    1: PrivateKeyResponse_Response.privateKey,
    2: PrivateKeyResponse_Response.serviceErrorCode,
    3: PrivateKeyResponse_Response.meshnetErrorCode,
    0: PrivateKeyResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PrivateKeyResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2, 3])
    ..aOS(1, _omitFieldNames ? '' : 'privateKey')
    ..e<$1.ServiceErrorCode>(
        2, _omitFieldNames ? '' : 'serviceErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: $1.ServiceErrorCode.NOT_LOGGED_IN,
        valueOf: $1.ServiceErrorCode.valueOf,
        enumValues: $1.ServiceErrorCode.values)
    ..e<$1.MeshnetErrorCode>(
        3, _omitFieldNames ? '' : 'meshnetErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: $1.MeshnetErrorCode.NOT_REGISTERED,
        valueOf: $1.MeshnetErrorCode.valueOf,
        enumValues: $1.MeshnetErrorCode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PrivateKeyResponse clone() => PrivateKeyResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PrivateKeyResponse copyWith(void Function(PrivateKeyResponse) updates) =>
      super.copyWith((message) => updates(message as PrivateKeyResponse))
          as PrivateKeyResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PrivateKeyResponse create() => PrivateKeyResponse._();
  @$core.override
  PrivateKeyResponse createEmptyInstance() => create();
  static $pb.PbList<PrivateKeyResponse> createRepeated() =>
      $pb.PbList<PrivateKeyResponse>();
  @$core.pragma('dart2js:noInline')
  static PrivateKeyResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PrivateKeyResponse>(create);
  static PrivateKeyResponse? _defaultInstance;

  PrivateKeyResponse_Response whichResponse() =>
      _PrivateKeyResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $core.String get privateKey => $_getSZ(0);
  @$pb.TagNumber(1)
  set privateKey($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPrivateKey() => $_has(0);
  @$pb.TagNumber(1)
  void clearPrivateKey() => $_clearField(1);

  @$pb.TagNumber(2)
  $1.ServiceErrorCode get serviceErrorCode => $_getN(1);
  @$pb.TagNumber(2)
  set serviceErrorCode($1.ServiceErrorCode value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasServiceErrorCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearServiceErrorCode() => $_clearField(2);

  @$pb.TagNumber(3)
  $1.MeshnetErrorCode get meshnetErrorCode => $_getN(2);
  @$pb.TagNumber(3)
  set meshnetErrorCode($1.MeshnetErrorCode value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasMeshnetErrorCode() => $_has(2);
  @$pb.TagNumber(3)
  void clearMeshnetErrorCode() => $_clearField(3);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
