// This is a generated file - do not edit.
//
// Generated from rate.proto.

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

class RateRequest extends $pb.GeneratedMessage {
  factory RateRequest({
    $fixnum.Int64? rating,
  }) {
    final result = create();
    if (rating != null) result.rating = rating;
    return result;
  }

  RateRequest._();

  factory RateRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RateRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RateRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pb'),
      createEmptyInstance: create)
    ..aInt64(2, _omitFieldNames ? '' : 'rating')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RateRequest clone() => RateRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RateRequest copyWith(void Function(RateRequest) updates) =>
      super.copyWith((message) => updates(message as RateRequest))
          as RateRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RateRequest create() => RateRequest._();
  @$core.override
  RateRequest createEmptyInstance() => create();
  static $pb.PbList<RateRequest> createRepeated() => $pb.PbList<RateRequest>();
  @$core.pragma('dart2js:noInline')
  static RateRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RateRequest>(create);
  static RateRequest? _defaultInstance;

  @$pb.TagNumber(2)
  $fixnum.Int64 get rating => $_getI64(0);
  @$pb.TagNumber(2)
  set rating($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(2)
  $core.bool hasRating() => $_has(0);
  @$pb.TagNumber(2)
  void clearRating() => $_clearField(2);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
