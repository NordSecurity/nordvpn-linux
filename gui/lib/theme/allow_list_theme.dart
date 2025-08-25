import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'allow_list_theme.tailor.dart';

@tailorMixin
final class AllowListTheme extends ThemeExtension<AllowListTheme>
    with _$AllowListThemeTailorMixin {
  @override
  final TextStyle labelStyle;
  @override
  final Color addCardBackground;
  @override
  final Color dividerColor;
  @override
  final Color listItemBackgroundColor;
  @override
  final TextStyle tableHeaderStyle;
  @override
  final TextStyle tableItemsStyle;

  AllowListTheme({
    required this.labelStyle,
    required this.addCardBackground,
    required this.dividerColor,
    required this.listItemBackgroundColor,
    required this.tableHeaderStyle,
    required this.tableItemsStyle,
  });
}
