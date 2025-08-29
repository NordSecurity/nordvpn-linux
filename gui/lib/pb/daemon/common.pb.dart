// This is a generated file - do not edit.
//
// Generated from common.proto.

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

export 'common.pbenum.dart';

class GetDaemonApiVersionRequest extends $pb.GeneratedMessage {
  factory GetDaemonApiVersionRequest() => create();

  GetDaemonApiVersionRequest._();

  factory GetDaemonApiVersionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetDaemonApiVersionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetDaemonApiVersionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDaemonApiVersionRequest clone() =>
      GetDaemonApiVersionRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDaemonApiVersionRequest copyWith(
          void Function(GetDaemonApiVersionRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetDaemonApiVersionRequest))
          as GetDaemonApiVersionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetDaemonApiVersionRequest create() => GetDaemonApiVersionRequest._();
  @$core.override
  GetDaemonApiVersionRequest createEmptyInstance() => create();
  static $pb.PbList<GetDaemonApiVersionRequest> createRepeated() =>
      $pb.PbList<GetDaemonApiVersionRequest>();
  @$core.pragma('dart2js:noInline')
  static GetDaemonApiVersionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetDaemonApiVersionRequest>(create);
  static GetDaemonApiVersionRequest? _defaultInstance;
}

class GetDaemonApiVersionResponse extends $pb.GeneratedMessage {
  factory GetDaemonApiVersionResponse({
    $core.int? apiVersion,
  }) {
    final result = create();
    if (apiVersion != null) result.apiVersion = apiVersion;
    return result;
  }

  GetDaemonApiVersionResponse._();

  factory GetDaemonApiVersionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetDaemonApiVersionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetDaemonApiVersionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'apiVersion', $pb.PbFieldType.OU3,
        protoName: 'apiVersion')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDaemonApiVersionResponse clone() =>
      GetDaemonApiVersionResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetDaemonApiVersionResponse copyWith(
          void Function(GetDaemonApiVersionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetDaemonApiVersionResponse))
          as GetDaemonApiVersionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetDaemonApiVersionResponse create() =>
      GetDaemonApiVersionResponse._();
  @$core.override
  GetDaemonApiVersionResponse createEmptyInstance() => create();
  static $pb.PbList<GetDaemonApiVersionResponse> createRepeated() =>
      $pb.PbList<GetDaemonApiVersionResponse>();
  @$core.pragma('dart2js:noInline')
  static GetDaemonApiVersionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetDaemonApiVersionResponse>(create);
  static GetDaemonApiVersionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get apiVersion => $_getIZ(0);
  @$pb.TagNumber(1)
  set apiVersion($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasApiVersion() => $_has(0);
  @$pb.TagNumber(1)
  void clearApiVersion() => $_clearField(1);
}

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
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
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

class Bool extends $pb.GeneratedMessage {
  factory Bool({
    $core.bool? value,
  }) {
    final result = create();
    if (value != null) result.value = value;
    return result;
  }

  Bool._();

  factory Bool.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Bool.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Bool',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'value')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Bool clone() => Bool()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Bool copyWith(void Function(Bool) updates) =>
      super.copyWith((message) => updates(message as Bool)) as Bool;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Bool create() => Bool._();
  @$core.override
  Bool createEmptyInstance() => create();
  static $pb.PbList<Bool> createRepeated() => $pb.PbList<Bool>();
  @$core.pragma('dart2js:noInline')
  static Bool getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Bool>(create);
  static Bool? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get value => $_getBF(0);
  @$pb.TagNumber(1)
  set value($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasValue() => $_has(0);
  @$pb.TagNumber(1)
  void clearValue() => $_clearField(1);
}

class Payload extends $pb.GeneratedMessage {
  factory Payload({
    $fixnum.Int64? type,
    $core.Iterable<$core.String>? data,
  }) {
    final result = create();
    if (type != null) result.type = type;
    if (data != null) result.data.addAll(data);
    return result;
  }

  Payload._();

  factory Payload.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Payload.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Payload',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aInt64(1, _omitFieldNames ? '' : 'type')
    ..pPS(2, _omitFieldNames ? '' : 'data')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Payload clone() => Payload()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Payload copyWith(void Function(Payload) updates) =>
      super.copyWith((message) => updates(message as Payload)) as Payload;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Payload create() => Payload._();
  @$core.override
  Payload createEmptyInstance() => create();
  static $pb.PbList<Payload> createRepeated() => $pb.PbList<Payload>();
  @$core.pragma('dart2js:noInline')
  static Payload getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Payload>(create);
  static Payload? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get type => $_getI64(0);
  @$pb.TagNumber(1)
  set type($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get data => $_getList(1);
}

class Allowlist extends $pb.GeneratedMessage {
  factory Allowlist({
    Ports? ports,
    $core.Iterable<$core.String>? subnets,
  }) {
    final result = create();
    if (ports != null) result.ports = ports;
    if (subnets != null) result.subnets.addAll(subnets);
    return result;
  }

  Allowlist._();

