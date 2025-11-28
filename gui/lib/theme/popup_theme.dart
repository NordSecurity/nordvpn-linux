import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'popup_theme.tailor.dart';

@tailorMixin
final class PopupTheme extends ThemeExtension<PopupTheme>
    with _$PopupThemeTailorMixin {
  // Widget dimensions
  @override
  final double widgetWidth;

  @override
  final BorderRadius widgetRadius;

  // Layout spacing
  @override
  final double contentAllPadding;

  @override
  final double xButtonAllPadding;

  @override
  final double gapBetweenElements;

  @override
  final double verticalElementSpacing;

  // Button dimensions
  @override
  final double singleButtonMinWidth;

  // Text styles
  @override
  final TextStyle textPrimary;

  @override
  final TextStyle textSecondary;

  PopupTheme({
    // Widget dimensions
    required this.widgetWidth,
    required this.widgetRadius,
    // Layout spacing
    required this.contentAllPadding,
    required this.xButtonAllPadding,
    required this.gapBetweenElements,
    required this.verticalElementSpacing,
    // Button dimensions
    required this.singleButtonMinWidth,
    // Text styles
    required this.textPrimary,
    required this.textSecondary,
  });
}
