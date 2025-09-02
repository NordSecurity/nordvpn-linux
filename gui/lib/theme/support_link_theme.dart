import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'support_link_theme.tailor.dart';

@tailorMixin
final class SupportLinkTheme extends ThemeExtension<SupportLinkTheme>
    with _$SupportLinkThemeTailorMixin {
  @override
  final TextStyle textStyle;

  @override
  final Color urlColor;

  SupportLinkTheme({required this.textStyle, required this.urlColor});
}
