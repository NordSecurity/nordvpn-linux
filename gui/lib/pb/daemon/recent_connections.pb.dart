// This is a generated file - do not edit.
//
// Generated from recent_connections.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'config/group.pbenum.dart' as $0;
import 'server_selection_rule.pbenum.dart' as $1;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

class RecentConnectionsResponse extends $pb.GeneratedMessage {
  factory RecentConnectionsResponse({
    $core.Iterable<RecentConnectionModel>? connections,
  }) {
    final result = create();
    if (connections != null) result.connections.addAll(connections);
    return result;
  }

  RecentConnectionsResponse._();

  factory RecentConnectionsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RecentConnectionsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RecentConnectionsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..pPM<RecentConnectionModel>(1, _omitFieldNames ? '' : 'connections',
        subBuilder: RecentConnectionModel.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecentConnectionsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecentConnectionsResponse copyWith(
          void Function(RecentConnectionsResponse) updates) =>
      super.copyWith((message) => updates(message as RecentConnectionsResponse))
          as RecentConnectionsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RecentConnectionsResponse create() => RecentConnectionsResponse._();
  @$core.override
  RecentConnectionsResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static RecentConnectionsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RecentConnectionsResponse>(create);
  static RecentConnectionsResponse? _defaultInstance;

  /// List of recent connections information
  @$pb.TagNumber(1)
  $pb.PbList<RecentConnectionModel> get connections => $_getList(0);
}

class RecentConnectionModel extends $pb.GeneratedMessage {
  factory RecentConnectionModel({
    $core.String? country,
    $core.String? city,
    $0.ServerGroup? group,
    $core.String? countryCode,
    $core.String? specificServerName,
    $core.String? specificServer,
    $1.ServerSelectionRule? connectionType,
    $core.bool? isVirtual,
  }) {
    final result = create();
    if (country != null) result.country = country;
    if (city != null) result.city = city;
    if (group != null) result.group = group;
    if (countryCode != null) result.countryCode = countryCode;
    if (specificServerName != null)
      result.specificServerName = specificServerName;
    if (specificServer != null) result.specificServer = specificServer;
    if (connectionType != null) result.connectionType = connectionType;
    if (isVirtual != null) result.isVirtual = isVirtual;
    return result;
  }

  RecentConnectionModel._();

  factory RecentConnectionModel.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RecentConnectionModel.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RecentConnectionModel',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'country')
    ..aOS(2, _omitFieldNames ? '' : 'city')
    ..aE<$0.ServerGroup>(3, _omitFieldNames ? '' : 'group',
        enumValues: $0.ServerGroup.values)
    ..aOS(4, _omitFieldNames ? '' : 'countryCode')
    ..aOS(5, _omitFieldNames ? '' : 'specificServerName')
    ..aOS(6, _omitFieldNames ? '' : 'specificServer')
    ..aE<$1.ServerSelectionRule>(7, _omitFieldNames ? '' : 'connectionType',
        enumValues: $1.ServerSelectionRule.values)
    ..aOB(8, _omitFieldNames ? '' : 'isVirtual')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecentConnectionModel clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecentConnectionModel copyWith(
          void Function(RecentConnectionModel) updates) =>
      super.copyWith((message) => updates(message as RecentConnectionModel))
          as RecentConnectionModel;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RecentConnectionModel create() => RecentConnectionModel._();
  @$core.override
  RecentConnectionModel createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static RecentConnectionModel getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RecentConnectionModel>(create);
  static RecentConnectionModel? _defaultInstance;

  /// Country name
  @$pb.TagNumber(1)
  $core.String get country => $_getSZ(0);
  @$pb.TagNumber(1)
  set country($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCountry() => $_has(0);
  @$pb.TagNumber(1)
  void clearCountry() => $_clearField(1);

  /// City name
  @$pb.TagNumber(2)
  $core.String get city => $_getSZ(1);
  @$pb.TagNumber(2)
  set city($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCity() => $_has(1);
  @$pb.TagNumber(2)
  void clearCity() => $_clearField(2);

  /// Group name
  @$pb.TagNumber(3)
  $0.ServerGroup get group => $_getN(2);
  @$pb.TagNumber(3)
  set group($0.ServerGroup value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasGroup() => $_has(2);
  @$pb.TagNumber(3)
  void clearGroup() => $_clearField(3);

  /// Country code (2 characters)
  @$pb.TagNumber(4)
  $core.String get countryCode => $_getSZ(3);
  @$pb.TagNumber(4)
  set countryCode($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCountryCode() => $_has(3);
  @$pb.TagNumber(4)
  void clearCountryCode() => $_clearField(4);

  /// Human-readable server name
  @$pb.TagNumber(5)
  $core.String get specificServerName => $_getSZ(4);
  @$pb.TagNumber(5)
  set specificServerName($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSpecificServerName() => $_has(4);
  @$pb.TagNumber(5)
  void clearSpecificServerName() => $_clearField(5);

  /// Specific server identifier
  @$pb.TagNumber(6)
  $core.String get specificServer => $_getSZ(5);
  @$pb.TagNumber(6)
  set specificServer($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasSpecificServer() => $_has(5);
  @$pb.TagNumber(6)
  void clearSpecificServer() => $_clearField(6);

  /// Connection type enum
  @$pb.TagNumber(7)
  $1.ServerSelectionRule get connectionType => $_getN(6);
  @$pb.TagNumber(7)
  set connectionType($1.ServerSelectionRule value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasConnectionType() => $_has(6);
  @$pb.TagNumber(7)
  void clearConnectionType() => $_clearField(7);

  /// whether the server is virtual
  @$pb.TagNumber(8)
  $core.bool get isVirtual => $_getBF(7);
  @$pb.TagNumber(8)
  set isVirtual($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasIsVirtual() => $_has(7);
  @$pb.TagNumber(8)
  void clearIsVirtual() => $_clearField(8);
}

class RecentConnectionsRequest extends $pb.GeneratedMessage {
  factory RecentConnectionsRequest({
    $fixnum.Int64? limit,
  }) {
    final result = create();
    if (limit != null) result.limit = limit;
    return result;
  }

  RecentConnectionsRequest._();

  factory RecentConnectionsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RecentConnectionsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RecentConnectionsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aInt64(1, _omitFieldNames ? '' : 'limit')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecentConnectionsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RecentConnectionsRequest copyWith(
          void Function(RecentConnectionsRequest) updates) =>
      super.copyWith((message) => updates(message as RecentConnectionsRequest))
          as RecentConnectionsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RecentConnectionsRequest create() => RecentConnectionsRequest._();
  @$core.override
  RecentConnectionsRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static RecentConnectionsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RecentConnectionsRequest>(create);
  static RecentConnectionsRequest? _defaultInstance;

  /// Limit maximum entries to be returned
  @$pb.TagNumber(1)
  $fixnum.Int64 get limit => $_getI64(0);
  @$pb.TagNumber(1)
  set limit($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasLimit() => $_has(0);
  @$pb.TagNumber(1)
  void clearLimit() => $_clearField(1);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
