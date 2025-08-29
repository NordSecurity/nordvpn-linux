// This is a generated file - do not edit.
//
// Generated from servers.proto.

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
import 'servers.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'servers.pbenum.dart';

class Server extends $pb.GeneratedMessage {
  factory Server({
    $fixnum.Int64? id,
    $core.String? hostName,
    $core.bool? virtual,
    $core.Iterable<$0.ServerGroup>? serverGroups,
    $core.Iterable<Technology>? technologies,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (hostName != null) result.hostName = hostName;
    if (virtual != null) result.virtual = virtual;
    if (serverGroups != null) result.serverGroups.addAll(serverGroups);
    if (technologies != null) result.technologies.addAll(technologies);
    return result;
  }

  Server._();

  factory Server.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Server.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Server',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aInt64(1, _omitFieldNames ? '' : 'id')
    ..aOS(4, _omitFieldNames ? '' : 'hostName')
    ..aOB(5, _omitFieldNames ? '' : 'virtual')
    ..pc<$0.ServerGroup>(
        6, _omitFieldNames ? '' : 'serverGroups', $pb.PbFieldType.KE,
        valueOf: $0.ServerGroup.valueOf,
        enumValues: $0.ServerGroup.values,
        defaultEnumValue: $0.ServerGroup.UNDEFINED)
    ..pc<Technology>(
        7, _omitFieldNames ? '' : 'technologies', $pb.PbFieldType.KE,
        valueOf: Technology.valueOf,
        enumValues: Technology.values,
        defaultEnumValue: Technology.UNKNOWN_TECHNLOGY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Server clone() => Server()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Server copyWith(void Function(Server) updates) =>
      super.copyWith((message) => updates(message as Server)) as Server;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Server create() => Server._();
  @$core.override
  Server createEmptyInstance() => create();
  static $pb.PbList<Server> createRepeated() => $pb.PbList<Server>();
  @$core.pragma('dart2js:noInline')
  static Server getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Server>(create);
  static Server? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get id => $_getI64(0);
  @$pb.TagNumber(1)
  set id($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(4)
  $core.String get hostName => $_getSZ(1);
  @$pb.TagNumber(4)
  set hostName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(4)
  $core.bool hasHostName() => $_has(1);
  @$pb.TagNumber(4)
  void clearHostName() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get virtual => $_getBF(2);
  @$pb.TagNumber(5)
  set virtual($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(5)
  $core.bool hasVirtual() => $_has(2);
  @$pb.TagNumber(5)
  void clearVirtual() => $_clearField(5);

  @$pb.TagNumber(6)
  $pb.PbList<$0.ServerGroup> get serverGroups => $_getList(3);

  @$pb.TagNumber(7)
  $pb.PbList<Technology> get technologies => $_getList(4);
}

class ServerCity extends $pb.GeneratedMessage {
  factory ServerCity({
    $core.String? cityName,
    $core.Iterable<Server>? servers,
  }) {
    final result = create();
    if (cityName != null) result.cityName = cityName;
    if (servers != null) result.servers.addAll(servers);
    return result;
  }

  ServerCity._();

  factory ServerCity.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ServerCity.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ServerCity',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'cityName')
    ..pc<Server>(2, _omitFieldNames ? '' : 'servers', $pb.PbFieldType.PM,
        subBuilder: Server.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServerCity clone() => ServerCity()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServerCity copyWith(void Function(ServerCity) updates) =>
      super.copyWith((message) => updates(message as ServerCity)) as ServerCity;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ServerCity create() => ServerCity._();
  @$core.override
  ServerCity createEmptyInstance() => create();
  static $pb.PbList<ServerCity> createRepeated() => $pb.PbList<ServerCity>();
  @$core.pragma('dart2js:noInline')
  static ServerCity getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ServerCity>(create);
  static ServerCity? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get cityName => $_getSZ(0);
  @$pb.TagNumber(1)
  set cityName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCityName() => $_has(0);
  @$pb.TagNumber(1)
  void clearCityName() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<Server> get servers => $_getList(1);
}

class ServerCountry extends $pb.GeneratedMessage {
  factory ServerCountry({
    $core.String? countryCode,
    $core.Iterable<ServerCity>? cities,
    $core.String? countryName,
  }) {
    final result = create();
    if (countryCode != null) result.countryCode = countryCode;
    if (cities != null) result.cities.addAll(cities);
    if (countryName != null) result.countryName = countryName;
    return result;
  }

  ServerCountry._();

  factory ServerCountry.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ServerCountry.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ServerCountry',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'countryCode')
    ..pc<ServerCity>(2, _omitFieldNames ? '' : 'cities', $pb.PbFieldType.PM,
        subBuilder: ServerCity.create)
    ..aOS(3, _omitFieldNames ? '' : 'countryName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServerCountry clone() => ServerCountry()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServerCountry copyWith(void Function(ServerCountry) updates) =>
      super.copyWith((message) => updates(message as ServerCountry))
          as ServerCountry;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ServerCountry create() => ServerCountry._();
  @$core.override
  ServerCountry createEmptyInstance() => create();
  static $pb.PbList<ServerCountry> createRepeated() =>
      $pb.PbList<ServerCountry>();
  @$core.pragma('dart2js:noInline')
  static ServerCountry getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ServerCountry>(create);
  static ServerCountry? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get countryCode => $_getSZ(0);
  @$pb.TagNumber(1)
  set countryCode($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCountryCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearCountryCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<ServerCity> get cities => $_getList(1);

  @$pb.TagNumber(3)
  $core.String get countryName => $_getSZ(2);
  @$pb.TagNumber(3)
  set countryName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCountryName() => $_has(2);
  @$pb.TagNumber(3)
  void clearCountryName() => $_clearField(3);
}

class ServersMap extends $pb.GeneratedMessage {
  factory ServersMap({
    $core.Iterable<ServerCountry>? serversByCountry,
  }) {
    final result = create();
    if (serversByCountry != null)
      result.serversByCountry.addAll(serversByCountry);
    return result;
  }

  ServersMap._();

  factory ServersMap.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ServersMap.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ServersMap',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..pc<ServerCountry>(
        1, _omitFieldNames ? '' : 'serversByCountry', $pb.PbFieldType.PM,
        subBuilder: ServerCountry.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServersMap clone() => ServersMap()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServersMap copyWith(void Function(ServersMap) updates) =>
      super.copyWith((message) => updates(message as ServersMap)) as ServersMap;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ServersMap create() => ServersMap._();
  @$core.override
  ServersMap createEmptyInstance() => create();
  static $pb.PbList<ServersMap> createRepeated() => $pb.PbList<ServersMap>();
  @$core.pragma('dart2js:noInline')
  static ServersMap getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ServersMap>(create);
  static ServersMap? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ServerCountry> get serversByCountry => $_getList(0);
}

enum ServersResponse_Response { servers, error, notSet }

class ServersResponse extends $pb.GeneratedMessage {
  factory ServersResponse({
    ServersMap? servers,
    ServersError? error,
  }) {
    final result = create();
    if (servers != null) result.servers = servers;
    if (error != null) result.error = error;
    return result;
  }

  ServersResponse._();

  factory ServersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ServersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, ServersResponse_Response>
      _ServersResponse_ResponseByTag = {
    1: ServersResponse_Response.servers,
    2: ServersResponse_Response.error,
    0: ServersResponse_Response.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ServersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..oo(0, [1, 2])
    ..aOM<ServersMap>(1, _omitFieldNames ? '' : 'servers',
        subBuilder: ServersMap.create)
    ..e<ServersError>(2, _omitFieldNames ? '' : 'error', $pb.PbFieldType.OE,
        defaultOrMaker: ServersError.NO_ERROR,
        valueOf: ServersError.valueOf,
        enumValues: ServersError.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServersResponse clone() => ServersResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServersResponse copyWith(void Function(ServersResponse) updates) =>
      super.copyWith((message) => updates(message as ServersResponse))
          as ServersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ServersResponse create() => ServersResponse._();
  @$core.override
  ServersResponse createEmptyInstance() => create();
  static $pb.PbList<ServersResponse> createRepeated() =>
      $pb.PbList<ServersResponse>();
  @$core.pragma('dart2js:noInline')
  static ServersResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ServersResponse>(create);
  static ServersResponse? _defaultInstance;

  ServersResponse_Response whichResponse() =>
      _ServersResponse_ResponseByTag[$_whichOneof(0)]!;
  void clearResponse() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  ServersMap get servers => $_getN(0);
  @$pb.TagNumber(1)
  set servers(ServersMap value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasServers() => $_has(0);
  @$pb.TagNumber(1)
  void clearServers() => $_clearField(1);
  @$pb.TagNumber(1)
  ServersMap ensureServers() => $_ensure(0);

  @$pb.TagNumber(2)
  ServersError get error => $_getN(1);
  @$pb.TagNumber(2)
  set error(ServersError value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasError() => $_has(1);
  @$pb.TagNumber(2)
  void clearError() => $_clearField(2);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
