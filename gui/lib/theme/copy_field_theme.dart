import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'copy_field_theme.tailor.dart';

@tailorMixin
final class CopyFieldTheme extends ThemeExtension<CopyFieldTheme>
    with _$CopyFieldThemeTailorMixin {
  @override
  final double borderRadius;

  @override
  final TextStyle commandTextStyle;

  @override
  final TextStyle descriptionTextStyle;

  CopyFieldTheme({
    required this.borderRadius,
    required this.commandTextStyle,
    required this.descriptionTextStyle,
  });
}
