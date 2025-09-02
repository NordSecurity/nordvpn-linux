import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'error_screen_theme.tailor.dart';

@tailorMixin
final class ErrorScreenTheme extends ThemeExtension<ErrorScreenTheme>
    with _$ErrorScreenThemeTailorMixin {
  @override
  final TextStyle titleTextStyle;

  @override
  final TextStyle descriptionTextStyle;

  ErrorScreenTheme({
    required this.titleTextStyle,
    required this.descriptionTextStyle,
  });
}
