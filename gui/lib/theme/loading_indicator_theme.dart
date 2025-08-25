import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'loading_indicator_theme.tailor.dart';

@tailorMixin
final class LoadingIndicatorTheme extends ThemeExtension<LoadingIndicatorTheme>
    with _$LoadingIndicatorThemeTailorMixin {
  @override
  final Color color;

  @override
  final double strokeWidth;

  LoadingIndicatorTheme({required this.color, required this.strokeWidth});
}
