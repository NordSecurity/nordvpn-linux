// This is a generated file - do not edit.
//
// Generated from account.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'common.pbenum.dart' as $0;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

class DedidcatedIPService extends $pb.GeneratedMessage {
  factory DedidcatedIPService({
    $core.Iterable<$fixnum.Int64>? serverIds,
    $core.String? dedicatedIpExpiresAt,
  }) {
    final result = create();
    if (serverIds != null) result.serverIds.addAll(serverIds);
    if (dedicatedIpExpiresAt != null)
      result.dedicatedIpExpiresAt = dedicatedIpExpiresAt;
    return result;
  }

  DedidcatedIPService._();

  factory DedidcatedIPService.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DedidcatedIPService.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DedidcatedIPService',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..p<$fixnum.Int64>(
        1, _omitFieldNames ? '' : 'serverIds', $pb.PbFieldType.K6)
    ..aOS(2, _omitFieldNames ? '' : 'dedicatedIpExpiresAt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DedidcatedIPService clone() => DedidcatedIPService()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DedidcatedIPService copyWith(void Function(DedidcatedIPService) updates) =>
      super.copyWith((message) => updates(message as DedidcatedIPService))
          as DedidcatedIPService;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DedidcatedIPService create() => DedidcatedIPService._();
  @$core.override
  DedidcatedIPService createEmptyInstance() => create();
  static $pb.PbList<DedidcatedIPService> createRepeated() =>
      $pb.PbList<DedidcatedIPService>();
  @$core.pragma('dart2js:noInline')
  static DedidcatedIPService getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DedidcatedIPService>(create);
  static DedidcatedIPService? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$fixnum.Int64> get serverIds => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get dedicatedIpExpiresAt => $_getSZ(1);
  @$pb.TagNumber(2)
  set dedicatedIpExpiresAt($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDedicatedIpExpiresAt() => $_has(1);
  @$pb.TagNumber(2)
  void clearDedicatedIpExpiresAt() => $_clearField(2);
}

class AccountResponse extends $pb.GeneratedMessage {
  factory AccountResponse({
    $fixnum.Int64? type,
    $core.String? username,
    $core.String? email,
    $core.String? expiresAt,
    $fixnum.Int64? dedicatedIpStatus,
    $core.String? lastDedicatedIpExpiresAt,
    $core.Iterable<DedidcatedIPService>? dedicatedIpServices,
    $0.TriState? mfaStatus,
  }) {
    final result = create();
    if (type != null) result.type = type;
    if (username != null) result.username = username;
    if (email != null) result.email = email;
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (dedicatedIpStatus != null) result.dedicatedIpStatus = dedicatedIpStatus;
    if (lastDedicatedIpExpiresAt != null)
      result.lastDedicatedIpExpiresAt = lastDedicatedIpExpiresAt;
    if (dedicatedIpServices != null)
      result.dedicatedIpServices.addAll(dedicatedIpServices);
    if (mfaStatus != null) result.mfaStatus = mfaStatus;
    return result;
  }

  AccountResponse._();

  factory AccountResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AccountResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AccountResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aInt64(1, _omitFieldNames ? '' : 'type')
    ..aOS(2, _omitFieldNames ? '' : 'username')
    ..aOS(3, _omitFieldNames ? '' : 'email')
    ..aOS(4, _omitFieldNames ? '' : 'expiresAt')
    ..aInt64(5, _omitFieldNames ? '' : 'dedicatedIpStatus')
    ..aOS(6, _omitFieldNames ? '' : 'lastDedicatedIpExpiresAt')
    ..pc<DedidcatedIPService>(
        7, _omitFieldNames ? '' : 'dedicatedIpServices', $pb.PbFieldType.PM,
        subBuilder: DedidcatedIPService.create)
    ..e<$0.TriState>(8, _omitFieldNames ? '' : 'mfaStatus', $pb.PbFieldType.OE,
        defaultOrMaker: $0.TriState.UNKNOWN,
        valueOf: $0.TriState.valueOf,
        enumValues: $0.TriState.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccountResponse clone() => AccountResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccountResponse copyWith(void Function(AccountResponse) updates) =>
      super.copyWith((message) => updates(message as AccountResponse))
          as AccountResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AccountResponse create() => AccountResponse._();
  @$core.override
  AccountResponse createEmptyInstance() => create();
  static $pb.PbList<AccountResponse> createRepeated() =>
      $pb.PbList<AccountResponse>();
  @$core.pragma('dart2js:noInline')
  static AccountResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AccountResponse>(create);
  static AccountResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get type => $_getI64(0);
  @$pb.TagNumber(1)
  set type($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get username => $_getSZ(1);
  @$pb.TagNumber(2)
  set username($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUsername() => $_has(1);
  @$pb.TagNumber(2)
  void clearUsername() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get email => $_getSZ(2);
  @$pb.TagNumber(3)
  set email($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasEmail() => $_has(2);
  @$pb.TagNumber(3)
  void clearEmail() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get expiresAt => $_getSZ(3);
  @$pb.TagNumber(4)
  set expiresAt($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasExpiresAt() => $_has(3);
  @$pb.TagNumber(4)
  void clearExpiresAt() => $_clearField(4);

  @$pb.TagNumber(5)
  $fixnum.Int64 get dedicatedIpStatus => $_getI64(4);
  @$pb.TagNumber(5)
  set dedicatedIpStatus($fixnum.Int64 value) => $_setInt64(4, value);
  @$pb.TagNumber(5)
  $core.bool hasDedicatedIpStatus() => $_has(4);
  @$pb.TagNumber(5)
  void clearDedicatedIpStatus() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get lastDedicatedIpExpiresAt => $_getSZ(5);
  @$pb.TagNumber(6)
  set lastDedicatedIpExpiresAt($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasLastDedicatedIpExpiresAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearLastDedicatedIpExpiresAt() => $_clearField(6);

  @$pb.TagNumber(7)
  $pb.PbList<DedidcatedIPService> get dedicatedIpServices => $_getList(6);

  @$pb.TagNumber(8)
  $0.TriState get mfaStatus => $_getN(7);
  @$pb.TagNumber(8)
  set mfaStatus($0.TriState value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasMfaStatus() => $_has(7);
  @$pb.TagNumber(8)
  void clearMfaStatus() => $_clearField(8);
}

class AccountRequest extends $pb.GeneratedMessage {
  factory AccountRequest({
    $core.bool? full,
  }) {
    final result = create();
    if (full != null) result.full = full;
    return result;
  }

  AccountRequest._();

  factory AccountRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AccountRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AccountRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'full')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccountRequest clone() => AccountRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccountRequest copyWith(void Function(AccountRequest) updates) =>
      super.copyWith((message) => updates(message as AccountRequest))
          as AccountRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AccountRequest create() => AccountRequest._();
  @$core.override
  AccountRequest createEmptyInstance() => create();
  static $pb.PbList<AccountRequest> createRepeated() =>
      $pb.PbList<AccountRequest>();
  @$core.pragma('dart2js:noInline')
  static AccountRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AccountRequest>(create);
  static AccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get full => $_getBF(0);
  @$pb.TagNumber(1)
  set full($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasFull() => $_has(0);
  @$pb.TagNumber(1)
  void clearFull() => $_clearField(1);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
