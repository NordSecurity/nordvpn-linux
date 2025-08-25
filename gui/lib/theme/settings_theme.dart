import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'settings_theme.tailor.dart';

// Define the theme style for Settings screens
@tailorMixin
final class SettingsTheme extends ThemeExtension<SettingsTheme>
    with _$SettingsThemeTailorMixin {
  @override
  final TextStyle currentPageNameStyle;
  @override
  final TextStyle parentPageStyle;
  @override
  final TextStyle itemTitleStyle;
  @override
  final TextStyle itemSubtitleStyle;
  @override
  final TextStyle vpnStatusStyle;
  @override
  final double textInputWidth;
  @override
  final TextStyle otherProductsTitle;
  @override
  final TextStyle otherProductsSubtitle;
  @override
  final double fwMarkInputSize;
  // default padding for settings items
  @override
  final EdgeInsets itemPadding;

  SettingsTheme({
    required this.itemTitleStyle,
    required this.itemSubtitleStyle,
    required this.currentPageNameStyle,
    required this.parentPageStyle,
    required this.vpnStatusStyle,
    required this.textInputWidth,
    required this.otherProductsTitle,
    required this.otherProductsSubtitle,
    required this.fwMarkInputSize,
    required this.itemPadding,
  });
}
