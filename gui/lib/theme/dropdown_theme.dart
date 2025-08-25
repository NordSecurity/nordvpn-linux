// ignore_for_file: annotate_overrides

import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'dropdown_theme.tailor.dart';

@tailorMixin
final class DropdownTheme extends ThemeExtension<DropdownTheme>
    with _$DropdownThemeTailorMixin {
  final Color color;
  final double borderRadius;
  final Color borderColor;
  final Color focusBorderColor;
  final Color errorBorderColor;
  final double borderWidth;
  final double horizontalPadding;

  DropdownTheme({
    required this.color,
    required this.borderRadius,
    required this.borderColor,
    required this.focusBorderColor,
    required this.errorBorderColor,
    required this.borderWidth,
    required this.horizontalPadding,
  });
}
