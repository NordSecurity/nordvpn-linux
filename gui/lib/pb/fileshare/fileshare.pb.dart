// This is a generated file - do not edit.
//
// Generated from fileshare.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'fileshare.pbenum.dart';
import 'google/protobuf/timestamp.pb.dart' as $1;
import 'transfer.pb.dart' as $0;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'fileshare.pbenum.dart';

/// Used when there is no error or there is no data to be sent
class Empty extends $pb.GeneratedMessage {
  factory Empty() => create();

  Empty._();

  factory Empty.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Empty.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Empty',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'filesharepb'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Empty clone() => Empty()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Empty copyWith(void Function(Empty) updates) =>
      super.copyWith((message) => updates(message as Empty)) as Empty;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Empty create() => Empty._();
  @$core.override
  Empty createEmptyInstance() => create();
  static $pb.PbList<Empty> createRepeated() => $pb.PbList<Empty>();
  @$core.pragma('dart2js:noInline')
  static Empty getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Empty>(create);
  static Empty? _defaultInstance;
}

enum Error_Response { empty, serviceError, fileshareError, notSet }

/// Generic error to be used through all responses. If empty then no error occurred.
/// If there's no data to be returned then this can be used as a response type,
/// otherwise it should be included as a field in the response.
/// Response handlers should always firstly check whether error is Empty (like Go err != nil check)
///
/// Previously (in meshnet) we have used oneof to either return data or an error. But the problem
/// with oneof is that when it is used the same messages are returned as different types
/// (SendResponse_FileshareResponse and ReceiveResponse_FileshareResponse for example). Because of that
/// we couldn't DRY their handling and that resulted in lots of almost duplicate code.
class Error extends $pb.GeneratedMessage {
  factory Error({
    Empty? empty,
    ServiceErrorCode? serviceError,
    FileshareErrorCode? fileshareError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (serviceError != null) result.serviceError = serviceError;
    if (fileshareError != null) result.fileshareError = fileshareError;
    return result;
  }

  Error._();

  factory Error.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Error.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, Error_Response> _Error_ResponseByTag = {
    1: Error_Response.empty,
    2: Error_Response.serviceError,
    3: Error_Response.fileshareError,
    0: Error_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Error',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'filesharepb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2, 3])
    ..aOM<Empty>(1, _omitFieldNames ? '' : 'empty', subBuilder: Empty.create)
    ..e<ServiceErrorCode>(
        2, _omitFieldNames ? '' : 'serviceError', $pb.PbFieldType.OE,
        defaultOrMaker: ServiceErrorCode.MESH_NOT_ENABLED,
        valueOf: ServiceErrorCode.valueOf,
        enumValues: ServiceErrorCode.values)
    ..e<FileshareErrorCode>(
        3, _omitFieldNames ? '' : 'fileshareError', $pb.PbFieldType.OE,
        defaultOrMaker: FileshareErrorCode.LIB_FAILURE,
        valueOf: FileshareErrorCode.valueOf,
        enumValues: FileshareErrorCode.values)
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

