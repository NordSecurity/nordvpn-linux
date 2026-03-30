import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'connection_card_theme.tailor.dart';

final class ConnectionCardLabelTheme {
  final Color disconnectedColor;
  final Color connectingColor;
  final Color connectedColor;
  final Color serverTypeColor;
  final double spacing;
  final TextStyle font;

  ConnectionCardLabelTheme({
    required this.disconnectedColor,
    required this.connectingColor,
    required this.connectedColor,
    required this.serverTypeColor,
    required this.spacing,
    required this.font,
  });
}

final class ConnectionCardIconTheme {
  final double iconSize;
  final double flagBorderSize;
  final double strokeWidth;
  final double dipIconWidth;
  final double dipIconHeight;
  final Color borderConnectedColor;
  final Color borderConnectingColor;
  final Color disconnectedBackgroundColor;
  final String disconnectedIcon;

  ConnectionCardIconTheme({
    required this.iconSize,
    required this.flagBorderSize,
    required this.strokeWidth,
    required this.dipIconWidth,
    required this.dipIconHeight,
    required this.borderConnectedColor,
    required this.borderConnectingColor,
    required this.disconnectedBackgroundColor,
    required this.disconnectedIcon,
  });
}

// Theme data for VPN status card
@tailorMixin
final class ConnectionCardTheme extends ThemeExtension<ConnectionCardTheme>
    with _$ConnectionCardThemeTailorMixin {
  @override
  final double height;

  @override
  final double maxConnectButtonWidth;

  @override
  final TextStyle primaryFont;

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
  final ConnectionCardLabelTheme labelTheme;

  @override
  final ConnectionCardIconTheme iconTheme;

  ConnectionCardTheme({
    required this.height,
    required this.primaryFont,
    required this.maxConnectButtonWidth,
    required this.secureMyConnectionButtonStyle,
    required this.cancelButtonStyle,
    required this.connectionCardPadding,
    required this.smallSpacing,
    required this.mediumSpacing,
    required this.labelTheme,
    required this.iconTheme,
  });
}
