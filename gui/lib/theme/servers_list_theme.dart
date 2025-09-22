import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'servers_list_theme.tailor.dart';

@tailorMixin
final class ServersListTheme extends ThemeExtension<ServersListTheme>
    with _$ServersListThemeTailorMixin {
  @override
  final double flagSize;

  @override
  final double loaderSize;

  @override
  final double listItemHeight;

  @override
  final EdgeInsetsGeometry paddingSearchGroupsLabel;

  @override
  final TextStyle searchHintStyle;

  @override
  final TextStyle obfuscationSearchWarningStyle;

  @override
  final TextStyle searchErrorStyle;

  @override
  final Color obfuscatedItemBackgroundColor;

  ServersListTheme({
    required this.flagSize,
    required this.loaderSize,
    required this.listItemHeight,
    required this.paddingSearchGroupsLabel,
    required this.searchHintStyle,
    required this.searchErrorStyle,
    required this.obfuscationSearchWarningStyle,
    required this.obfuscatedItemBackgroundColor,
  });
}