  Error_Response whichResponse() => _Error_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  Empty get empty => $_getN(0);
  @$pb.TagNumber(1)
  set empty(Empty value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasEmpty() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmpty() => $_clearField(1);
  @$pb.TagNumber(1)
  Empty ensureEmpty() => $_ensure(0);

  @$pb.TagNumber(2)
  ServiceErrorCode get serviceError => $_getN(1);
  @$pb.TagNumber(2)
  set serviceError(ServiceErrorCode value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasServiceError() => $_has(1);
  @$pb.TagNumber(2)
  void clearServiceError() => $_clearField(2);

  @$pb.TagNumber(3)
  FileshareErrorCode get fileshareError => $_getN(2);
  @$pb.TagNumber(3)
  set fileshareError(FileshareErrorCode value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasFileshareError() => $_has(2);
  @$pb.TagNumber(3)
  void clearFileshareError() => $_clearField(3);
}

class SendRequest extends $pb.GeneratedMessage {
  factory SendRequest({
    $core.String? peer,
    $core.Iterable<$core.String>? paths,
    $core.bool? silent,
  }) {
    final result = create();
    if (peer != null) result.peer = peer;
    if (paths != null) result.paths.addAll(paths);
    if (silent != null) result.silent = silent;
    return result;
  }

  SendRequest._();

  factory SendRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SendRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SendRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'filesharepb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'peer')
    ..pPS(2, _omitFieldNames ? '' : 'paths')
    ..aOB(3, _omitFieldNames ? '' : 'silent')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendRequest clone() => SendRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendRequest copyWith(void Function(SendRequest) updates) =>
      super.copyWith((message) => updates(message as SendRequest))
          as SendRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SendRequest create() => SendRequest._();
  @$core.override
  SendRequest createEmptyInstance() => create();
  static $pb.PbList<SendRequest> createRepeated() => $pb.PbList<SendRequest>();
  @$core.pragma('dart2js:noInline')
  static SendRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SendRequest>(create);
  static SendRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get peer => $_getSZ(0);
  @$pb.TagNumber(1)
  set peer($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPeer() => $_has(0);
  @$pb.TagNumber(1)
  void clearPeer() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get paths => $_getList(1);

  @$pb.TagNumber(3)
  $core.bool get silent => $_getBF(2);
  @$pb.TagNumber(3)
  set silent($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSilent() => $_has(2);
  @$pb.TagNumber(3)
  void clearSilent() => $_clearField(3);
}

class AcceptRequest extends $pb.GeneratedMessage {
  factory AcceptRequest({
    $core.String? transferId,
    $core.String? dstPath,
    $core.bool? silent,
    $core.Iterable<$core.String>? files,
  }) {
    final result = create();
    if (transferId != null) result.transferId = transferId;
    if (dstPath != null) result.dstPath = dstPath;
    if (silent != null) result.silent = silent;
    if (files != null) result.files.addAll(files);
    return result;
  }

  AcceptRequest._();

  factory AcceptRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AcceptRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AcceptRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'filesharepb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'transferId')
    ..aOS(2, _omitFieldNames ? '' : 'dstPath')
    ..aOB(3, _omitFieldNames ? '' : 'silent')
    ..pPS(4, _omitFieldNames ? '' : 'files')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AcceptRequest clone() => AcceptRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AcceptRequest copyWith(void Function(AcceptRequest) updates) =>
      super.copyWith((message) => updates(message as AcceptRequest))
          as AcceptRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AcceptRequest create() => AcceptRequest._();
  @$core.override
  AcceptRequest createEmptyInstance() => create();
  static $pb.PbList<AcceptRequest> createRepeated() =>
      $pb.PbList<AcceptRequest>();
  @$core.pragma('dart2js:noInline')
  static AcceptRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AcceptRequest>(create);
  static AcceptRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get transferId => $_getSZ(0);
  @$pb.TagNumber(1)
  set transferId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTransferId() => $_has(0);
  @$pb.TagNumber(1)
  void clearTransferId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get dstPath => $_getSZ(1);
  @$pb.TagNumber(2)
  set dstPath($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDstPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearDstPath() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get silent => $_getBF(2);
  @$pb.TagNumber(3)
  set silent($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSilent() => $_has(2);
  @$pb.TagNumber(3)
  void clearSilent() => $_clearField(3);

  @$pb.TagNumber(4)
  $pb.PbList<$core.String> get files => $_getList(3);
}

class StatusResponse extends $pb.GeneratedMessage {
  factory StatusResponse({
    Error? error,
    $core.String? transferId,
    $core.int? progress,
    $0.Status? status,
  }) {
    final result = create();
    if (error != null) result.error = error;
    if (transferId != null) result.transferId = transferId;
    if (progress != null) result.progress = progress;
    if (status != null) result.status = status;
    return result;
  }

  StatusResponse._();

  factory StatusResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory StatusResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StatusResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'filesharepb'),
      createEmptyInstance: create)
    ..aOM<Error>(1, _omitFieldNames ? '' : 'error', subBuilder: Error.create)
    ..aOS(2, _omitFieldNames ? '' : 'transferId')
    ..a<$core.int>(3, _omitFieldNames ? '' : 'progress', $pb.PbFieldType.OU3)
    ..e<$0.Status>(4, _omitFieldNames ? '' : 'status', $pb.PbFieldType.OE,
        defaultOrMaker: $0.Status.SUCCESS,
        valueOf: $0.Status.valueOf,
        enumValues: $0.Status.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StatusResponse clone() => StatusResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StatusResponse copyWith(void Function(StatusResponse) updates) =>
      super.copyWith((message) => updates(message as StatusResponse))
          as StatusResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static StatusResponse create() => StatusResponse._();
  @$core.override
  StatusResponse createEmptyInstance() => create();
  static $pb.PbList<StatusResponse> createRepeated() =>
      $pb.PbList<StatusResponse>();
  @$core.pragma('dart2js:noInline')
  static StatusResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StatusResponse>(create);
  static StatusResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Error get error => $_getN(0);
  @$pb.TagNumber(1)
  set error(Error value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasError() => $_has(0);
  @$pb.TagNumber(1)
  void clearError() => $_clearField(1);
  @$pb.TagNumber(1)
  Error ensureError() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.String get transferId => $_getSZ(1);
  @$pb.TagNumber(2)
  set transferId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTransferId() => $_has(1);
  @$pb.TagNumber(2)
  void clearTransferId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get progress => $_getIZ(2);
  @$pb.TagNumber(3)
  set progress($core.int value) => $_setUnsignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasProgress() => $_has(2);
  @$pb.TagNumber(3)
  void clearProgress() => $_clearField(3);

  @$pb.TagNumber(4)
  $0.Status get status => $_getN(3);
  @$pb.TagNumber(4)
  set status($0.Status value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasStatus() => $_has(3);
  @$pb.TagNumber(4)
  void clearStatus() => $_clearField(4);
}

class CancelRequest extends $pb.GeneratedMessage {
  factory CancelRequest({
    $core.String? transferId,
  }) {
    final result = create();
    if (transferId != null) result.transferId = transferId;
    return result;
  }

  CancelRequest._();

  factory CancelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CancelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'filesharepb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'transferId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelRequest clone() => CancelRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelRequest copyWith(void Function(CancelRequest) updates) =>
      super.copyWith((message) => updates(message as CancelRequest))
          as CancelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CancelRequest create() => CancelRequest._();
  @$core.override
  CancelRequest createEmptyInstance() => create();
  static $pb.PbList<CancelRequest> createRepeated() =>
      $pb.PbList<CancelRequest>();
  @$core.pragma('dart2js:noInline')
  static CancelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelRequest>(create);
  static CancelRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get transferId => $_getSZ(0);
  @$pb.TagNumber(1)
  set transferId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTransferId() => $_has(0);
  @$pb.TagNumber(1)
  void clearTransferId() => $_clearField(1);
}

class ListResponse extends $pb.GeneratedMessage {
  factory ListResponse({
    Error? error,
    $core.Iterable<$0.Transfer>? transfers,
  }) {
    final result = create();
    if (error != null) result.error = error;
    if (transfers != null) result.transfers.addAll(transfers);
    return result;
  }

  ListResponse._();

  factory ListResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'filesharepb'),
      createEmptyInstance: create)
    ..aOM<Error>(1, _omitFieldNames ? '' : 'error', subBuilder: Error.create)
    ..pc<$0.Transfer>(2, _omitFieldNames ? '' : 'transfers', $pb.PbFieldType.PM,
        subBuilder: $0.Transfer.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListResponse clone() => ListResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListResponse copyWith(void Function(ListResponse) updates) =>
      super.copyWith((message) => updates(message as ListResponse))
          as ListResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListResponse create() => ListResponse._();
  @$core.override
  ListResponse createEmptyInstance() => create();
  static $pb.PbList<ListResponse> createRepeated() =>
      $pb.PbList<ListResponse>();
  @$core.pragma('dart2js:noInline')
  static ListResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListResponse>(create);
  static ListResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Error get error => $_getN(0);
  @$pb.TagNumber(1)
  set error(Error value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasError() => $_has(0);
  @$pb.TagNumber(1)
  void clearError() => $_clearField(1);
  @$pb.TagNumber(1)
  Error ensureError() => $_ensure(0);

  /// Transfers are sorted by creation date from oldest to newest
  @$pb.TagNumber(2)
  $pb.PbList<$0.Transfer> get transfers => $_getList(1);
}

class CancelFileRequest extends $pb.GeneratedMessage {
  factory CancelFileRequest({
    $core.String? transferId,
    $core.String? filePath,
  }) {
    final result = create();
    if (transferId != null) result.transferId = transferId;
    if (filePath != null) result.filePath = filePath;
    return result;
  }

  CancelFileRequest._();

  factory CancelFileRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CancelFileRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelFileRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'filesharepb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'transferId')
    ..aOS(2, _omitFieldNames ? '' : 'filePath')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelFileRequest clone() => CancelFileRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelFileRequest copyWith(void Function(CancelFileRequest) updates) =>
      super.copyWith((message) => updates(message as CancelFileRequest))
          as CancelFileRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CancelFileRequest create() => CancelFileRequest._();
  @$core.override
  CancelFileRequest createEmptyInstance() => create();
  static $pb.PbList<CancelFileRequest> createRepeated() =>
      $pb.PbList<CancelFileRequest>();
  @$core.pragma('dart2js:noInline')
  static CancelFileRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelFileRequest>(create);
  static CancelFileRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get transferId => $_getSZ(0);
  @$pb.TagNumber(1)
  set transferId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTransferId() => $_has(0);
  @$pb.TagNumber(1)
  void clearTransferId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get filePath => $_getSZ(1);
  @$pb.TagNumber(2)
  set filePath($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasFilePath() => $_has(1);
  @$pb.TagNumber(2)
  void clearFilePath() => $_clearField(2);
}

class SetNotificationsRequest extends $pb.GeneratedMessage {
  factory SetNotificationsRequest({
    $core.bool? enable,
  }) {
    final result = create();
    if (enable != null) result.enable = enable;
    return result;
  }

  SetNotificationsRequest._();

  factory SetNotificationsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetNotificationsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetNotificationsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'filesharepb'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'enable')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetNotificationsRequest clone() =>
      SetNotificationsRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetNotificationsRequest copyWith(
          void Function(SetNotificationsRequest) updates) =>
      super.copyWith((message) => updates(message as SetNotificationsRequest))
          as SetNotificationsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetNotificationsRequest create() => SetNotificationsRequest._();
  @$core.override
  SetNotificationsRequest createEmptyInstance() => create();
  static $pb.PbList<SetNotificationsRequest> createRepeated() =>
      $pb.PbList<SetNotificationsRequest>();
  @$core.pragma('dart2js:noInline')
  static SetNotificationsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetNotificationsRequest>(create);
  static SetNotificationsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get enable => $_getBF(0);
  @$pb.TagNumber(1)
  set enable($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEnable() => $_has(0);
  @$pb.TagNumber(1)
  void clearEnable() => $_clearField(1);
}

class SetNotificationsResponse extends $pb.GeneratedMessage {
  factory SetNotificationsResponse({
    SetNotificationsStatus? status,
  }) {
    final result = create();
    if (status != null) result.status = status;
    return result;
  }

  SetNotificationsResponse._();

  factory SetNotificationsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetNotificationsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetNotificationsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'filesharepb'),
      createEmptyInstance: create)
    ..e<SetNotificationsStatus>(
        1, _omitFieldNames ? '' : 'status', $pb.PbFieldType.OE,
        defaultOrMaker: SetNotificationsStatus.SET_SUCCESS,
        valueOf: SetNotificationsStatus.valueOf,
        enumValues: SetNotificationsStatus.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetNotificationsResponse clone() =>
      SetNotificationsResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetNotificationsResponse copyWith(
          void Function(SetNotificationsResponse) updates) =>
      super.copyWith((message) => updates(message as SetNotificationsResponse))
          as SetNotificationsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetNotificationsResponse create() => SetNotificationsResponse._();
  @$core.override
  SetNotificationsResponse createEmptyInstance() => create();
  static $pb.PbList<SetNotificationsResponse> createRepeated() =>
      $pb.PbList<SetNotificationsResponse>();
  @$core.pragma('dart2js:noInline')
  static SetNotificationsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetNotificationsResponse>(create);
  static SetNotificationsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SetNotificationsStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(SetNotificationsStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
}

class PurgeTransfersUntilRequest extends $pb.GeneratedMessage {
  factory PurgeTransfersUntilRequest({
    $1.Timestamp? until,
  }) {
    final result = create();
    if (until != null) result.until = until;
    return result;
  }

  PurgeTransfersUntilRequest._();

  factory PurgeTransfersUntilRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory PurgeTransfersUntilRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PurgeTransfersUntilRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'filesharepb'),
      createEmptyInstance: create)
    ..aOM<$1.Timestamp>(1, _omitFieldNames ? '' : 'until',
        subBuilder: $1.Timestamp.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PurgeTransfersUntilRequest clone() =>
      PurgeTransfersUntilRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PurgeTransfersUntilRequest copyWith(
          void Function(PurgeTransfersUntilRequest) updates) =>
      super.copyWith(
              (message) => updates(message as PurgeTransfersUntilRequest))
          as PurgeTransfersUntilRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PurgeTransfersUntilRequest create() => PurgeTransfersUntilRequest._();
  @$core.override
  PurgeTransfersUntilRequest createEmptyInstance() => create();
  static $pb.PbList<PurgeTransfersUntilRequest> createRepeated() =>
      $pb.PbList<PurgeTransfersUntilRequest>();
  @$core.pragma('dart2js:noInline')
  static PurgeTransfersUntilRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PurgeTransfersUntilRequest>(create);
  static PurgeTransfersUntilRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $1.Timestamp get until => $_getN(0);
  @$pb.TagNumber(1)
  set until($1.Timestamp value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasUntil() => $_has(0);
  @$pb.TagNumber(1)
  void clearUntil() => $_clearField(1);
  @$pb.TagNumber(1)
  $1.Timestamp ensureUntil() => $_ensure(0);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
