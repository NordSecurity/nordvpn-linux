// This is a generated file - do not edit.
//
// Generated from token.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

class TokenInfoResponse extends $pb.GeneratedMessage {
  factory TokenInfoResponse({
    $fixnum.Int64? type,
    $core.String? token,
    $core.String? expiresAt,
    $core.String? trustedPassToken,
    $core.String? trustedPassOwnerId,
  }) {
    final result = create();
    if (type != null) result.type = type;
    if (token != null) result.token = token;
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (trustedPassToken != null) result.trustedPassToken = trustedPassToken;
    if (trustedPassOwnerId != null)
      result.trustedPassOwnerId = trustedPassOwnerId;
    return result;
  }

  TokenInfoResponse._();

  factory TokenInfoResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TokenInfoResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TokenInfoResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aInt64(1, _omitFieldNames ? '' : 'type')
    ..aOS(2, _omitFieldNames ? '' : 'token')
    ..aOS(3, _omitFieldNames ? '' : 'expiresAt')
    ..aOS(4, _omitFieldNames ? '' : 'trustedPassToken')
    ..aOS(5, _omitFieldNames ? '' : 'trustedPassOwnerId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TokenInfoResponse clone() => TokenInfoResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TokenInfoResponse copyWith(void Function(TokenInfoResponse) updates) =>
      super.copyWith((message) => updates(message as TokenInfoResponse))
          as TokenInfoResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TokenInfoResponse create() => TokenInfoResponse._();
  @$core.override
  TokenInfoResponse createEmptyInstance() => create();
  static $pb.PbList<TokenInfoResponse> createRepeated() =>
      $pb.PbList<TokenInfoResponse>();
  @$core.pragma('dart2js:noInline')
  static TokenInfoResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TokenInfoResponse>(create);
  static TokenInfoResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get type => $_getI64(0);
  @$pb.TagNumber(1)
  set type($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get token => $_getSZ(1);
  @$pb.TagNumber(2)
  set token($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearToken() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get expiresAt => $_getSZ(2);
  @$pb.TagNumber(3)
  set expiresAt($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasExpiresAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearExpiresAt() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get trustedPassToken => $_getSZ(3);
  @$pb.TagNumber(4)
  set trustedPassToken($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTrustedPassToken() => $_has(3);
  @$pb.TagNumber(4)
  void clearTrustedPassToken() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get trustedPassOwnerId => $_getSZ(4);
  @$pb.TagNumber(5)
  set trustedPassOwnerId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasTrustedPassOwnerId() => $_has(4);
  @$pb.TagNumber(5)
  void clearTrustedPassOwnerId() => $_clearField(5);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
