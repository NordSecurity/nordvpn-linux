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

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'config/protocol.pbenum.dart' as $0;
import 'config/technology.pbenum.dart' as $1;
import 'set.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'set.pbenum.dart';

class SetAutoconnectRequest extends $pb.GeneratedMessage {
  factory SetAutoconnectRequest({
    $core.bool? enabled,
    $core.String? serverTag,
    $core.String? serverGroup,
  }) {
    final result = create();
    if (enabled != null) result.enabled = enabled;
    if (serverTag != null) result.serverTag = serverTag;
    if (serverGroup != null) result.serverGroup = serverGroup;
    return result;
  }

  SetAutoconnectRequest._();

  factory SetAutoconnectRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetAutoconnectRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetAutoconnectRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'enabled')
    ..aOS(2, _omitFieldNames ? '' : 'serverTag')
    ..aOS(3, _omitFieldNames ? '' : 'serverGroup')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetAutoconnectRequest clone() =>
      SetAutoconnectRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetAutoconnectRequest copyWith(
          void Function(SetAutoconnectRequest) updates) =>
      super.copyWith((message) => updates(message as SetAutoconnectRequest))
          as SetAutoconnectRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetAutoconnectRequest create() => SetAutoconnectRequest._();
  @$core.override
  SetAutoconnectRequest createEmptyInstance() => create();
  static $pb.PbList<SetAutoconnectRequest> createRepeated() =>
      $pb.PbList<SetAutoconnectRequest>();
  @$core.pragma('dart2js:noInline')
  static SetAutoconnectRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetAutoconnectRequest>(create);
  static SetAutoconnectRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get enabled => $_getBF(0);
  @$pb.TagNumber(1)
  set enabled($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEnabled() => $_has(0);
  @$pb.TagNumber(1)
  void clearEnabled() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get serverTag => $_getSZ(1);
  @$pb.TagNumber(2)
  set serverTag($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasServerTag() => $_has(1);
  @$pb.TagNumber(2)
  void clearServerTag() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get serverGroup => $_getSZ(2);
  @$pb.TagNumber(3)
  set serverGroup($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasServerGroup() => $_has(2);
  @$pb.TagNumber(3)
  void clearServerGroup() => $_clearField(3);
}

class SetGenericRequest extends $pb.GeneratedMessage {
  factory SetGenericRequest({
    $core.bool? enabled,
  }) {
    final result = create();
    if (enabled != null) result.enabled = enabled;
    return result;
  }

  SetGenericRequest._();

  factory SetGenericRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetGenericRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetGenericRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'enabled')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetGenericRequest clone() => SetGenericRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetGenericRequest copyWith(void Function(SetGenericRequest) updates) =>
      super.copyWith((message) => updates(message as SetGenericRequest))
          as SetGenericRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetGenericRequest create() => SetGenericRequest._();
  @$core.override
  SetGenericRequest createEmptyInstance() => create();
  static $pb.PbList<SetGenericRequest> createRepeated() =>
      $pb.PbList<SetGenericRequest>();
  @$core.pragma('dart2js:noInline')
  static SetGenericRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetGenericRequest>(create);
  static SetGenericRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get enabled => $_getBF(0);
  @$pb.TagNumber(1)
  set enabled($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEnabled() => $_has(0);
  @$pb.TagNumber(1)
  void clearEnabled() => $_clearField(1);
}

class SetUint32Request extends $pb.GeneratedMessage {
  factory SetUint32Request({
    $core.int? value,
  }) {
    final result = create();
    if (value != null) result.value = value;
    return result;
  }

  SetUint32Request._();

  factory SetUint32Request.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetUint32Request.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetUint32Request',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'value', $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetUint32Request clone() => SetUint32Request()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetUint32Request copyWith(void Function(SetUint32Request) updates) =>
      super.copyWith((message) => updates(message as SetUint32Request))
          as SetUint32Request;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetUint32Request create() => SetUint32Request._();
  @$core.override
  SetUint32Request createEmptyInstance() => create();
  static $pb.PbList<SetUint32Request> createRepeated() =>
      $pb.PbList<SetUint32Request>();
  @$core.pragma('dart2js:noInline')
  static SetUint32Request getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetUint32Request>(create);
  static SetUint32Request? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get value => $_getIZ(0);
  @$pb.TagNumber(1)
  set value($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasValue() => $_has(0);
  @$pb.TagNumber(1)
  void clearValue() => $_clearField(1);
}

class SetThreatProtectionLiteRequest extends $pb.GeneratedMessage {
  factory SetThreatProtectionLiteRequest({
    $core.bool? threatProtectionLite,
  }) {
    final result = create();
    if (threatProtectionLite != null)
      result.threatProtectionLite = threatProtectionLite;
    return result;
  }

  SetThreatProtectionLiteRequest._();

  factory SetThreatProtectionLiteRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetThreatProtectionLiteRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetThreatProtectionLiteRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'threatProtectionLite')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetThreatProtectionLiteRequest clone() =>
      SetThreatProtectionLiteRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetThreatProtectionLiteRequest copyWith(
          void Function(SetThreatProtectionLiteRequest) updates) =>
      super.copyWith(
              (message) => updates(message as SetThreatProtectionLiteRequest))
          as SetThreatProtectionLiteRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetThreatProtectionLiteRequest create() =>
      SetThreatProtectionLiteRequest._();
  @$core.override
  SetThreatProtectionLiteRequest createEmptyInstance() => create();
  static $pb.PbList<SetThreatProtectionLiteRequest> createRepeated() =>
      $pb.PbList<SetThreatProtectionLiteRequest>();
  @$core.pragma('dart2js:noInline')
  static SetThreatProtectionLiteRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetThreatProtectionLiteRequest>(create);
  static SetThreatProtectionLiteRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get threatProtectionLite => $_getBF(0);
  @$pb.TagNumber(1)
  set threatProtectionLite($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasThreatProtectionLite() => $_has(0);
  @$pb.TagNumber(1)
  void clearThreatProtectionLite() => $_clearField(1);
}

enum SetThreatProtectionLiteResponse_Response {
  errorCode,
  setThreatProtectionLiteStatus,
  notSet
}

class SetThreatProtectionLiteResponse extends $pb.GeneratedMessage {
  factory SetThreatProtectionLiteResponse({
    SetErrorCode? errorCode,
    SetThreatProtectionLiteStatus? setThreatProtectionLiteStatus,
  }) {
    final result = create();
    if (errorCode != null) result.errorCode = errorCode;
    if (setThreatProtectionLiteStatus != null)
      result.setThreatProtectionLiteStatus = setThreatProtectionLiteStatus;
    return result;
  }

  SetThreatProtectionLiteResponse._();

  factory SetThreatProtectionLiteResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetThreatProtectionLiteResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, SetThreatProtectionLiteResponse_Response>
      _SetThreatProtectionLiteResponse_ResponseByTag = {
    1: SetThreatProtectionLiteResponse_Response.errorCode,
    2: SetThreatProtectionLiteResponse_Response.setThreatProtectionLiteStatus,
    0: SetThreatProtectionLiteResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetThreatProtectionLiteResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2])
    ..e<SetErrorCode>(1, _omitFieldNames ? '' : 'errorCode', $pb.PbFieldType.OE,
        defaultOrMaker: SetErrorCode.FAILURE,
        valueOf: SetErrorCode.valueOf,
        enumValues: SetErrorCode.values)
    ..e<SetThreatProtectionLiteStatus>(
        2,
        _omitFieldNames ? '' : 'setThreatProtectionLiteStatus',
        $pb.PbFieldType.OE,
        defaultOrMaker: SetThreatProtectionLiteStatus.TPL_CONFIGURED,
        valueOf: SetThreatProtectionLiteStatus.valueOf,
        enumValues: SetThreatProtectionLiteStatus.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetThreatProtectionLiteResponse clone() =>
      SetThreatProtectionLiteResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetThreatProtectionLiteResponse copyWith(
          void Function(SetThreatProtectionLiteResponse) updates) =>
      super.copyWith(
              (message) => updates(message as SetThreatProtectionLiteResponse))
          as SetThreatProtectionLiteResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetThreatProtectionLiteResponse create() =>
      SetThreatProtectionLiteResponse._();
  @$core.override
  SetThreatProtectionLiteResponse createEmptyInstance() => create();
  static $pb.PbList<SetThreatProtectionLiteResponse> createRepeated() =>
      $pb.PbList<SetThreatProtectionLiteResponse>();
  @$core.pragma('dart2js:noInline')
  static SetThreatProtectionLiteResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetThreatProtectionLiteResponse>(
          create);
  static SetThreatProtectionLiteResponse? _defaultInstance;

  SetThreatProtectionLiteResponse_Response whichResponse() =>
      _SetThreatProtectionLiteResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  SetErrorCode get errorCode => $_getN(0);
  @$pb.TagNumber(1)
  set errorCode(SetErrorCode value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasErrorCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearErrorCode() => $_clearField(1);

  @$pb.TagNumber(2)
  SetThreatProtectionLiteStatus get setThreatProtectionLiteStatus => $_getN(1);
  @$pb.TagNumber(2)
  set setThreatProtectionLiteStatus(SetThreatProtectionLiteStatus value) =>
      $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasSetThreatProtectionLiteStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearSetThreatProtectionLiteStatus() => $_clearField(2);
}

class SetDNSRequest extends $pb.GeneratedMessage {
  factory SetDNSRequest({
    $core.Iterable<$core.String>? dns,
    $core.bool? threatProtectionLite,
  }) {
    final result = create();
    if (dns != null) result.dns.addAll(dns);
    if (threatProtectionLite != null)
      result.threatProtectionLite = threatProtectionLite;
    return result;
  }

  SetDNSRequest._();

  factory SetDNSRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetDNSRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetDNSRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..pPS(2, _omitFieldNames ? '' : 'dns')
    ..aOB(3, _omitFieldNames ? '' : 'threatProtectionLite')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetDNSRequest clone() => SetDNSRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetDNSRequest copyWith(void Function(SetDNSRequest) updates) =>
      super.copyWith((message) => updates(message as SetDNSRequest))
          as SetDNSRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetDNSRequest create() => SetDNSRequest._();
  @$core.override
  SetDNSRequest createEmptyInstance() => create();
  static $pb.PbList<SetDNSRequest> createRepeated() =>
      $pb.PbList<SetDNSRequest>();
  @$core.pragma('dart2js:noInline')
  static SetDNSRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetDNSRequest>(create);
  static SetDNSRequest? _defaultInstance;

  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get dns => $_getList(0);

  @$pb.TagNumber(3)
  $core.bool get threatProtectionLite => $_getBF(1);
  @$pb.TagNumber(3)
  set threatProtectionLite($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(3)
  $core.bool hasThreatProtectionLite() => $_has(1);
  @$pb.TagNumber(3)
  void clearThreatProtectionLite() => $_clearField(3);
}

enum SetDNSResponse_Response { errorCode, setDnsStatus, notSet }

class SetDNSResponse extends $pb.GeneratedMessage {
  factory SetDNSResponse({
    SetErrorCode? errorCode,
    SetDNSStatus? setDnsStatus,
  }) {
    final result = create();
    if (errorCode != null) result.errorCode = errorCode;
    if (setDnsStatus != null) result.setDnsStatus = setDnsStatus;
    return result;
  }

  SetDNSResponse._();

  factory SetDNSResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetDNSResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, SetDNSResponse_Response>
      _SetDNSResponse_ResponseByTag = {
    2: SetDNSResponse_Response.errorCode,
    3: SetDNSResponse_Response.setDnsStatus,
    0: SetDNSResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetDNSResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..oo(0, [2, 3])
    ..e<SetErrorCode>(2, _omitFieldNames ? '' : 'errorCode', $pb.PbFieldType.OE,
        defaultOrMaker: SetErrorCode.FAILURE,
        valueOf: SetErrorCode.valueOf,
        enumValues: SetErrorCode.values)
    ..e<SetDNSStatus>(
        3, _omitFieldNames ? '' : 'setDnsStatus', $pb.PbFieldType.OE,
        defaultOrMaker: SetDNSStatus.DNS_CONFIGURED,
        valueOf: SetDNSStatus.valueOf,
        enumValues: SetDNSStatus.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetDNSResponse clone() => SetDNSResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetDNSResponse copyWith(void Function(SetDNSResponse) updates) =>
      super.copyWith((message) => updates(message as SetDNSResponse))
          as SetDNSResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetDNSResponse create() => SetDNSResponse._();
  @$core.override
  SetDNSResponse createEmptyInstance() => create();
  static $pb.PbList<SetDNSResponse> createRepeated() =>
      $pb.PbList<SetDNSResponse>();
  @$core.pragma('dart2js:noInline')
  static SetDNSResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetDNSResponse>(create);
  static SetDNSResponse? _defaultInstance;

  SetDNSResponse_Response whichResponse() =>
      _SetDNSResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(2)
  SetErrorCode get errorCode => $_getN(0);
  @$pb.TagNumber(2)
  set errorCode(SetErrorCode value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasErrorCode() => $_has(0);
  @$pb.TagNumber(2)
  void clearErrorCode() => $_clearField(2);

  @$pb.TagNumber(3)
  SetDNSStatus get setDnsStatus => $_getN(1);
  @$pb.TagNumber(3)
  set setDnsStatus(SetDNSStatus value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasSetDnsStatus() => $_has(1);
  @$pb.TagNumber(3)
  void clearSetDnsStatus() => $_clearField(3);
}

class SetKillSwitchRequest extends $pb.GeneratedMessage {
  factory SetKillSwitchRequest({
    $core.bool? killSwitch,
  }) {
    final result = create();
    if (killSwitch != null) result.killSwitch = killSwitch;
    return result;
  }

  SetKillSwitchRequest._();

  factory SetKillSwitchRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetKillSwitchRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetKillSwitchRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOB(2, _omitFieldNames ? '' : 'killSwitch')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetKillSwitchRequest clone() =>
      SetKillSwitchRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetKillSwitchRequest copyWith(void Function(SetKillSwitchRequest) updates) =>
      super.copyWith((message) => updates(message as SetKillSwitchRequest))
          as SetKillSwitchRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetKillSwitchRequest create() => SetKillSwitchRequest._();
  @$core.override
  SetKillSwitchRequest createEmptyInstance() => create();
  static $pb.PbList<SetKillSwitchRequest> createRepeated() =>
      $pb.PbList<SetKillSwitchRequest>();
  @$core.pragma('dart2js:noInline')
  static SetKillSwitchRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetKillSwitchRequest>(create);
  static SetKillSwitchRequest? _defaultInstance;

  @$pb.TagNumber(2)
  $core.bool get killSwitch => $_getBF(0);
  @$pb.TagNumber(2)
  set killSwitch($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(2)
  $core.bool hasKillSwitch() => $_has(0);
  @$pb.TagNumber(2)
  void clearKillSwitch() => $_clearField(2);
}

class SetNotifyRequest extends $pb.GeneratedMessage {
  factory SetNotifyRequest({
    $core.bool? notify,
  }) {
    final result = create();
    if (notify != null) result.notify = notify;
    return result;
  }

  SetNotifyRequest._();

  factory SetNotifyRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetNotifyRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetNotifyRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOB(3, _omitFieldNames ? '' : 'notify')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetNotifyRequest clone() => SetNotifyRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetNotifyRequest copyWith(void Function(SetNotifyRequest) updates) =>
      super.copyWith((message) => updates(message as SetNotifyRequest))
          as SetNotifyRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetNotifyRequest create() => SetNotifyRequest._();
  @$core.override
  SetNotifyRequest createEmptyInstance() => create();
  static $pb.PbList<SetNotifyRequest> createRepeated() =>
      $pb.PbList<SetNotifyRequest>();
  @$core.pragma('dart2js:noInline')
  static SetNotifyRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetNotifyRequest>(create);
  static SetNotifyRequest? _defaultInstance;

  @$pb.TagNumber(3)
  $core.bool get notify => $_getBF(0);
  @$pb.TagNumber(3)
  set notify($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(3)
  $core.bool hasNotify() => $_has(0);
  @$pb.TagNumber(3)
  void clearNotify() => $_clearField(3);
}

class SetTrayRequest extends $pb.GeneratedMessage {
  factory SetTrayRequest({
    $fixnum.Int64? uid,
    $core.bool? tray,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    if (tray != null) result.tray = tray;
    return result;
  }

  SetTrayRequest._();

  factory SetTrayRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetTrayRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetTrayRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aInt64(2, _omitFieldNames ? '' : 'uid')
    ..aOB(3, _omitFieldNames ? '' : 'tray')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetTrayRequest clone() => SetTrayRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetTrayRequest copyWith(void Function(SetTrayRequest) updates) =>
      super.copyWith((message) => updates(message as SetTrayRequest))
          as SetTrayRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetTrayRequest create() => SetTrayRequest._();
  @$core.override
  SetTrayRequest createEmptyInstance() => create();
  static $pb.PbList<SetTrayRequest> createRepeated() =>
      $pb.PbList<SetTrayRequest>();
  @$core.pragma('dart2js:noInline')
  static SetTrayRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetTrayRequest>(create);
  static SetTrayRequest? _defaultInstance;

  @$pb.TagNumber(2)
  $fixnum.Int64 get uid => $_getI64(0);
  @$pb.TagNumber(2)
  set uid($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(2)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(2)
  void clearUid() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get tray => $_getBF(1);
  @$pb.TagNumber(3)
  set tray($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(3)
  $core.bool hasTray() => $_has(1);
  @$pb.TagNumber(3)
  void clearTray() => $_clearField(3);
}

class SetProtocolRequest extends $pb.GeneratedMessage {
  factory SetProtocolRequest({
    $0.Protocol? protocol,
  }) {
    final result = create();
    if (protocol != null) result.protocol = protocol;
    return result;
  }

  SetProtocolRequest._();

  factory SetProtocolRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetProtocolRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetProtocolRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..e<$0.Protocol>(2, _omitFieldNames ? '' : 'protocol', $pb.PbFieldType.OE,
        defaultOrMaker: $0.Protocol.UNKNOWN_PROTOCOL,
        valueOf: $0.Protocol.valueOf,
        enumValues: $0.Protocol.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetProtocolRequest clone() => SetProtocolRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetProtocolRequest copyWith(void Function(SetProtocolRequest) updates) =>
      super.copyWith((message) => updates(message as SetProtocolRequest))
          as SetProtocolRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetProtocolRequest create() => SetProtocolRequest._();
  @$core.override
  SetProtocolRequest createEmptyInstance() => create();
  static $pb.PbList<SetProtocolRequest> createRepeated() =>
      $pb.PbList<SetProtocolRequest>();
  @$core.pragma('dart2js:noInline')
  static SetProtocolRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetProtocolRequest>(create);
  static SetProtocolRequest? _defaultInstance;

  @$pb.TagNumber(2)
  $0.Protocol get protocol => $_getN(0);
  @$pb.TagNumber(2)
  set protocol($0.Protocol value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasProtocol() => $_has(0);
  @$pb.TagNumber(2)
  void clearProtocol() => $_clearField(2);
}

enum SetProtocolResponse_Response { errorCode, setProtocolStatus, notSet }

class SetProtocolResponse extends $pb.GeneratedMessage {
  factory SetProtocolResponse({
    SetErrorCode? errorCode,
    SetProtocolStatus? setProtocolStatus,
  }) {
    final result = create();
    if (errorCode != null) result.errorCode = errorCode;
    if (setProtocolStatus != null) result.setProtocolStatus = setProtocolStatus;
    return result;
  }

  SetProtocolResponse._();

  factory SetProtocolResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetProtocolResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, SetProtocolResponse_Response>
      _SetProtocolResponse_ResponseByTag = {
    1: SetProtocolResponse_Response.errorCode,
    2: SetProtocolResponse_Response.setProtocolStatus,
    0: SetProtocolResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetProtocolResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2])
    ..e<SetErrorCode>(1, _omitFieldNames ? '' : 'errorCode', $pb.PbFieldType.OE,
        defaultOrMaker: SetErrorCode.FAILURE,
        valueOf: SetErrorCode.valueOf,
        enumValues: SetErrorCode.values)
    ..e<SetProtocolStatus>(
        2, _omitFieldNames ? '' : 'setProtocolStatus', $pb.PbFieldType.OE,
        defaultOrMaker: SetProtocolStatus.PROTOCOL_CONFIGURED,
        valueOf: SetProtocolStatus.valueOf,
        enumValues: SetProtocolStatus.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetProtocolResponse clone() => SetProtocolResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetProtocolResponse copyWith(void Function(SetProtocolResponse) updates) =>
      super.copyWith((message) => updates(message as SetProtocolResponse))
          as SetProtocolResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetProtocolResponse create() => SetProtocolResponse._();
  @$core.override
  SetProtocolResponse createEmptyInstance() => create();
  static $pb.PbList<SetProtocolResponse> createRepeated() =>
      $pb.PbList<SetProtocolResponse>();
  @$core.pragma('dart2js:noInline')
  static SetProtocolResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetProtocolResponse>(create);
  static SetProtocolResponse? _defaultInstance;

  SetProtocolResponse_Response whichResponse() =>
      _SetProtocolResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  SetErrorCode get errorCode => $_getN(0);
  @$pb.TagNumber(1)
  set errorCode(SetErrorCode value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasErrorCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearErrorCode() => $_clearField(1);

  @$pb.TagNumber(2)
  SetProtocolStatus get setProtocolStatus => $_getN(1);
  @$pb.TagNumber(2)
  set setProtocolStatus(SetProtocolStatus value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasSetProtocolStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearSetProtocolStatus() => $_clearField(2);
}

class SetTechnologyRequest extends $pb.GeneratedMessage {
  factory SetTechnologyRequest({
    $1.Technology? technology,
  }) {
    final result = create();
    if (technology != null) result.technology = technology;
    return result;
  }

  SetTechnologyRequest._();

  factory SetTechnologyRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetTechnologyRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetTechnologyRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..e<$1.Technology>(
        2, _omitFieldNames ? '' : 'technology', $pb.PbFieldType.OE,
        defaultOrMaker: $1.Technology.UNKNOWN_TECHNOLOGY,
        valueOf: $1.Technology.valueOf,
        enumValues: $1.Technology.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetTechnologyRequest clone() =>
      SetTechnologyRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetTechnologyRequest copyWith(void Function(SetTechnologyRequest) updates) =>
      super.copyWith((message) => updates(message as SetTechnologyRequest))
          as SetTechnologyRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetTechnologyRequest create() => SetTechnologyRequest._();
  @$core.override
  SetTechnologyRequest createEmptyInstance() => create();
  static $pb.PbList<SetTechnologyRequest> createRepeated() =>
      $pb.PbList<SetTechnologyRequest>();
  @$core.pragma('dart2js:noInline')
  static SetTechnologyRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetTechnologyRequest>(create);
  static SetTechnologyRequest? _defaultInstance;

  @$pb.TagNumber(2)
  $1.Technology get technology => $_getN(0);
  @$pb.TagNumber(2)
  set technology($1.Technology value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasTechnology() => $_has(0);
  @$pb.TagNumber(2)
  void clearTechnology() => $_clearField(2);
}

class PortRange extends $pb.GeneratedMessage {
  factory PortRange({
    $fixnum.Int64? startPort,
    $fixnum.Int64? endPort,
  }) {
    final result = create();
    if (startPort != null) result.startPort = startPort;
    if (endPort != null) result.endPort = endPort;
    return result;
  }

  PortRange._();

  factory PortRange.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory PortRange.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PortRange',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aInt64(1, _omitFieldNames ? '' : 'startPort')
    ..aInt64(2, _omitFieldNames ? '' : 'endPort')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PortRange clone() => PortRange()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PortRange copyWith(void Function(PortRange) updates) =>
      super.copyWith((message) => updates(message as PortRange)) as PortRange;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PortRange create() => PortRange._();
  @$core.override
  PortRange createEmptyInstance() => create();
  static $pb.PbList<PortRange> createRepeated() => $pb.PbList<PortRange>();
  @$core.pragma('dart2js:noInline')
  static PortRange getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PortRange>(create);
  static PortRange? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get startPort => $_getI64(0);
  @$pb.TagNumber(1)
  set startPort($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasStartPort() => $_has(0);
  @$pb.TagNumber(1)
  void clearStartPort() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get endPort => $_getI64(1);
  @$pb.TagNumber(2)
  set endPort($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEndPort() => $_has(1);
  @$pb.TagNumber(2)
  void clearEndPort() => $_clearField(2);
}

class SetAllowlistSubnetRequest extends $pb.GeneratedMessage {
  factory SetAllowlistSubnetRequest({
    $core.String? subnet,
  }) {
    final result = create();
    if (subnet != null) result.subnet = subnet;
    return result;
  }

  SetAllowlistSubnetRequest._();

  factory SetAllowlistSubnetRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetAllowlistSubnetRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetAllowlistSubnetRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'subnet')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetAllowlistSubnetRequest clone() =>
      SetAllowlistSubnetRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetAllowlistSubnetRequest copyWith(
          void Function(SetAllowlistSubnetRequest) updates) =>
      super.copyWith((message) => updates(message as SetAllowlistSubnetRequest))
          as SetAllowlistSubnetRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetAllowlistSubnetRequest create() => SetAllowlistSubnetRequest._();
  @$core.override
  SetAllowlistSubnetRequest createEmptyInstance() => create();
  static $pb.PbList<SetAllowlistSubnetRequest> createRepeated() =>
      $pb.PbList<SetAllowlistSubnetRequest>();
  @$core.pragma('dart2js:noInline')
  static SetAllowlistSubnetRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetAllowlistSubnetRequest>(create);
  static SetAllowlistSubnetRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get subnet => $_getSZ(0);
  @$pb.TagNumber(1)
  set subnet($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSubnet() => $_has(0);
  @$pb.TagNumber(1)
  void clearSubnet() => $_clearField(1);
}

class SetAllowlistPortsRequest extends $pb.GeneratedMessage {
  factory SetAllowlistPortsRequest({
    $core.bool? isUdp,
    $core.bool? isTcp,
    PortRange? portRange,
  }) {
    final result = create();
    if (isUdp != null) result.isUdp = isUdp;
    if (isTcp != null) result.isTcp = isTcp;
    if (portRange != null) result.portRange = portRange;
    return result;
  }

  SetAllowlistPortsRequest._();

  factory SetAllowlistPortsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetAllowlistPortsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetAllowlistPortsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'isUdp')
    ..aOB(2, _omitFieldNames ? '' : 'isTcp')
    ..aOM<PortRange>(3, _omitFieldNames ? '' : 'portRange',
        subBuilder: PortRange.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetAllowlistPortsRequest clone() =>
      SetAllowlistPortsRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetAllowlistPortsRequest copyWith(
          void Function(SetAllowlistPortsRequest) updates) =>
      super.copyWith((message) => updates(message as SetAllowlistPortsRequest))
          as SetAllowlistPortsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetAllowlistPortsRequest create() => SetAllowlistPortsRequest._();
  @$core.override
  SetAllowlistPortsRequest createEmptyInstance() => create();
  static $pb.PbList<SetAllowlistPortsRequest> createRepeated() =>
      $pb.PbList<SetAllowlistPortsRequest>();
  @$core.pragma('dart2js:noInline')
  static SetAllowlistPortsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetAllowlistPortsRequest>(create);
  static SetAllowlistPortsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get isUdp => $_getBF(0);
  @$pb.TagNumber(1)
  set isUdp($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasIsUdp() => $_has(0);
  @$pb.TagNumber(1)
  void clearIsUdp() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get isTcp => $_getBF(1);
  @$pb.TagNumber(2)
  set isTcp($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasIsTcp() => $_has(1);
  @$pb.TagNumber(2)
  void clearIsTcp() => $_clearField(2);

  @$pb.TagNumber(3)
  PortRange get portRange => $_getN(2);
  @$pb.TagNumber(3)
  set portRange(PortRange value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasPortRange() => $_has(2);
  @$pb.TagNumber(3)
  void clearPortRange() => $_clearField(3);
  @$pb.TagNumber(3)
  PortRange ensurePortRange() => $_ensure(2);
}

enum SetAllowlistRequest_Request {
  setAllowlistSubnetRequest,
  setAllowlistPortsRequest,
  notSet
}

class SetAllowlistRequest extends $pb.GeneratedMessage {
  factory SetAllowlistRequest({
    SetAllowlistSubnetRequest? setAllowlistSubnetRequest,
    SetAllowlistPortsRequest? setAllowlistPortsRequest,
  }) {
    final result = create();
    if (setAllowlistSubnetRequest != null)
      result.setAllowlistSubnetRequest = setAllowlistSubnetRequest;
    if (setAllowlistPortsRequest != null)
      result.setAllowlistPortsRequest = setAllowlistPortsRequest;
    return result;
  }

  SetAllowlistRequest._();

  factory SetAllowlistRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetAllowlistRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, SetAllowlistRequest_Request>
      _SetAllowlistRequest_RequestByTag = {
    1: SetAllowlistRequest_Request.setAllowlistSubnetRequest,
    2: SetAllowlistRequest_Request.setAllowlistPortsRequest,
    0: SetAllowlistRequest_Request.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetAllowlistRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2])
    ..aOM<SetAllowlistSubnetRequest>(
        1, _omitFieldNames ? '' : 'setAllowlistSubnetRequest',
        subBuilder: SetAllowlistSubnetRequest.create)
    ..aOM<SetAllowlistPortsRequest>(
        2, _omitFieldNames ? '' : 'setAllowlistPortsRequest',
        subBuilder: SetAllowlistPortsRequest.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetAllowlistRequest clone() => SetAllowlistRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetAllowlistRequest copyWith(void Function(SetAllowlistRequest) updates) =>
      super.copyWith((message) => updates(message as SetAllowlistRequest))
          as SetAllowlistRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetAllowlistRequest create() => SetAllowlistRequest._();
  @$core.override
  SetAllowlistRequest createEmptyInstance() => create();
  static $pb.PbList<SetAllowlistRequest> createRepeated() =>
      $pb.PbList<SetAllowlistRequest>();
  @$core.pragma('dart2js:noInline')
  static SetAllowlistRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetAllowlistRequest>(create);
  static SetAllowlistRequest? _defaultInstance;

  SetAllowlistRequest_Request whichRequest() =>
      _SetAllowlistRequest_RequestByTag[$_whichOneof(0)]!;
  void clearRequest() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  SetAllowlistSubnetRequest get setAllowlistSubnetRequest => $_getN(0);
  @$pb.TagNumber(1)
  set setAllowlistSubnetRequest(SetAllowlistSubnetRequest value) =>
      $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSetAllowlistSubnetRequest() => $_has(0);
  @$pb.TagNumber(1)
  void clearSetAllowlistSubnetRequest() => $_clearField(1);
  @$pb.TagNumber(1)
  SetAllowlistSubnetRequest ensureSetAllowlistSubnetRequest() => $_ensure(0);

  @$pb.TagNumber(2)
  SetAllowlistPortsRequest get setAllowlistPortsRequest => $_getN(1);
  @$pb.TagNumber(2)
  set setAllowlistPortsRequest(SetAllowlistPortsRequest value) =>
      $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasSetAllowlistPortsRequest() => $_has(1);
  @$pb.TagNumber(2)
  void clearSetAllowlistPortsRequest() => $_clearField(2);
  @$pb.TagNumber(2)
  SetAllowlistPortsRequest ensureSetAllowlistPortsRequest() => $_ensure(1);
}

class SetLANDiscoveryRequest extends $pb.GeneratedMessage {
  factory SetLANDiscoveryRequest({
    $core.bool? enabled,
  }) {
    final result = create();
    if (enabled != null) result.enabled = enabled;
    return result;
  }

  SetLANDiscoveryRequest._();

  factory SetLANDiscoveryRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetLANDiscoveryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetLANDiscoveryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'enabled')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetLANDiscoveryRequest clone() =>
      SetLANDiscoveryRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetLANDiscoveryRequest copyWith(
          void Function(SetLANDiscoveryRequest) updates) =>
      super.copyWith((message) => updates(message as SetLANDiscoveryRequest))
          as SetLANDiscoveryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetLANDiscoveryRequest create() => SetLANDiscoveryRequest._();
  @$core.override
  SetLANDiscoveryRequest createEmptyInstance() => create();
  static $pb.PbList<SetLANDiscoveryRequest> createRepeated() =>
      $pb.PbList<SetLANDiscoveryRequest>();
  @$core.pragma('dart2js:noInline')
  static SetLANDiscoveryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetLANDiscoveryRequest>(create);
  static SetLANDiscoveryRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get enabled => $_getBF(0);
  @$pb.TagNumber(1)
  set enabled($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEnabled() => $_has(0);
  @$pb.TagNumber(1)
  void clearEnabled() => $_clearField(1);
}

enum SetLANDiscoveryResponse_Response {
  errorCode,
  setLanDiscoveryStatus,
  notSet
}

class SetLANDiscoveryResponse extends $pb.GeneratedMessage {
  factory SetLANDiscoveryResponse({
    SetErrorCode? errorCode,
    SetLANDiscoveryStatus? setLanDiscoveryStatus,
  }) {
    final result = create();
    if (errorCode != null) result.errorCode = errorCode;
    if (setLanDiscoveryStatus != null)
      result.setLanDiscoveryStatus = setLanDiscoveryStatus;
    return result;
  }

  SetLANDiscoveryResponse._();

  factory SetLANDiscoveryResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SetLANDiscoveryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, SetLANDiscoveryResponse_Response>
      _SetLANDiscoveryResponse_ResponseByTag = {
    1: SetLANDiscoveryResponse_Response.errorCode,
    2: SetLANDiscoveryResponse_Response.setLanDiscoveryStatus,
    0: SetLANDiscoveryResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetLANDiscoveryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2])
    ..e<SetErrorCode>(1, _omitFieldNames ? '' : 'errorCode', $pb.PbFieldType.OE,
        defaultOrMaker: SetErrorCode.FAILURE,
        valueOf: SetErrorCode.valueOf,
        enumValues: SetErrorCode.values)
    ..e<SetLANDiscoveryStatus>(
        2, _omitFieldNames ? '' : 'setLanDiscoveryStatus', $pb.PbFieldType.OE,
        defaultOrMaker: SetLANDiscoveryStatus.DISCOVERY_CONFIGURED,
        valueOf: SetLANDiscoveryStatus.valueOf,
        enumValues: SetLANDiscoveryStatus.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetLANDiscoveryResponse clone() =>
      SetLANDiscoveryResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetLANDiscoveryResponse copyWith(
          void Function(SetLANDiscoveryResponse) updates) =>
      super.copyWith((message) => updates(message as SetLANDiscoveryResponse))
          as SetLANDiscoveryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetLANDiscoveryResponse create() => SetLANDiscoveryResponse._();
  @$core.override
  SetLANDiscoveryResponse createEmptyInstance() => create();
  static $pb.PbList<SetLANDiscoveryResponse> createRepeated() =>
      $pb.PbList<SetLANDiscoveryResponse>();
  @$core.pragma('dart2js:noInline')
  static SetLANDiscoveryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetLANDiscoveryResponse>(create);
  static SetLANDiscoveryResponse? _defaultInstance;

  SetLANDiscoveryResponse_Response whichResponse() =>
      _SetLANDiscoveryResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  SetErrorCode get errorCode => $_getN(0);
  @$pb.TagNumber(1)
  set errorCode(SetErrorCode value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasErrorCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearErrorCode() => $_clearField(1);

  @$pb.TagNumber(2)
  SetLANDiscoveryStatus get setLanDiscoveryStatus => $_getN(1);
  @$pb.TagNumber(2)
  set setLanDiscoveryStatus(SetLANDiscoveryStatus value) =>
      $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasSetLanDiscoveryStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearSetLanDiscoveryStatus() => $_clearField(2);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
