import 'package:flutter/material.dart';
import 'package:freezed_annotation/freezed_annotation.dart';
import 'package:nordvpn/router/routes.dart';

part 'metadata.freezed.dart';

@freezed
abstract class RouteMetadata with _$RouteMetadata {
  const RouteMetadata._();

  const factory RouteMetadata({
    required AppRoute route,
    required Widget screen,
    String? displayName,
    Function(BuildContext)? onPressed,
  }) = _RouteMetadata;
}
