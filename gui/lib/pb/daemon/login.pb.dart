// This is a generated file - do not edit.
//
// Generated from login.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'login.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'login.pbenum.dart';

class LoginOAuth2Request extends $pb.GeneratedMessage {
  factory LoginOAuth2Request({
    LoginType? type,
  }) {
    final result = create();
    if (type != null) result.type = type;
    return result;
  }

  LoginOAuth2Request._();

  factory LoginOAuth2Request.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LoginOAuth2Request.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginOAuth2Request',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..e<LoginType>(1, _omitFieldNames ? '' : 'type', $pb.PbFieldType.OE,
        defaultOrMaker: LoginType.LoginType_UNKNOWN,
        valueOf: LoginType.valueOf,
        enumValues: LoginType.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginOAuth2Request clone() => LoginOAuth2Request()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginOAuth2Request copyWith(void Function(LoginOAuth2Request) updates) =>
      super.copyWith((message) => updates(message as LoginOAuth2Request))
          as LoginOAuth2Request;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LoginOAuth2Request create() => LoginOAuth2Request._();
  @$core.override
  LoginOAuth2Request createEmptyInstance() => create();
  static $pb.PbList<LoginOAuth2Request> createRepeated() =>
      $pb.PbList<LoginOAuth2Request>();
  @$core.pragma('dart2js:noInline')
  static LoginOAuth2Request getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LoginOAuth2Request>(create);
  static LoginOAuth2Request? _defaultInstance;

  @$pb.TagNumber(1)
  LoginType get type => $_getN(0);
  @$pb.TagNumber(1)
  set type(LoginType value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);
}

class LoginOAuth2CallbackRequest extends $pb.GeneratedMessage {
  factory LoginOAuth2CallbackRequest({
    $core.String? token,
    LoginType? type,
  }) {
    final result = create();
    if (token != null) result.token = token;
    if (type != null) result.type = type;
    return result;
  }

  LoginOAuth2CallbackRequest._();

