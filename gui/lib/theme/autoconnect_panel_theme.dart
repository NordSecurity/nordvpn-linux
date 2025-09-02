import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'autoconnect_panel_theme.tailor.dart';

@tailorMixin
final class AutoconnectPanelTheme extends ThemeExtension<AutoconnectPanelTheme>
    with _$AutoconnectPanelThemeTailorMixin {
  @override
  final TextStyle primaryFont;

  @override
  final TextStyle secondaryFont;

  @override
  final double iconSize;

  @override
  final double loaderSize;

  AutoconnectPanelTheme({
    required this.primaryFont,
    required this.secondaryFont,
    required this.iconSize,
    required this.loaderSize,
  });
}
