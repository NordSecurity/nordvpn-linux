import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'consent_screen_theme.tailor.dart';

@tailorMixin
final class ConsentScreenTheme extends ThemeExtension<ConsentScreenTheme>
    with _$ConsentScreenThemeTailorMixin {
  @override
  final Color overlayColor;

  @override
  final TextStyle bodyTextStyle;

  @override
  final TextStyle titleTextStyle;

  @override
  final TextStyle titleBarTextStyle;

  @override
  final double width;

  @override
  final double height;

  @override
  final double padding;

  @override
  final TextStyle listItemTitle;

  @override
  final TextStyle listItemSubtitle;

  @override
  final double titleBarWidth;

  ConsentScreenTheme({
    required this.overlayColor,
    required this.bodyTextStyle,
    required this.titleTextStyle,
    required this.titleBarTextStyle,
    required this.width,
    required this.height,
    required this.padding,
    required this.listItemTitle,
    required this.listItemSubtitle,
    required this.titleBarWidth,
  });
}
