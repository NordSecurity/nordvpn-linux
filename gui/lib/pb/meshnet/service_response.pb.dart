// This is a generated file - do not edit.
//
// Generated from service_response.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'empty.pb.dart' as $0;
import 'service_response.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'service_response.pbenum.dart';

enum MeshnetResponse_Response { empty, serviceError, meshnetError, notSet }

/// MeshnetErrorCode is one of the:
/// - Empty response
/// - Service error
/// - Meshnet error
class MeshnetResponse extends $pb.GeneratedMessage {
  factory MeshnetResponse({
    $0.Empty? empty,
    ServiceErrorCode? serviceError,
    MeshnetErrorCode? meshnetError,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (serviceError != null) result.serviceError = serviceError;
    if (meshnetError != null) result.meshnetError = meshnetError;
    return result;
  }

  MeshnetResponse._();

  factory MeshnetResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory MeshnetResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, MeshnetResponse_Response>
      _MeshnetResponse_ResponseByTag = {
    1: MeshnetResponse_Response.empty,
    2: MeshnetResponse_Response.serviceError,
    3: MeshnetResponse_Response.meshnetError,
    0: MeshnetResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MeshnetResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2, 3])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<ServiceErrorCode>(
        2, _omitFieldNames ? '' : 'serviceError', $pb.PbFieldType.OE,
        defaultOrMaker: ServiceErrorCode.NOT_LOGGED_IN,
        valueOf: ServiceErrorCode.valueOf,
        enumValues: ServiceErrorCode.values)
    ..e<MeshnetErrorCode>(
        3, _omitFieldNames ? '' : 'meshnetError', $pb.PbFieldType.OE,
        defaultOrMaker: MeshnetErrorCode.NOT_REGISTERED,
        valueOf: MeshnetErrorCode.valueOf,
        enumValues: MeshnetErrorCode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MeshnetResponse clone() => MeshnetResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MeshnetResponse copyWith(void Function(MeshnetResponse) updates) =>
      super.copyWith((message) => updates(message as MeshnetResponse))
          as MeshnetResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static MeshnetResponse create() => MeshnetResponse._();
  @$core.override
  MeshnetResponse createEmptyInstance() => create();
  static $pb.PbList<MeshnetResponse> createRepeated() =>
      $pb.PbList<MeshnetResponse>();
  @$core.pragma('dart2js:noInline')
  static MeshnetResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MeshnetResponse>(create);
  static MeshnetResponse? _defaultInstance;

  MeshnetResponse_Response whichResponse() =>
      _MeshnetResponse_ResponseByTag[$_whichOneof(0)]!;
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
  ServiceErrorCode get serviceError => $_getN(1);
  @$pb.TagNumber(2)
  set serviceError(ServiceErrorCode value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasServiceError() => $_has(1);
  @$pb.TagNumber(2)
  void clearServiceError() => $_clearField(2);

  @$pb.TagNumber(3)
  MeshnetErrorCode get meshnetError => $_getN(2);
  @$pb.TagNumber(3)
  set meshnetError(MeshnetErrorCode value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasMeshnetError() => $_has(2);
  @$pb.TagNumber(3)
  void clearMeshnetError() => $_clearField(3);
}

enum ServiceResponse_Response { empty, errorCode, notSet }

/// ServiceResponse is either an empty response or a service error
class ServiceResponse extends $pb.GeneratedMessage {
  factory ServiceResponse({
    $0.Empty? empty,
    ServiceErrorCode? errorCode,
  }) {
    final result = create();
    if (empty != null) result.empty = empty;
    if (errorCode != null) result.errorCode = errorCode;
    return result;
  }

  ServiceResponse._();

  factory ServiceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ServiceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, ServiceResponse_Response>
      _ServiceResponse_ResponseByTag = {
    1: ServiceResponse_Response.empty,
    2: ServiceResponse_Response.errorCode,
    0: ServiceResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ServiceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2])
    ..aOM<$0.Empty>(1, _omitFieldNames ? '' : 'empty',
        subBuilder: $0.Empty.create)
    ..e<ServiceErrorCode>(
        2, _omitFieldNames ? '' : 'errorCode', $pb.PbFieldType.OE,
        defaultOrMaker: ServiceErrorCode.NOT_LOGGED_IN,
        valueOf: ServiceErrorCode.valueOf,
        enumValues: ServiceErrorCode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServiceResponse clone() => ServiceResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServiceResponse copyWith(void Function(ServiceResponse) updates) =>
      super.copyWith((message) => updates(message as ServiceResponse))
          as ServiceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ServiceResponse create() => ServiceResponse._();
  @$core.override
  ServiceResponse createEmptyInstance() => create();
  static $pb.PbList<ServiceResponse> createRepeated() =>
      $pb.PbList<ServiceResponse>();
  @$core.pragma('dart2js:noInline')
  static ServiceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ServiceResponse>(create);
  static ServiceResponse? _defaultInstance;

  ServiceResponse_Response whichResponse() =>
      _ServiceResponse_ResponseByTag[$_whichOneof(0)]!;
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
  ServiceErrorCode get errorCode => $_getN(1);
  @$pb.TagNumber(2)
  set errorCode(ServiceErrorCode value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasErrorCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearErrorCode() => $_clearField(2);
}

enum ServiceBoolResponse_Response { value, errorCode, notSet }

/// ServiceBoolResponse is either a bool response or a service error
class ServiceBoolResponse extends $pb.GeneratedMessage {
  factory ServiceBoolResponse({
    $core.bool? value,
    ServiceErrorCode? errorCode,
  }) {
    final result = create();
    if (value != null) result.value = value;
    if (errorCode != null) result.errorCode = errorCode;
    return result;
  }

  ServiceBoolResponse._();

  factory ServiceBoolResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ServiceBoolResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, ServiceBoolResponse_Response>
      _ServiceBoolResponse_ResponseByTag = {
    1: ServiceBoolResponse_Response.value,
    2: ServiceBoolResponse_Response.errorCode,
    0: ServiceBoolResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ServiceBoolResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2])
    ..aOB(1, _omitFieldNames ? '' : 'value')
    ..e<ServiceErrorCode>(
        2, _omitFieldNames ? '' : 'errorCode', $pb.PbFieldType.OE,
        defaultOrMaker: ServiceErrorCode.NOT_LOGGED_IN,
        valueOf: ServiceErrorCode.valueOf,
        enumValues: ServiceErrorCode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServiceBoolResponse clone() => ServiceBoolResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServiceBoolResponse copyWith(void Function(ServiceBoolResponse) updates) =>
      super.copyWith((message) => updates(message as ServiceBoolResponse))
          as ServiceBoolResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ServiceBoolResponse create() => ServiceBoolResponse._();
  @$core.override
  ServiceBoolResponse createEmptyInstance() => create();
  static $pb.PbList<ServiceBoolResponse> createRepeated() =>
      $pb.PbList<ServiceBoolResponse>();
  @$core.pragma('dart2js:noInline')
  static ServiceBoolResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ServiceBoolResponse>(create);
  static ServiceBoolResponse? _defaultInstance;

  ServiceBoolResponse_Response whichResponse() =>
      _ServiceBoolResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $core.bool get value => $_getBF(0);
  @$pb.TagNumber(1)
  set value($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasValue() => $_has(0);
  @$pb.TagNumber(1)
  void clearValue() => $_clearField(1);

  @$pb.TagNumber(2)
  ServiceErrorCode get errorCode => $_getN(1);
  @$pb.TagNumber(2)
  set errorCode(ServiceErrorCode value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasErrorCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearErrorCode() => $_clearField(2);
}

class EnabledStatus extends $pb.GeneratedMessage {
  factory EnabledStatus({
    $core.bool? value,
    $core.int? uid,
  }) {
    final result = create();
    if (value != null) result.value = value;
    if (uid != null) result.uid = uid;
    return result;
  }

  EnabledStatus._();

  factory EnabledStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory EnabledStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EnabledStatus',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'value')
    ..a<$core.int>(2, _omitFieldNames ? '' : 'uid', $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EnabledStatus clone() => EnabledStatus()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EnabledStatus copyWith(void Function(EnabledStatus) updates) =>
      super.copyWith((message) => updates(message as EnabledStatus))
          as EnabledStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static EnabledStatus create() => EnabledStatus._();
  @$core.override
  EnabledStatus createEmptyInstance() => create();
  static $pb.PbList<EnabledStatus> createRepeated() =>
      $pb.PbList<EnabledStatus>();
  @$core.pragma('dart2js:noInline')
  static EnabledStatus getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<EnabledStatus>(create);
  static EnabledStatus? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get value => $_getBF(0);
  @$pb.TagNumber(1)
  set value($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasValue() => $_has(0);
  @$pb.TagNumber(1)
  void clearValue() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get uid => $_getIZ(1);
  @$pb.TagNumber(2)
  set uid($core.int value) => $_setUnsignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUid() => $_has(1);
  @$pb.TagNumber(2)
  void clearUid() => $_clearField(2);
}

enum IsEnabledResponse_Response { status, errorCode, notSet }

class IsEnabledResponse extends $pb.GeneratedMessage {
  factory IsEnabledResponse({
    EnabledStatus? status,
    ServiceErrorCode? errorCode,
  }) {
    final result = create();
    if (status != null) result.status = status;
    if (errorCode != null) result.errorCode = errorCode;
    return result;
  }

  IsEnabledResponse._();

  factory IsEnabledResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory IsEnabledResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, IsEnabledResponse_Response>
      _IsEnabledResponse_ResponseByTag = {
    1: IsEnabledResponse_Response.status,
    2: IsEnabledResponse_Response.errorCode,
    0: IsEnabledResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'IsEnabledResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'meshpb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2])
    ..aOM<EnabledStatus>(1, _omitFieldNames ? '' : 'status',
        subBuilder: EnabledStatus.create)
    ..e<ServiceErrorCode>(
        2, _omitFieldNames ? '' : 'errorCode', $pb.PbFieldType.OE,
        defaultOrMaker: ServiceErrorCode.NOT_LOGGED_IN,
        valueOf: ServiceErrorCode.valueOf,
        enumValues: ServiceErrorCode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IsEnabledResponse clone() => IsEnabledResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IsEnabledResponse copyWith(void Function(IsEnabledResponse) updates) =>
      super.copyWith((message) => updates(message as IsEnabledResponse))
          as IsEnabledResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static IsEnabledResponse create() => IsEnabledResponse._();
  @$core.override
  IsEnabledResponse createEmptyInstance() => create();
  static $pb.PbList<IsEnabledResponse> createRepeated() =>
      $pb.PbList<IsEnabledResponse>();
  @$core.pragma('dart2js:noInline')
  static IsEnabledResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<IsEnabledResponse>(create);
  static IsEnabledResponse? _defaultInstance;

  IsEnabledResponse_Response whichResponse() =>
      _IsEnabledResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  EnabledStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(EnabledStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
  @$pb.TagNumber(1)
  EnabledStatus ensureStatus() => $_ensure(0);

  @$pb.TagNumber(2)
  ServiceErrorCode get errorCode => $_getN(1);
  @$pb.TagNumber(2)
  set errorCode(ServiceErrorCode value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasErrorCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearErrorCode() => $_clearField(2);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
