import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'inline_loading_indicator_theme.tailor.dart';

@tailorMixin
final class InlineLoadingIndicatorTheme
    extends ThemeExtension<InlineLoadingIndicatorTheme>
    with _$InlineLoadingIndicatorThemeTailorMixin {
  @override
  final double width;

  @override
  final double height;

  @override
  final double stroke;

  @override
  final Color color;

  @override
  final Color alternativeColor;

  InlineLoadingIndicatorTheme({
    required this.width,
    required this.height,
    required this.stroke,
    required this.color,
    required this.alternativeColor,
  });
}