  factory LoginOAuth2CallbackRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LoginOAuth2CallbackRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginOAuth2CallbackRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'token')
    ..e<LoginType>(2, _omitFieldNames ? '' : 'type', $pb.PbFieldType.OE,
        defaultOrMaker: LoginType.LoginType_UNKNOWN,
        valueOf: LoginType.valueOf,
        enumValues: LoginType.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginOAuth2CallbackRequest clone() =>
      LoginOAuth2CallbackRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginOAuth2CallbackRequest copyWith(
          void Function(LoginOAuth2CallbackRequest) updates) =>
      super.copyWith(
              (message) => updates(message as LoginOAuth2CallbackRequest))
          as LoginOAuth2CallbackRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LoginOAuth2CallbackRequest create() => LoginOAuth2CallbackRequest._();
  @$core.override
  LoginOAuth2CallbackRequest createEmptyInstance() => create();
  static $pb.PbList<LoginOAuth2CallbackRequest> createRepeated() =>
      $pb.PbList<LoginOAuth2CallbackRequest>();
  @$core.pragma('dart2js:noInline')
  static LoginOAuth2CallbackRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LoginOAuth2CallbackRequest>(create);
  static LoginOAuth2CallbackRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get token => $_getSZ(0);
  @$pb.TagNumber(1)
  set token($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => $_clearField(1);

  @$pb.TagNumber(2)
  LoginType get type => $_getN(1);
  @$pb.TagNumber(2)
  set type(LoginType value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => $_clearField(2);
}

class LoginResponse extends $pb.GeneratedMessage {
  factory LoginResponse({
    $fixnum.Int64? type,
    $core.String? url,
  }) {
    final result = create();
    if (type != null) result.type = type;
    if (url != null) result.url = url;
    return result;
  }

  LoginResponse._();

  factory LoginResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LoginResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aInt64(1, _omitFieldNames ? '' : 'type')
    ..aOS(5, _omitFieldNames ? '' : 'url')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginResponse clone() => LoginResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginResponse copyWith(void Function(LoginResponse) updates) =>
      super.copyWith((message) => updates(message as LoginResponse))
          as LoginResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LoginResponse create() => LoginResponse._();
  @$core.override
  LoginResponse createEmptyInstance() => create();
  static $pb.PbList<LoginResponse> createRepeated() =>
      $pb.PbList<LoginResponse>();
  @$core.pragma('dart2js:noInline')
  static LoginResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LoginResponse>(create);
  static LoginResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get type => $_getI64(0);
  @$pb.TagNumber(1)
  set type($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  @$pb.TagNumber(5)
  $core.String get url => $_getSZ(1);
  @$pb.TagNumber(5)
  set url($core.String value) => $_setString(1, value);
  @$pb.TagNumber(5)
  $core.bool hasUrl() => $_has(1);
  @$pb.TagNumber(5)
  void clearUrl() => $_clearField(5);
}

class LoginOAuth2Response extends $pb.GeneratedMessage {
  factory LoginOAuth2Response({
    LoginStatus? status,
    $core.String? url,
  }) {
    final result = create();
    if (status != null) result.status = status;
    if (url != null) result.url = url;
    return result;
  }

  LoginOAuth2Response._();

  factory LoginOAuth2Response.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LoginOAuth2Response.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginOAuth2Response',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..e<LoginStatus>(1, _omitFieldNames ? '' : 'status', $pb.PbFieldType.OE,
        defaultOrMaker: LoginStatus.SUCCESS,
        valueOf: LoginStatus.valueOf,
        enumValues: LoginStatus.values)
    ..aOS(2, _omitFieldNames ? '' : 'url')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginOAuth2Response clone() => LoginOAuth2Response()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginOAuth2Response copyWith(void Function(LoginOAuth2Response) updates) =>
      super.copyWith((message) => updates(message as LoginOAuth2Response))
          as LoginOAuth2Response;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LoginOAuth2Response create() => LoginOAuth2Response._();
  @$core.override
  LoginOAuth2Response createEmptyInstance() => create();
  static $pb.PbList<LoginOAuth2Response> createRepeated() =>
      $pb.PbList<LoginOAuth2Response>();
  @$core.pragma('dart2js:noInline')
  static LoginOAuth2Response getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LoginOAuth2Response>(create);
  static LoginOAuth2Response? _defaultInstance;

  @$pb.TagNumber(1)
  LoginStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(LoginStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get url => $_getSZ(1);
  @$pb.TagNumber(2)
  set url($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUrl() => $_has(1);
  @$pb.TagNumber(2)
  void clearUrl() => $_clearField(2);
}

class LoginOAuth2CallbackResponse extends $pb.GeneratedMessage {
  factory LoginOAuth2CallbackResponse({
    LoginStatus? status,
  }) {
    final result = create();
    if (status != null) result.status = status;
    return result;
  }

  LoginOAuth2CallbackResponse._();

  factory LoginOAuth2CallbackResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LoginOAuth2CallbackResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginOAuth2CallbackResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..e<LoginStatus>(1, _omitFieldNames ? '' : 'status', $pb.PbFieldType.OE,
        defaultOrMaker: LoginStatus.SUCCESS,
        valueOf: LoginStatus.valueOf,
        enumValues: LoginStatus.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginOAuth2CallbackResponse clone() =>
      LoginOAuth2CallbackResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginOAuth2CallbackResponse copyWith(
          void Function(LoginOAuth2CallbackResponse) updates) =>
      super.copyWith(
              (message) => updates(message as LoginOAuth2CallbackResponse))
          as LoginOAuth2CallbackResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LoginOAuth2CallbackResponse create() =>
      LoginOAuth2CallbackResponse._();
  @$core.override
  LoginOAuth2CallbackResponse createEmptyInstance() => create();
  static $pb.PbList<LoginOAuth2CallbackResponse> createRepeated() =>
      $pb.PbList<LoginOAuth2CallbackResponse>();
  @$core.pragma('dart2js:noInline')
  static LoginOAuth2CallbackResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LoginOAuth2CallbackResponse>(create);
  static LoginOAuth2CallbackResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LoginStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(LoginStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
}

class IsLoggedInResponse extends $pb.GeneratedMessage {
  factory IsLoggedInResponse({
    $core.bool? isLoggedIn,
    LoginStatus? status,
  }) {
    final result = create();
    if (isLoggedIn != null) result.isLoggedIn = isLoggedIn;
    if (status != null) result.status = status;
    return result;
  }

  IsLoggedInResponse._();

  factory IsLoggedInResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory IsLoggedInResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'IsLoggedInResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'isLoggedIn')
    ..e<LoginStatus>(2, _omitFieldNames ? '' : 'status', $pb.PbFieldType.OE,
        defaultOrMaker: LoginStatus.SUCCESS,
        valueOf: LoginStatus.valueOf,
        enumValues: LoginStatus.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IsLoggedInResponse clone() => IsLoggedInResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  IsLoggedInResponse copyWith(void Function(IsLoggedInResponse) updates) =>
      super.copyWith((message) => updates(message as IsLoggedInResponse))
          as IsLoggedInResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static IsLoggedInResponse create() => IsLoggedInResponse._();
  @$core.override
  IsLoggedInResponse createEmptyInstance() => create();
  static $pb.PbList<IsLoggedInResponse> createRepeated() =>
      $pb.PbList<IsLoggedInResponse>();
  @$core.pragma('dart2js:noInline')
  static IsLoggedInResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<IsLoggedInResponse>(create);
  static IsLoggedInResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get isLoggedIn => $_getBF(0);
  @$pb.TagNumber(1)
  set isLoggedIn($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasIsLoggedIn() => $_has(0);
  @$pb.TagNumber(1)
  void clearIsLoggedIn() => $_clearField(1);

  @$pb.TagNumber(2)
  LoginStatus get status => $_getN(1);
  @$pb.TagNumber(2)
  set status(LoginStatus value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearStatus() => $_clearField(2);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
