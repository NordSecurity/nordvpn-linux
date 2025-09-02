// This is a generated file - do not edit.
//
// Generated from status.proto.

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
import 'config/protocol.pbenum.dart' as $2;
import 'config/technology.pbenum.dart' as $1;
import 'status.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'status.pbenum.dart';

class ConnectionParameters extends $pb.GeneratedMessage {
  factory ConnectionParameters({
    ConnectionSource? source,
    $core.String? country,
    $core.String? city,
    $0.ServerGroup? group,
    $core.String? serverName,
    $core.String? countryCode,
  }) {
    final result = create();
    if (source != null) result.source = source;
    if (country != null) result.country = country;
    if (city != null) result.city = city;
    if (group != null) result.group = group;
    if (serverName != null) result.serverName = serverName;
    if (countryCode != null) result.countryCode = countryCode;
    return result;
  }

  ConnectionParameters._();

  factory ConnectionParameters.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ConnectionParameters.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ConnectionParameters',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..e<ConnectionSource>(
        1, _omitFieldNames ? '' : 'source', $pb.PbFieldType.OE,
        defaultOrMaker: ConnectionSource.UNKNOWN_SOURCE,
        valueOf: ConnectionSource.valueOf,
        enumValues: ConnectionSource.values)
    ..aOS(2, _omitFieldNames ? '' : 'country')
    ..aOS(3, _omitFieldNames ? '' : 'city')
    ..e<$0.ServerGroup>(4, _omitFieldNames ? '' : 'group', $pb.PbFieldType.OE,
        defaultOrMaker: $0.ServerGroup.UNDEFINED,
        valueOf: $0.ServerGroup.valueOf,
        enumValues: $0.ServerGroup.values)
    ..aOS(5, _omitFieldNames ? '' : 'serverName')
    ..aOS(6, _omitFieldNames ? '' : 'countryCode')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConnectionParameters clone() =>
      ConnectionParameters()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConnectionParameters copyWith(void Function(ConnectionParameters) updates) =>
      super.copyWith((message) => updates(message as ConnectionParameters))
          as ConnectionParameters;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ConnectionParameters create() => ConnectionParameters._();
  @$core.override
  ConnectionParameters createEmptyInstance() => create();
  static $pb.PbList<ConnectionParameters> createRepeated() =>
      $pb.PbList<ConnectionParameters>();
  @$core.pragma('dart2js:noInline')
  static ConnectionParameters getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ConnectionParameters>(create);
  static ConnectionParameters? _defaultInstance;

  @$pb.TagNumber(1)
  ConnectionSource get source => $_getN(0);
  @$pb.TagNumber(1)
  set source(ConnectionSource value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSource() => $_has(0);
  @$pb.TagNumber(1)
  void clearSource() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get country => $_getSZ(1);
  @$pb.TagNumber(2)
  set country($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCountry() => $_has(1);
  @$pb.TagNumber(2)
  void clearCountry() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get city => $_getSZ(2);
  @$pb.TagNumber(3)
  set city($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCity() => $_has(2);
  @$pb.TagNumber(3)
  void clearCity() => $_clearField(3);

  @$pb.TagNumber(4)
  $0.ServerGroup get group => $_getN(3);
  @$pb.TagNumber(4)
  set group($0.ServerGroup value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasGroup() => $_has(3);
  @$pb.TagNumber(4)
  void clearGroup() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get serverName => $_getSZ(4);
  @$pb.TagNumber(5)
  set serverName($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasServerName() => $_has(4);
  @$pb.TagNumber(5)
  void clearServerName() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get countryCode => $_getSZ(5);
  @$pb.TagNumber(6)
  set countryCode($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCountryCode() => $_has(5);
  @$pb.TagNumber(6)
  void clearCountryCode() => $_clearField(6);
}

class StatusResponse extends $pb.GeneratedMessage {
  factory StatusResponse({
    ConnectionState? state,
    $1.Technology? technology,
    $2.Protocol? protocol,
    $core.String? ip,
    $core.String? hostname,
    $core.String? country,
    $core.String? city,
    $fixnum.Int64? download,
    $fixnum.Int64? upload,
    $fixnum.Int64? uptime,
    $core.String? name,
    $core.bool? virtualLocation,
    ConnectionParameters? parameters,
    $core.bool? postQuantum,
    $core.bool? isMeshPeer,
    $core.bool? byUser,
    $core.String? countryCode,
    $core.bool? obfuscated,
  }) {
    final result = create();
    if (state != null) result.state = state;
    if (technology != null) result.technology = technology;
    if (protocol != null) result.protocol = protocol;
    if (ip != null) result.ip = ip;
    if (hostname != null) result.hostname = hostname;
    if (country != null) result.country = country;
    if (city != null) result.city = city;
    if (download != null) result.download = download;
    if (upload != null) result.upload = upload;
    if (uptime != null) result.uptime = uptime;
    if (name != null) result.name = name;
    if (virtualLocation != null) result.virtualLocation = virtualLocation;
    if (parameters != null) result.parameters = parameters;
    if (postQuantum != null) result.postQuantum = postQuantum;
    if (isMeshPeer != null) result.isMeshPeer = isMeshPeer;
    if (byUser != null) result.byUser = byUser;
    if (countryCode != null) result.countryCode = countryCode;
    if (obfuscated != null) result.obfuscated = obfuscated;
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
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..e<ConnectionState>(1, _omitFieldNames ? '' : 'state', $pb.PbFieldType.OE,
        defaultOrMaker: ConnectionState.UNKNOWN_STATE,
        valueOf: ConnectionState.valueOf,
        enumValues: ConnectionState.values)
    ..e<$1.Technology>(
        2, _omitFieldNames ? '' : 'technology', $pb.PbFieldType.OE,
        defaultOrMaker: $1.Technology.UNKNOWN_TECHNOLOGY,
        valueOf: $1.Technology.valueOf,
        enumValues: $1.Technology.values)
    ..e<$2.Protocol>(3, _omitFieldNames ? '' : 'protocol', $pb.PbFieldType.OE,
        defaultOrMaker: $2.Protocol.UNKNOWN_PROTOCOL,
        valueOf: $2.Protocol.valueOf,
        enumValues: $2.Protocol.values)
    ..aOS(4, _omitFieldNames ? '' : 'ip')
    ..aOS(5, _omitFieldNames ? '' : 'hostname')
    ..aOS(6, _omitFieldNames ? '' : 'country')
    ..aOS(7, _omitFieldNames ? '' : 'city')
    ..a<$fixnum.Int64>(
        8, _omitFieldNames ? '' : 'download', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(9, _omitFieldNames ? '' : 'upload', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..aInt64(10, _omitFieldNames ? '' : 'uptime')
    ..aOS(11, _omitFieldNames ? '' : 'name')
    ..aOB(12, _omitFieldNames ? '' : 'virtualLocation',
        protoName: 'virtualLocation')
    ..aOM<ConnectionParameters>(13, _omitFieldNames ? '' : 'parameters',
        subBuilder: ConnectionParameters.create)
    ..aOB(14, _omitFieldNames ? '' : 'postQuantum', protoName: 'postQuantum')
    ..aOB(15, _omitFieldNames ? '' : 'isMeshPeer')
    ..aOB(16, _omitFieldNames ? '' : 'byUser')
    ..aOS(17, _omitFieldNames ? '' : 'countryCode')
    ..aOB(18, _omitFieldNames ? '' : 'obfuscated')
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
  ConnectionState get state => $_getN(0);
  @$pb.TagNumber(1)
  set state(ConnectionState value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasState() => $_has(0);
  @$pb.TagNumber(1)
  void clearState() => $_clearField(1);

  @$pb.TagNumber(2)
  $1.Technology get technology => $_getN(1);
  @$pb.TagNumber(2)
  set technology($1.Technology value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasTechnology() => $_has(1);
  @$pb.TagNumber(2)
  void clearTechnology() => $_clearField(2);

  @$pb.TagNumber(3)
  $2.Protocol get protocol => $_getN(2);
  @$pb.TagNumber(3)
  set protocol($2.Protocol value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasProtocol() => $_has(2);
  @$pb.TagNumber(3)
  void clearProtocol() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get ip => $_getSZ(3);
  @$pb.TagNumber(4)
  set ip($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIp() => $_has(3);
  @$pb.TagNumber(4)
  void clearIp() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get hostname => $_getSZ(4);
  @$pb.TagNumber(5)
  set hostname($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasHostname() => $_has(4);
  @$pb.TagNumber(5)
  void clearHostname() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get country => $_getSZ(5);
  @$pb.TagNumber(6)
  set country($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCountry() => $_has(5);
  @$pb.TagNumber(6)
  void clearCountry() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get city => $_getSZ(6);
  @$pb.TagNumber(7)
  set city($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasCity() => $_has(6);
  @$pb.TagNumber(7)
  void clearCity() => $_clearField(7);

  @$pb.TagNumber(8)
  $fixnum.Int64 get download => $_getI64(7);
  @$pb.TagNumber(8)
  set download($fixnum.Int64 value) => $_setInt64(7, value);
  @$pb.TagNumber(8)
  $core.bool hasDownload() => $_has(7);
  @$pb.TagNumber(8)
  void clearDownload() => $_clearField(8);

  @$pb.TagNumber(9)
  $fixnum.Int64 get upload => $_getI64(8);
  @$pb.TagNumber(9)
  set upload($fixnum.Int64 value) => $_setInt64(8, value);
  @$pb.TagNumber(9)
  $core.bool hasUpload() => $_has(8);
  @$pb.TagNumber(9)
  void clearUpload() => $_clearField(9);

  @$pb.TagNumber(10)
  $fixnum.Int64 get uptime => $_getI64(9);
  @$pb.TagNumber(10)
  set uptime($fixnum.Int64 value) => $_setInt64(9, value);
  @$pb.TagNumber(10)
  $core.bool hasUptime() => $_has(9);
  @$pb.TagNumber(10)
  void clearUptime() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.String get name => $_getSZ(10);
  @$pb.TagNumber(11)
  set name($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasName() => $_has(10);
  @$pb.TagNumber(11)
  void clearName() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.bool get virtualLocation => $_getBF(11);
  @$pb.TagNumber(12)
  set virtualLocation($core.bool value) => $_setBool(11, value);
  @$pb.TagNumber(12)
  $core.bool hasVirtualLocation() => $_has(11);
  @$pb.TagNumber(12)
  void clearVirtualLocation() => $_clearField(12);

  @$pb.TagNumber(13)
  ConnectionParameters get parameters => $_getN(12);
  @$pb.TagNumber(13)
  set parameters(ConnectionParameters value) => $_setField(13, value);
  @$pb.TagNumber(13)
  $core.bool hasParameters() => $_has(12);
  @$pb.TagNumber(13)
  void clearParameters() => $_clearField(13);
  @$pb.TagNumber(13)
  ConnectionParameters ensureParameters() => $_ensure(12);

  @$pb.TagNumber(14)
  $core.bool get postQuantum => $_getBF(13);
  @$pb.TagNumber(14)
  set postQuantum($core.bool value) => $_setBool(13, value);
  @$pb.TagNumber(14)
  $core.bool hasPostQuantum() => $_has(13);
  @$pb.TagNumber(14)
  void clearPostQuantum() => $_clearField(14);

  @$pb.TagNumber(15)
  $core.bool get isMeshPeer => $_getBF(14);
  @$pb.TagNumber(15)
  set isMeshPeer($core.bool value) => $_setBool(14, value);
  @$pb.TagNumber(15)
  $core.bool hasIsMeshPeer() => $_has(14);
  @$pb.TagNumber(15)
  void clearIsMeshPeer() => $_clearField(15);

  @$pb.TagNumber(16)
  $core.bool get byUser => $_getBF(15);
  @$pb.TagNumber(16)
  set byUser($core.bool value) => $_setBool(15, value);
  @$pb.TagNumber(16)
  $core.bool hasByUser() => $_has(15);
  @$pb.TagNumber(16)
  void clearByUser() => $_clearField(16);

  @$pb.TagNumber(17)
  $core.String get countryCode => $_getSZ(16);
  @$pb.TagNumber(17)
  set countryCode($core.String value) => $_setString(16, value);
  @$pb.TagNumber(17)
  $core.bool hasCountryCode() => $_has(16);
  @$pb.TagNumber(17)
  void clearCountryCode() => $_clearField(17);

  @$pb.TagNumber(18)
  $core.bool get obfuscated => $_getBF(17);
  @$pb.TagNumber(18)
  set obfuscated($core.bool value) => $_setBool(17, value);
  @$pb.TagNumber(18)
  $core.bool hasObfuscated() => $_has(17);
  @$pb.TagNumber(18)
  void clearObfuscated() => $_clearField(18);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
