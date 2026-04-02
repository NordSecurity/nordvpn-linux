import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'connection_card_theme.tailor.dart';

@tailorMixinComponent
final class ConnectionCardLabelTheme
    extends ThemeExtension<ConnectionCardLabelTheme>
    with _$ConnectionCardLabelThemeTailorMixin {
  @override
  final Color disconnectedColor;

  @override
  final Color connectingColor;

  @override
  final Color connectedColor;

  @override
  final Color serverTypeColor;

  @override
  final double spacing;

  @override
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

@tailorMixinComponent
final class ConnectionCardIconTheme
    extends ThemeExtension<ConnectionCardIconTheme>
    with _$ConnectionCardIconThemeTailorMixin {
  @override
  final double iconSize;

  @override
  final double flagBorderSize;

  @override
  final double dipIconWidth;

  @override
  final double dipIconHeight;

  @override
  final Color borderConnectedColor;

  @override
  final Color borderConnectingColor;

  @override
  final Color disconnectedBackgroundColor;

  @override
  final String disconnectedIcon;

  @override
  final EdgeInsetsGeometry disconnectedPadding;

  ConnectionCardIconTheme({
    required this.iconSize,
    required this.flagBorderSize,
    required this.dipIconWidth,
    required this.dipIconHeight,
    required this.borderConnectedColor,
    required this.borderConnectingColor,
    required this.disconnectedBackgroundColor,
    required this.disconnectedIcon,
    required this.disconnectedPadding,
  });
}

@tailorMixinComponent
final class ConnectionCardButtonTheme
    extends ThemeExtension<ConnectionCardButtonTheme>
    with _$ConnectionCardButtonThemeTailorMixin {
  @override
  final double maxConnectButtonWidth;

  @override
  final ButtonStyle secureMyConnectionButtonStyle;

  @override
  final ButtonStyle cancelButtonStyle;

  @override
  final ButtonStyle connectionDetailsButtonStyle;

  ConnectionCardButtonTheme({
    required this.maxConnectButtonWidth,
    required this.secureMyConnectionButtonStyle,
    required this.cancelButtonStyle,
    required this.connectionDetailsButtonStyle,
  });
}

// Theme data for VPN status card
@tailorMixin
final class ConnectionCardTheme extends ThemeExtension<ConnectionCardTheme>
    with _$ConnectionCardThemeTailorMixin {
  @override
  final TextStyle primaryFont;

  @override
  final EdgeInsetsGeometry mapPadding;

  @override
  final EdgeInsetsGeometry connectionCardPadding;

  @override
  final EdgeInsetsGeometry margin;

  @override
  final BorderRadius borderRadius;

  @override
  final double minWidth;

  @override
  final double smallSpacing;

  @override
  final double mediumSpacing;

  @override
  final ConnectionCardLabelTheme labelTheme;

  @override
  final ConnectionCardIconTheme iconTheme;

  @override
  final ConnectionCardButtonTheme buttonTheme;

  ConnectionCardTheme({
    required this.primaryFont,
    required this.mapPadding,
    required this.connectionCardPadding,
    required this.margin,
    required this.borderRadius,
    required this.minWidth,
    required this.smallSpacing,
    required this.mediumSpacing,
    required this.labelTheme,
    required this.iconTheme,
    required this.buttonTheme,
  });
}
