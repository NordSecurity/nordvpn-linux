// This is a generated file - do not edit.
//
// Generated from invite.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'empty.pb.dart' as $1;
import 'google/protobuf/timestamp.pb.dart' as $0;
import 'invite.pbenum.dart';
import 'service_response.pbenum.dart' as $2;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'invite.pbenum.dart';

enum GetInvitesResponse_Response {
  invites,
  serviceErrorCode,
  meshnetErrorCode,
  notSet
}

/// GetInvitesResponse defines a response for GetInvites request
class GetInvitesResponse extends $pb.GeneratedMessage {
  factory GetInvitesResponse({
    InvitesList? invites,
    $2.ServiceErrorCode? serviceErrorCode,
    $2.MeshnetErrorCode? meshnetErrorCode,
  }) {
    final result = create();
    if (invites != null) result.invites = invites;
    if (serviceErrorCode != null) result.serviceErrorCode = serviceErrorCode;
    if (meshnetErrorCode != null) result.meshnetErrorCode = meshnetErrorCode;
    return result;
  }

  GetInvitesResponse._();

  factory GetInvitesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetInvitesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, GetInvitesResponse_Response>
      _GetInvitesResponse_ResponseByTag = {
    1: GetInvitesResponse_Response.invites,
    2: GetInvitesResponse_Response.serviceErrorCode,
    3: GetInvitesResponse_Response.meshnetErrorCode,
    0: GetInvitesResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetInvitesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2, 3])
    ..aOM<InvitesList>(1, _omitFieldNames ? '' : 'invites',
        subBuilder: InvitesList.create)
    ..e<$2.ServiceErrorCode>(
        2, _omitFieldNames ? '' : 'serviceErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: $2.ServiceErrorCode.NOT_LOGGED_IN,
        valueOf: $2.ServiceErrorCode.valueOf,
        enumValues: $2.ServiceErrorCode.values)
    ..e<$2.MeshnetErrorCode>(
        3, _omitFieldNames ? '' : 'meshnetErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: $2.MeshnetErrorCode.NOT_REGISTERED,
        valueOf: $2.MeshnetErrorCode.valueOf,
        enumValues: $2.MeshnetErrorCode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetInvitesResponse clone() => GetInvitesResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetInvitesResponse copyWith(void Function(GetInvitesResponse) updates) =>
      super.copyWith((message) => updates(message as GetInvitesResponse))
          as GetInvitesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetInvitesResponse create() => GetInvitesResponse._();
  @$core.override
  GetInvitesResponse createEmptyInstance() => create();
  static $pb.PbList<GetInvitesResponse> createRepeated() =>
      $pb.PbList<GetInvitesResponse>();
  @$core.pragma('dart2js:noInline')
  static GetInvitesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetInvitesResponse>(create);
  static GetInvitesResponse? _defaultInstance;

  GetInvitesResponse_Response whichResponse() =>
      _GetInvitesResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  InvitesList get invites => $_getN(0);
  @$pb.TagNumber(1)
  set invites(InvitesList value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasInvites() => $_has(0);
  @$pb.TagNumber(1)
  void clearInvites() => $_clearField(1);
  @$pb.TagNumber(1)
  InvitesList ensureInvites() => $_ensure(0);

  @$pb.TagNumber(2)
  $2.ServiceErrorCode get serviceErrorCode => $_getN(1);
  @$pb.TagNumber(2)
  set serviceErrorCode($2.ServiceErrorCode value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasServiceErrorCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearServiceErrorCode() => $_clearField(2);

  @$pb.TagNumber(3)
  $2.MeshnetErrorCode get meshnetErrorCode => $_getN(2);
  @$pb.TagNumber(3)
  set meshnetErrorCode($2.MeshnetErrorCode value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasMeshnetErrorCode() => $_has(2);
  @$pb.TagNumber(3)
  void clearMeshnetErrorCode() => $_clearField(3);
}

/// InvitesList defines the list of sent and received invitations
class InvitesList extends $pb.GeneratedMessage {
  factory InvitesList({
    $core.Iterable<Invite>? sent,
    $core.Iterable<Invite>? received,
  }) {
    final result = create();
    if (sent != null) result.sent.addAll(sent);
    if (received != null) result.received.addAll(received);
    return result;
  }

  InvitesList._();

  factory InvitesList.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory InvitesList.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'InvitesList',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..pc<Invite>(1, _omitFieldNames ? '' : 'sent', $pb.PbFieldType.PM,
        subBuilder: Invite.create)
    ..pc<Invite>(2, _omitFieldNames ? '' : 'received', $pb.PbFieldType.PM,
        subBuilder: Invite.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InvitesList clone() => InvitesList()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InvitesList copyWith(void Function(InvitesList) updates) =>
      super.copyWith((message) => updates(message as InvitesList))
          as InvitesList;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static InvitesList create() => InvitesList._();
  @$core.override
  InvitesList createEmptyInstance() => create();
  static $pb.PbList<InvitesList> createRepeated() => $pb.PbList<InvitesList>();
  @$core.pragma('dart2js:noInline')
  static InvitesList getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<InvitesList>(create);
  static InvitesList? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Invite> get sent => $_getList(0);

  @$pb.TagNumber(2)
  $pb.PbList<Invite> get received => $_getList(1);
}

/// Invite defines the structure of the meshnet invite
class Invite extends $pb.GeneratedMessage {
  factory Invite({
    $core.String? email,
    $0.Timestamp? expiresAt,
    $core.String? os,
  }) {
    final result = create();
    if (email != null) result.email = email;
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (os != null) result.os = os;
    return result;
  }

  Invite._();

  factory Invite.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Invite.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Invite',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'email')
    ..aOM<$0.Timestamp>(2, _omitFieldNames ? '' : 'expiresAt',
        subBuilder: $0.Timestamp.create)
    ..aOS(3, _omitFieldNames ? '' : 'os')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Invite clone() => Invite()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Invite copyWith(void Function(Invite) updates) =>
      super.copyWith((message) => updates(message as Invite)) as Invite;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Invite create() => Invite._();
  @$core.override
  Invite createEmptyInstance() => create();
  static $pb.PbList<Invite> createRepeated() => $pb.PbList<Invite>();
  @$core.pragma('dart2js:noInline')
  static Invite getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Invite>(create);
  static Invite? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get email => $_getSZ(0);
  @$pb.TagNumber(1)
  set email($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEmail() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmail() => $_clearField(1);

  @$pb.TagNumber(2)
  $0.Timestamp get expiresAt => $_getN(1);
  @$pb.TagNumber(2)
  set expiresAt($0.Timestamp value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasExpiresAt() => $_has(1);
  @$pb.TagNumber(2)
  void clearExpiresAt() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Timestamp ensureExpiresAt() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.String get os => $_getSZ(2);
  @$pb.TagNumber(3)
  set os($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasOs() => $_has(2);
  @$pb.TagNumber(3)
  void clearOs() => $_clearField(3);
}

/// InviteRequest defines an accepting response request for a
/// meshnet invitation
/// InviteRequest is the same as the accepting to the invitation.
/// Both specify the email and allow traffic flags
class InviteRequest extends $pb.GeneratedMessage {
  factory InviteRequest({
    $core.String? email,
    $core.bool? allowIncomingTraffic,
    $core.bool? allowTrafficRouting,
    $core.bool? allowLocalNetwork,
    $core.bool? allowFileshare,
  }) {
    final result = create();
    if (email != null) result.email = email;
    if (allowIncomingTraffic != null)
      result.allowIncomingTraffic = allowIncomingTraffic;
    if (allowTrafficRouting != null)
      result.allowTrafficRouting = allowTrafficRouting;
    if (allowLocalNetwork != null) result.allowLocalNetwork = allowLocalNetwork;
    if (allowFileshare != null) result.allowFileshare = allowFileshare;
    return result;
  }

  InviteRequest._();

  factory InviteRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory InviteRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'InviteRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'email')
    ..aOB(2, _omitFieldNames ? '' : 'allowIncomingTraffic',
        protoName: 'allowIncomingTraffic')
    ..aOB(3, _omitFieldNames ? '' : 'allowTrafficRouting',
        protoName: 'allowTrafficRouting')
    ..aOB(4, _omitFieldNames ? '' : 'allowLocalNetwork',
        protoName: 'allowLocalNetwork')
    ..aOB(5, _omitFieldNames ? '' : 'allowFileshare',
        protoName: 'allowFileshare')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InviteRequest clone() => InviteRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InviteRequest copyWith(void Function(InviteRequest) updates) =>
      super.copyWith((message) => updates(message as InviteRequest))
          as InviteRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static InviteRequest create() => InviteRequest._();
  @$core.override
  InviteRequest createEmptyInstance() => create();
  static $pb.PbList<InviteRequest> createRepeated() =>
      $pb.PbList<InviteRequest>();
  @$core.pragma('dart2js:noInline')
  static InviteRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<InviteRequest>(create);
  static InviteRequest? _defaultInstance;

  /// email is the email of the invitation sender
  @$pb.TagNumber(1)
  $core.String get email => $_getSZ(0);
  @$pb.TagNumber(1)
  set email($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEmail() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmail() => $_clearField(1);

  /// allowIncomingTraffic defines that another peer is allowed
  /// to send traffic to this device
  @$pb.TagNumber(2)
  $core.bool get allowIncomingTraffic => $_getBF(1);
  @$pb.TagNumber(2)
  set allowIncomingTraffic($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAllowIncomingTraffic() => $_has(1);
  @$pb.TagNumber(2)
  void clearAllowIncomingTraffic() => $_clearField(2);

  /// AllowTrafficRouting defines that another peer is allowed to
  /// route traffic through this device
  @$pb.TagNumber(3)
  $core.bool get allowTrafficRouting => $_getBF(2);
  @$pb.TagNumber(3)
  set allowTrafficRouting($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAllowTrafficRouting() => $_has(2);
  @$pb.TagNumber(3)
  void clearAllowTrafficRouting() => $_clearField(3);

  /// AllowLocalNetwork defines that another peer is allowed to
  /// access device's local network when routing traffic through this device
  @$pb.TagNumber(4)
  $core.bool get allowLocalNetwork => $_getBF(3);
  @$pb.TagNumber(4)
  set allowLocalNetwork($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAllowLocalNetwork() => $_has(3);
  @$pb.TagNumber(4)
  void clearAllowLocalNetwork() => $_clearField(4);

  /// AllowLocalNetwork defines that another peer is allowed to send files to this device
  @$pb.TagNumber(5)
  $core.bool get allowFileshare => $_getBF(4);
  @$pb.TagNumber(5)
  set allowFileshare($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAllowFileshare() => $_has(4);
  @$pb.TagNumber(5)
  void clearAllowFileshare() => $_clearField(5);
}

/// DenyInviteRequest defines a denying response request for a meshnet
/// invitation
class DenyInviteRequest extends $pb.GeneratedMessage {
  factory DenyInviteRequest({
    $core.String? email,
  }) {
    final result = create();
    if (email != null) result.email = email;
    return result;
  }

  DenyInviteRequest._();

  factory DenyInviteRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DenyInviteRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DenyInviteRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'email')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DenyInviteRequest clone() => DenyInviteRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DenyInviteRequest copyWith(void Function(DenyInviteRequest) updates) =>
      super.copyWith((message) => updates(message as DenyInviteRequest))
          as DenyInviteRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DenyInviteRequest create() => DenyInviteRequest._();
  @$core.override
  DenyInviteRequest createEmptyInstance() => create();
  static $pb.PbList<DenyInviteRequest> createRepeated() =>
      $pb.PbList<DenyInviteRequest>();
  @$core.pragma('dart2js:noInline')
  static DenyInviteRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DenyInviteRequest>(create);
  static DenyInviteRequest? _defaultInstance;

  /// email is the email of the invitation sender
  @$pb.TagNumber(1)
  $core.String get email => $_getSZ(0);
  @$pb.TagNumber(1)
  set email($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEmail() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmail() => $_clearField(1);
}

enum RespondToInviteResponse_Response {
  empty,
  respondToInviteErrorCode,
  serviceErrorCode,
  meshnetErrorCode,
  notSet
}

/// RespondToInviteResponse defines an empty gRPC response with the
/// status
class RespondToInviteResponse extends $pb.GeneratedMessage {
  factory RespondToInviteResponse({
    $1.Empty? empty,
    RespondToInviteErrorCode? respondToInviteErrorCode,
    $2.ServiceErrorCode? serviceErrorCode,
    $2.MeshnetErrorCode? meshnetErrorCode,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (respondToInviteErrorCode != null)
      result.respondToInviteErrorCode = respondToInviteErrorCode;
    if (serviceErrorCode != null) result.serviceErrorCode = serviceErrorCode;
    if (meshnetErrorCode != null) result.meshnetErrorCode = meshnetErrorCode;
    return result;
  }

  RespondToInviteResponse._();

  factory RespondToInviteResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RespondToInviteResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, RespondToInviteResponse_Response>
      _RespondToInviteResponse_ResponseByTag = {
    1: RespondToInviteResponse_Response.empty,
    2: RespondToInviteResponse_Response.respondToInviteErrorCode,
    3: RespondToInviteResponse_Response.serviceErrorCode,
    4: RespondToInviteResponse_Response.meshnetErrorCode,
    0: RespondToInviteResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RespondToInviteResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2, 3, 4])
    ..aOM<$1.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $1.Empty.create)
    ..e<RespondToInviteErrorCode>(2,
        _omitFieldNames ? '' : 'respondToInviteErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: RespondToInviteErrorCode.UNKNOWN,
        valueOf: RespondToInviteErrorCode.valueOf,
        enumValues: RespondToInviteErrorCode.values)
    ..e<$2.ServiceErrorCode>(
        3, _omitFieldNames ? '' : 'serviceErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: $2.ServiceErrorCode.NOT_LOGGED_IN,
        valueOf: $2.ServiceErrorCode.valueOf,
        enumValues: $2.ServiceErrorCode.values)
    ..e<$2.MeshnetErrorCode>(
        4, _omitFieldNames ? '' : 'meshnetErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: $2.MeshnetErrorCode.NOT_REGISTERED,
        valueOf: $2.MeshnetErrorCode.valueOf,
        enumValues: $2.MeshnetErrorCode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RespondToInviteResponse clone() =>
      RespondToInviteResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RespondToInviteResponse copyWith(
          void Function(RespondToInviteResponse) updates) =>
      super.copyWith((message) => updates(message as RespondToInviteResponse))
          as RespondToInviteResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RespondToInviteResponse create() => RespondToInviteResponse._();
  @$core.override
  RespondToInviteResponse createEmptyInstance() => create();
  static $pb.PbList<RespondToInviteResponse> createRepeated() =>
      $pb.PbList<RespondToInviteResponse>();
  @$core.pragma('dart2js:noInline')
  static RespondToInviteResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RespondToInviteResponse>(create);
  static RespondToInviteResponse? _defaultInstance;

  RespondToInviteResponse_Response whichResponse() =>
      _RespondToInviteResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $1.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($1.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $1.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(2)
  RespondToInviteErrorCode get respondToInviteErrorCode => $_getN(1);
  @$pb.TagNumber(2)
  set respondToInviteErrorCode(RespondToInviteErrorCode value) =>
      $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasRespondToInviteErrorCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearRespondToInviteErrorCode() => $_clearField(2);

  @$pb.TagNumber(3)
  $2.ServiceErrorCode get serviceErrorCode => $_getN(2);
  @$pb.TagNumber(3)
  set serviceErrorCode($2.ServiceErrorCode value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasServiceErrorCode() => $_has(2);
  @$pb.TagNumber(3)
  void clearServiceErrorCode() => $_clearField(3);

  @$pb.TagNumber(4)
  $2.MeshnetErrorCode get meshnetErrorCode => $_getN(3);
  @$pb.TagNumber(4)
  set meshnetErrorCode($2.MeshnetErrorCode value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasMeshnetErrorCode() => $_has(3);
  @$pb.TagNumber(4)
  void clearMeshnetErrorCode() => $_clearField(4);
}

enum InviteResponse_Response {
  empty,
  inviteResponseErrorCode,
  serviceErrorCode,
  meshnetErrorCode,
  notSet
}

/// InviteResponse defines the response to the invite send
class InviteResponse extends $pb.GeneratedMessage {
  factory InviteResponse({
    $1.Empty? empty,
    InviteResponseErrorCode? inviteResponseErrorCode,
    $2.ServiceErrorCode? serviceErrorCode,
    $2.MeshnetErrorCode? meshnetErrorCode,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (inviteResponseErrorCode != null)
      result.inviteResponseErrorCode = inviteResponseErrorCode;
    if (serviceErrorCode != null) result.serviceErrorCode = serviceErrorCode;
    if (meshnetErrorCode != null) result.meshnetErrorCode = meshnetErrorCode;
    return result;
  }

  InviteResponse._();

  factory InviteResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory InviteResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, InviteResponse_Response>
      _InviteResponse_ResponseByTag = {
    1: InviteResponse_Response.empty,
    2: InviteResponse_Response.inviteResponseErrorCode,
    3: InviteResponse_Response.serviceErrorCode,
    4: InviteResponse_Response.meshnetErrorCode,
    0: InviteResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'InviteResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2, 3, 4])
    ..aOM<$1.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $1.Empty.create)
    ..e<InviteResponseErrorCode>(
        2, _omitFieldNames ? '' : 'inviteResponseErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: InviteResponseErrorCode.ALREADY_EXISTS,
        valueOf: InviteResponseErrorCode.valueOf,
        enumValues: InviteResponseErrorCode.values)
    ..e<$2.ServiceErrorCode>(
        3, _omitFieldNames ? '' : 'serviceErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: $2.ServiceErrorCode.NOT_LOGGED_IN,
        valueOf: $2.ServiceErrorCode.valueOf,
        enumValues: $2.ServiceErrorCode.values)
    ..e<$2.MeshnetErrorCode>(
        4, _omitFieldNames ? '' : 'meshnetErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: $2.MeshnetErrorCode.NOT_REGISTERED,
        valueOf: $2.MeshnetErrorCode.valueOf,
        enumValues: $2.MeshnetErrorCode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InviteResponse clone() => InviteResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InviteResponse copyWith(void Function(InviteResponse) updates) =>
      super.copyWith((message) => updates(message as InviteResponse))
          as InviteResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static InviteResponse create() => InviteResponse._();
  @$core.override
  InviteResponse createEmptyInstance() => create();
  static $pb.PbList<InviteResponse> createRepeated() =>
      $pb.PbList<InviteResponse>();
  @$core.pragma('dart2js:noInline')
  static InviteResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<InviteResponse>(create);
  static InviteResponse? _defaultInstance;

  InviteResponse_Response whichResponse() =>
      _InviteResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $1.Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty($1.Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  $1.Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(2)
  InviteResponseErrorCode get inviteResponseErrorCode => $_getN(1);
  @$pb.TagNumber(2)
  set inviteResponseErrorCode(InviteResponseErrorCode value) =>
      $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasInviteResponseErrorCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearInviteResponseErrorCode() => $_clearField(2);

  @$pb.TagNumber(3)
  $2.ServiceErrorCode get serviceErrorCode => $_getN(2);
  @$pb.TagNumber(3)
  set serviceErrorCode($2.ServiceErrorCode value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasServiceErrorCode() => $_has(2);
  @$pb.TagNumber(3)
  void clearServiceErrorCode() => $_clearField(3);

  @$pb.TagNumber(4)
  $2.MeshnetErrorCode get meshnetErrorCode => $_getN(3);
  @$pb.TagNumber(4)
  set meshnetErrorCode($2.MeshnetErrorCode value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasMeshnetErrorCode() => $_has(3);
  @$pb.TagNumber(4)
  void clearMeshnetErrorCode() => $_clearField(4);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
