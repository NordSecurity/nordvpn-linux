import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'interactive_list_view_theme.tailor.dart';

@tailorMixin
final class InteractiveListViewTheme
    extends ThemeExtension<InteractiveListViewTheme>
    with _$InteractiveListViewThemeTailorMixin {
  @override
  final double borderRadius;

  @override
  final Color borderColor;

  @override
  final Color focusBorderColor;

  @override
  final double borderWidth;

  InteractiveListViewTheme({
    required this.borderRadius,
    required this.borderColor,
    required this.focusBorderColor,
    required this.borderWidth,
  });
}
