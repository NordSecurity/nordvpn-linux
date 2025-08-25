// This is a generated file - do not edit.
//
// Generated from countries.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use countriesResponseDescriptor instead')
const CountriesResponse$json = {
  '1': 'CountriesResponse',
  '2': [
    {
      '1': 'countries',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.pb.Country',
      '10': 'countries'
    },
  ],
};

/// Descriptor for `CountriesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List countriesResponseDescriptor = $convert.base64Decode(
    'ChFDb3VudHJpZXNSZXNwb25zZRIpCgljb3VudHJpZXMYASADKAsyCy5wYi5Db3VudHJ5Ugljb3'
    'VudHJpZXM=');

@$core.Deprecated('Use countryDescriptor instead')
const Country$json = {
  '1': 'Country',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '10': 'code'},
  ],
};

/// Descriptor for `Country`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List countryDescriptor = $convert.base64Decode(
    'CgdDb3VudHJ5EhIKBG5hbWUYASABKAlSBG5hbWUSEgoEY29kZRgCIAEoCVIEY29kZQ==');