  factory Allowlist.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Allowlist.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Allowlist',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOM<Ports>(1, _omitFieldNames ? '' : 'ports', subBuilder: Ports.create)
    ..pPS(2, _omitFieldNames ? '' : 'subnets')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Allowlist clone() => Allowlist()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Allowlist copyWith(void Function(Allowlist) updates) =>
      super.copyWith((message) => updates(message as Allowlist)) as Allowlist;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Allowlist create() => Allowlist._();
  @$core.override
  Allowlist createEmptyInstance() => create();
  static $pb.PbList<Allowlist> createRepeated() => $pb.PbList<Allowlist>();
  @$core.pragma('dart2js:noInline')
  static Allowlist getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Allowlist>(create);
  static Allowlist? _defaultInstance;

  @$pb.TagNumber(1)
  Ports get ports => $_getN(0);
  @$pb.TagNumber(1)
  set ports(Ports value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasPorts() => $_has(0);
  @$pb.TagNumber(1)
  void clearPorts() => $_clearField(1);
  @$pb.TagNumber(1)
  Ports ensurePorts() => $_ensure(0);

  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get subnets => $_getList(1);
}

class Ports extends $pb.GeneratedMessage {
  factory Ports({
    $core.Iterable<$fixnum.Int64>? udp,
    $core.Iterable<$fixnum.Int64>? tcp,
  }) {
    final result = create();
    if (udp != null) result.udp.addAll(udp);
    if (tcp != null) result.tcp.addAll(tcp);
    return result;
  }

  Ports._();

  factory Ports.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Ports.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Ports',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..p<$fixnum.Int64>(1, _omitFieldNames ? '' : 'udp', $pb.PbFieldType.K6)
    ..p<$fixnum.Int64>(2, _omitFieldNames ? '' : 'tcp', $pb.PbFieldType.K6)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Ports clone() => Ports()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Ports copyWith(void Function(Ports) updates) =>
      super.copyWith((message) => updates(message as Ports)) as Ports;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Ports create() => Ports._();
  @$core.override
  Ports createEmptyInstance() => create();
  static $pb.PbList<Ports> createRepeated() => $pb.PbList<Ports>();
  @$core.pragma('dart2js:noInline')
  static Ports getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Ports>(create);
  static Ports? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$fixnum.Int64> get udp => $_getList(0);

  @$pb.TagNumber(2)
  $pb.PbList<$fixnum.Int64> get tcp => $_getList(1);
}

class ServerGroup extends $pb.GeneratedMessage {
  factory ServerGroup({
    $core.String? name,
    $core.bool? virtualLocation,
  }) {
    final result = create();
    if (name != null) result.name = name;
    if (virtualLocation != null) result.virtualLocation = virtualLocation;
    return result;
  }

  ServerGroup._();

  factory ServerGroup.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ServerGroup.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ServerGroup',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOB(2, _omitFieldNames ? '' : 'virtualLocation',
        protoName: 'virtualLocation')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServerGroup clone() => ServerGroup()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServerGroup copyWith(void Function(ServerGroup) updates) =>
      super.copyWith((message) => updates(message as ServerGroup))
          as ServerGroup;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ServerGroup create() => ServerGroup._();
  @$core.override
  ServerGroup createEmptyInstance() => create();
  static $pb.PbList<ServerGroup> createRepeated() => $pb.PbList<ServerGroup>();
  @$core.pragma('dart2js:noInline')
  static ServerGroup getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ServerGroup>(create);
  static ServerGroup? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get virtualLocation => $_getBF(1);
  @$pb.TagNumber(2)
  set virtualLocation($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasVirtualLocation() => $_has(1);
  @$pb.TagNumber(2)
  void clearVirtualLocation() => $_clearField(2);
}

class ServerGroupsList extends $pb.GeneratedMessage {
  factory ServerGroupsList({
    $fixnum.Int64? type,
    $core.Iterable<ServerGroup>? servers,
  }) {
    final result = create();
    if (type != null) result.type = type;
    if (servers != null) result.servers.addAll(servers);
    return result;
  }

  ServerGroupsList._();

  factory ServerGroupsList.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ServerGroupsList.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ServerGroupsList',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aInt64(1, _omitFieldNames ? '' : 'type')
    ..pc<ServerGroup>(2, _omitFieldNames ? '' : 'servers', $pb.PbFieldType.PM,
        subBuilder: ServerGroup.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServerGroupsList clone() => ServerGroupsList()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServerGroupsList copyWith(void Function(ServerGroupsList) updates) =>
      super.copyWith((message) => updates(message as ServerGroupsList))
          as ServerGroupsList;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ServerGroupsList create() => ServerGroupsList._();
  @$core.override
  ServerGroupsList createEmptyInstance() => create();
  static $pb.PbList<ServerGroupsList> createRepeated() =>
      $pb.PbList<ServerGroupsList>();
  @$core.pragma('dart2js:noInline')
  static ServerGroupsList getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ServerGroupsList>(create);
  static ServerGroupsList? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get type => $_getI64(0);
  @$pb.TagNumber(1)
  set type($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<ServerGroup> get servers => $_getList(1);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
