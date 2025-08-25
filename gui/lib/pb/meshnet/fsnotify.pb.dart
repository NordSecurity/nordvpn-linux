// This is a generated file - do not edit.
//
// Generated from fsnotify.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'empty.pb.dart' as $0;
import 'peer.pbenum.dart' as $1;
import 'service_response.pbenum.dart' as $2;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// NewTransferNotification defines a notification structure about a new transfer
class NewTransferNotification extends $pb.GeneratedMessage {
  factory NewTransferNotification({
    $core.String? identifier,
    $core.String? os,
    $core.String? fileName,
    $core.int? fileCount,
    $core.String? transferId,
  }) {
    final result = create();
    if (identifier != null) result.identifier = identifier;
    if (os != null) result.os = os;
    if (fileName != null) result.fileName = fileName;
    if (fileCount != null) result.fileCount = fileCount;
    if (transferId != null) result.transferId = transferId;
    return result;
  }

  NewTransferNotification._();

  factory NewTransferNotification.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory NewTransferNotification.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'NewTransferNotification',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'identifier')
    ..aOS(2, _omitFieldNames ? '' : 'os')
    ..aOS(3, _omitFieldNames ? '' : 'fileName')
    ..a<$core.int>(4, _omitFieldNames ? '' : 'fileCount', $pb.PbFieldType.O3)
    ..aOS(5, _omitFieldNames ? '' : 'transferId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NewTransferNotification clone() =>
      NewTransferNotification()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NewTransferNotification copyWith(
          void Function(NewTransferNotification) updates) =>
      super.copyWith((message) => updates(message as NewTransferNotification))
          as NewTransferNotification;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static NewTransferNotification create() => NewTransferNotification._();
  @$core.override
  NewTransferNotification createEmptyInstance() => create();
  static $pb.PbList<NewTransferNotification> createRepeated() =>
      $pb.PbList<NewTransferNotification>();
  @$core.pragma('dart2js:noInline')
  static NewTransferNotification getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<NewTransferNotification>(create);
  static NewTransferNotification? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get identifier => $_getSZ(0);
  @$pb.TagNumber(1)
  set identifier($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasIdentifier() => $_has(0);
  @$pb.TagNumber(1)
  void clearIdentifier() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get os => $_getSZ(1);
  @$pb.TagNumber(2)
  set os($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasOs() => $_has(1);
  @$pb.TagNumber(2)
  void clearOs() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get fileName => $_getSZ(2);
  @$pb.TagNumber(3)
  set fileName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasFileName() => $_has(2);
  @$pb.TagNumber(3)
  void clearFileName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get fileCount => $_getIZ(3);
  @$pb.TagNumber(4)
  set fileCount($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasFileCount() => $_has(3);
  @$pb.TagNumber(4)
  void clearFileCount() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get transferId => $_getSZ(4);
  @$pb.TagNumber(5)
  set transferId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasTransferId() => $_has(4);
  @$pb.TagNumber(5)
  void clearTransferId() => $_clearField(5);
}

enum NotifyNewTransferResponse_Response {
  empty,
  updatePeerErrorCode,
  serviceErrorCode,
  meshnetErrorCode,
  notSet
}

/// NotifyNewTransferResponse defines a response of new transfer notification
class NotifyNewTransferResponse extends $pb.GeneratedMessage {
  factory NotifyNewTransferResponse({
    $0.Empty? empty,
    $1.UpdatePeerErrorCode? updatePeerErrorCode,
    $2.ServiceErrorCode? serviceErrorCode,
    $2.MeshnetErrorCode? meshnetErrorCode,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (updatePeerErrorCode != null)
      result.updatePeerErrorCode = updatePeerErrorCode;
    if (serviceErrorCode != null) result.serviceErrorCode = serviceErrorCode;
    if (meshnetErrorCode != null) result.meshnetErrorCode = meshnetErrorCode;
    return result;
  }

  NotifyNewTransferResponse._();

  factory NotifyNewTransferResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory NotifyNewTransferResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, NotifyNewTransferResponse_Response>
      _NotifyNewTransferResponse_ResponseByTag = {
    1: NotifyNewTransferResponse_Response.empty,
    2: NotifyNewTransferResponse_Response.updatePeerErrorCode,
    3: NotifyNewTransferResponse_Response.serviceErrorCode,
    4: NotifyNewTransferResponse_Response.meshnetErrorCode,
    0: NotifyNewTransferResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'NotifyNewTransferResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2, 3, 4])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<$1.UpdatePeerErrorCode>(
        2, _omitFieldNames ? '' : 'updatePeerErrorCode', $pb.PbFieldType.OE,
        defaultOrMaker: $1.UpdatePeerErrorCode.PEER_NOT_FOUND,
        valueOf: $1.UpdatePeerErrorCode.valueOf,
        enumValues: $1.UpdatePeerErrorCode.values)
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
  NotifyNewTransferResponse clone() =>
      NotifyNewTransferResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NotifyNewTransferResponse copyWith(
          void Function(NotifyNewTransferResponse) updates) =>
      super.copyWith((message) => updates(message as NotifyNewTransferResponse))
          as NotifyNewTransferResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static NotifyNewTransferResponse create() => NotifyNewTransferResponse._();
  @$core.override
  NotifyNewTransferResponse createEmptyInstance() => create();
  static $pb.PbList<NotifyNewTransferResponse> createRepeated() =>
      $pb.PbList<NotifyNewTransferResponse>();
  @$core.pragma('dart2js:noInline')
  static NotifyNewTransferResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<NotifyNewTransferResponse>(create);
  static NotifyNewTransferResponse? _defaultInstance;

  NotifyNewTransferResponse_Response whichResponse() =>
      _NotifyNewTransferResponse_ResponseByTag[$_whichOneof(0)]!;
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

  @$pb.TagNumber(2)
  $1.UpdatePeerErrorCode get updatePeerErrorCode => $_getN(1);
  @$pb.TagNumber(2)
  set updatePeerErrorCode($1.UpdatePeerErrorCode value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasUpdatePeerErrorCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearUpdatePeerErrorCode() => $_clearField(2);

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
