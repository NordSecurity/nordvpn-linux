import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'vpn_status_card_theme.tailor.dart';

final class ConnectionCardLabelThemeStyle {
  final Color disconnectedColor;
  final Color connectingColor;
  final Color connectedColor;

  ConnectionCardLabelThemeStyle({
    required this.disconnectedColor,
    required this.connectingColor,
    required this.connectedColor,
  });
}

final class ConnectionCardIconThemeStyle {
  final double iconSize;
  final double flagBorderSize;
  final double strokeWidth;
  final Color borderConnectedColor;
  final Color borderConnectingColor;
  final Color disconnectedBackgroundColor;
  final String disconnectedIcon;

  ConnectionCardIconThemeStyle({
    required this.iconSize,
    required this.flagBorderSize,
    required this.strokeWidth,
    required this.borderConnectedColor,
    required this.borderConnectingColor,
    required this.disconnectedBackgroundColor,
    required this.disconnectedIcon,
  });
}

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
  final ButtonStyle secureMyConnectionButtonStyle;

  @override
  final ButtonStyle cancelButtonStyle;

  @override
  final EdgeInsetsGeometry connectionCardPadding;

  @override
  final double smallSpacing;

  @override
  final double mediumSpacing;

  @override
  final ConnectionCardLabelThemeStyle labelStyle;

  @override
  final ConnectionCardIconThemeStyle iconStyle;

  VpnStatusCardTheme({
    required this.height,
    required this.primaryFont,
    required this.secondaryFont,
    required this.maxConnectButtonWidth,
    required this.secureMyConnectionButtonStyle,
    required this.cancelButtonStyle,
    required this.connectionCardPadding,
    required this.smallSpacing,
    required this.mediumSpacing,
    required this.labelStyle,
    required this.iconStyle,
  });
}
