// This is a generated file - do not edit.
//
// Generated from transfer.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'google/protobuf/timestamp.pb.dart' as $0;
import 'transfer.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'transfer.pbenum.dart';

class Transfer extends $pb.GeneratedMessage {
  factory Transfer({
    $core.String? id,
    Direction? direction,
    $core.String? peer,
    Status? status,
    $0.Timestamp? created,
    $core.Iterable<File>? files,
    $core.String? path,
    $fixnum.Int64? totalSize,
    $fixnum.Int64? totalTransferred,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (direction != null) result.direction = direction;
    if (peer != null) result.peer = peer;
    if (status != null) result.status = status;
    if (created != null) result.created = created;
    if (files != null) result.files.addAll(files);
    if (path != null) result.path = path;
    if (totalSize != null) result.totalSize = totalSize;
    if (totalTransferred != null) result.totalTransferred = totalTransferred;
    return result;
  }

  Transfer._();

  factory Transfer.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Transfer.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Transfer',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'filesharepb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..e<Direction>(2, _omitFieldNames ? '' : 'direction', $pb.PbFieldType.OE,
        defaultOrMaker: Direction.UNKNOWN_DIRECTION,
        valueOf: Direction.valueOf,
        enumValues: Direction.values)
    ..aOS(3, _omitFieldNames ? '' : 'peer')
    ..e<Status>(4, _omitFieldNames ? '' : 'status', $pb.PbFieldType.OE,
        defaultOrMaker: Status.SUCCESS,
        valueOf: Status.valueOf,
        enumValues: Status.values)
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'created',
        subBuilder: $0.Timestamp.create)
    ..pc<File>(6, _omitFieldNames ? '' : 'files', $pb.PbFieldType.PM,
        subBuilder: File.create)
    ..aOS(7, _omitFieldNames ? '' : 'path')
    ..a<$fixnum.Int64>(
        8, _omitFieldNames ? '' : 'totalSize', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(
        9, _omitFieldNames ? '' : 'totalTransferred', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Transfer clone() => Transfer()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Transfer copyWith(void Function(Transfer) updates) =>
      super.copyWith((message) => updates(message as Transfer)) as Transfer;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Transfer create() => Transfer._();
  @$core.override
  Transfer createEmptyInstance() => create();
  static $pb.PbList<Transfer> createRepeated() => $pb.PbList<Transfer>();
  @$core.pragma('dart2js:noInline')
  static Transfer getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Transfer>(create);
  static Transfer? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  Direction get direction => $_getN(1);
  @$pb.TagNumber(2)
  set direction(Direction value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasDirection() => $_has(1);
  @$pb.TagNumber(2)
  void clearDirection() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get peer => $_getSZ(2);
  @$pb.TagNumber(3)
  set peer($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPeer() => $_has(2);
  @$pb.TagNumber(3)
  void clearPeer() => $_clearField(3);

  @$pb.TagNumber(4)
  Status get status => $_getN(3);
  @$pb.TagNumber(4)
  set status(Status value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasStatus() => $_has(3);
  @$pb.TagNumber(4)
  void clearStatus() => $_clearField(4);

  @$pb.TagNumber(5)
  $0.Timestamp get created => $_getN(4);
  @$pb.TagNumber(5)
  set created($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasCreated() => $_has(4);
  @$pb.TagNumber(5)
  void clearCreated() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureCreated() => $_ensure(4);

  @$pb.TagNumber(6)
  $pb.PbList<File> get files => $_getList(5);

  /// For outgoing transfers the user provided path to be sent
  /// For incoming transfers path where the files will be downloaded to
  @$pb.TagNumber(7)
  $core.String get path => $_getSZ(6);
  @$pb.TagNumber(7)
  set path($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasPath() => $_has(6);
  @$pb.TagNumber(7)
  void clearPath() => $_clearField(7);

  @$pb.TagNumber(8)
  $fixnum.Int64 get totalSize => $_getI64(7);
  @$pb.TagNumber(8)
  set totalSize($fixnum.Int64 value) => $_setInt64(7, value);
  @$pb.TagNumber(8)
  $core.bool hasTotalSize() => $_has(7);
  @$pb.TagNumber(8)
  void clearTotalSize() => $_clearField(8);

  @$pb.TagNumber(9)
  $fixnum.Int64 get totalTransferred => $_getI64(8);
  @$pb.TagNumber(9)
  set totalTransferred($fixnum.Int64 value) => $_setInt64(8, value);
  @$pb.TagNumber(9)
  $core.bool hasTotalTransferred() => $_has(8);
  @$pb.TagNumber(9)
  void clearTotalTransferred() => $_clearField(9);
}

class File extends $pb.GeneratedMessage {
  factory File({
    $core.String? id,
    $fixnum.Int64? size,
    $fixnum.Int64? transferred,
    Status? status,
    $core.Iterable<$core.MapEntry<$core.String, File>>? children,
    $core.String? path,
    $core.String? fullPath,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (size != null) result.size = size;
    if (transferred != null) result.transferred = transferred;
    if (status != null) result.status = status;
    if (children != null) result.children.addEntries(children);
    if (path != null) result.path = path;
    if (fullPath != null) result.fullPath = fullPath;
    return result;
  }

  File._();

  factory File.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory File.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'File',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'filesharepb'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..a<$fixnum.Int64>(2, _omitFieldNames ? '' : 'size', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(
        3, _omitFieldNames ? '' : 'transferred', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..e<Status>(4, _omitFieldNames ? '' : 'status', $pb.PbFieldType.OE,
        defaultOrMaker: Status.SUCCESS,
        valueOf: Status.valueOf,
        enumValues: Status.values)
    ..m<$core.String, File>(5, _omitFieldNames ? '' : 'children',
        entryClassName: 'File.ChildrenEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OM,
        valueCreator: File.create,
        valueDefaultOrMaker: File.getDefault,
        packageName: const $pb.PackageName('filesharepb'))
    ..aOS(6, _omitFieldNames ? '' : 'path')
    ..aOS(7, _omitFieldNames ? '' : 'fullPath', protoName: 'fullPath')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  File clone() => File()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  File copyWith(void Function(File) updates) =>
      super.copyWith((message) => updates(message as File)) as File;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static File create() => File._();
  @$core.override
  File createEmptyInstance() => create();
  static $pb.PbList<File> createRepeated() => $pb.PbList<File>();
  @$core.pragma('dart2js:noInline')
  static File getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<File>(create);
  static File? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get size => $_getI64(1);
  @$pb.TagNumber(2)
  set size($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSize() => $_has(1);
  @$pb.TagNumber(2)
  void clearSize() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get transferred => $_getI64(2);
  @$pb.TagNumber(3)
  set transferred($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTransferred() => $_has(2);
  @$pb.TagNumber(3)
  void clearTransferred() => $_clearField(3);

  @$pb.TagNumber(4)
  Status get status => $_getN(3);
  @$pb.TagNumber(4)
  set status(Status value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasStatus() => $_has(3);
  @$pb.TagNumber(4)
  void clearStatus() => $_clearField(4);

  /// Not used anymore, file lists should always be flat, kept for history file compatibility
  @$pb.TagNumber(5)
  $pb.PbMap<$core.String, File> get children => $_getMap(4);

  @$pb.TagNumber(6)
  $core.String get path => $_getSZ(5);
  @$pb.TagNumber(6)
  set path($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPath() => $_has(5);
  @$pb.TagNumber(6)
  void clearPath() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get fullPath => $_getSZ(6);
  @$pb.TagNumber(7)
  set fullPath($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasFullPath() => $_has(6);
  @$pb.TagNumber(7)
  void clearFullPath() => $_clearField(7);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
