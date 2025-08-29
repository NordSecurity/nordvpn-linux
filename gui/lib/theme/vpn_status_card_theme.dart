import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'vpn_status_card_theme.tailor.dart';

// Theme data for VPN status card
@tailorMixin
final class VpnStatusCardTheme extends ThemeExtension<VpnStatusCardTheme>
    with _$VpnStatusCardThemeTailorMixin {
  @override
  final double height;

  @override
  final double maxConnectButtonWidth;

  @override
  final TextStyle primaryFont;

  @override
  final TextStyle secondaryFont;

  @override
  final double iconSize;

  VpnStatusCardTheme({
    required this.height,
    required this.primaryFont,
    required this.secondaryFont,
    required this.iconSize,
    required this.maxConnectButtonWidth,
  });
}
