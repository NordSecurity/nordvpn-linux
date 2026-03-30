// ignore_for_file: annotate_overrides

import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'context_menu_theme.tailor.dart';

@tailorMixin
final class ContextMenuTheme extends ThemeExtension<ContextMenuTheme>
    with _$ContextMenuThemeTailorMixin {
  final double menuWidth;
  final BorderRadius menuRadius;
  final EdgeInsets menuPadding;
  final Color menuColor;
  final Color menuBorderColor;
  final double menuBorderWidth;
  final double itemHeight;
  final EdgeInsets itemPadding;
  final Color itemHoverColor;
  final TextStyle itemTextStyle;

  ContextMenuTheme({
    required this.menuWidth,
    required this.menuRadius,
    required this.menuPadding,
    required this.menuColor,
    required this.menuBorderColor,
    required this.menuBorderWidth,
    required this.itemHeight,
    required this.itemPadding,
    required this.itemHoverColor,
    required this.itemTextStyle,
  });
}
