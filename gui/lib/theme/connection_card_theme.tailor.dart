// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'connection_card_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$ConnectionCardLabelThemeTailorMixin
    on ThemeExtension<ConnectionCardLabelTheme> {
  Color get disconnectedColor;
  Color get connectingColor;
  Color get connectedColor;
  Color get serverTypeColor;
  double get spacing;
  TextStyle get font;

  @override
  ConnectionCardLabelTheme copyWith({
    Color? disconnectedColor,
    Color? connectingColor,
    Color? connectedColor,
    Color? serverTypeColor,
    double? spacing,
    TextStyle? font,
  }) {
    return ConnectionCardLabelTheme(
      disconnectedColor: disconnectedColor ?? this.disconnectedColor,
      connectingColor: connectingColor ?? this.connectingColor,
      connectedColor: connectedColor ?? this.connectedColor,
      serverTypeColor: serverTypeColor ?? this.serverTypeColor,
      spacing: spacing ?? this.spacing,
      font: font ?? this.font,
    );
  }

  @override
  ConnectionCardLabelTheme lerp(
    covariant ThemeExtension<ConnectionCardLabelTheme>? other,
    double t,
  ) {
    if (other is! ConnectionCardLabelTheme)
      return this as ConnectionCardLabelTheme;
    return ConnectionCardLabelTheme(
      disconnectedColor: Color.lerp(
        disconnectedColor,
        other.disconnectedColor,
        t,
      )!,
      connectingColor: Color.lerp(connectingColor, other.connectingColor, t)!,
      connectedColor: Color.lerp(connectedColor, other.connectedColor, t)!,
      serverTypeColor: Color.lerp(serverTypeColor, other.serverTypeColor, t)!,
      spacing: t < 0.5 ? spacing : other.spacing,
      font: TextStyle.lerp(font, other.font, t)!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is ConnectionCardLabelTheme &&
            const DeepCollectionEquality().equals(
              disconnectedColor,
              other.disconnectedColor,
            ) &&
            const DeepCollectionEquality().equals(
              connectingColor,
              other.connectingColor,
            ) &&
            const DeepCollectionEquality().equals(
              connectedColor,
              other.connectedColor,
            ) &&
            const DeepCollectionEquality().equals(
              serverTypeColor,
              other.serverTypeColor,
            ) &&
            const DeepCollectionEquality().equals(spacing, other.spacing) &&
            const DeepCollectionEquality().equals(font, other.font));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(disconnectedColor),
      const DeepCollectionEquality().hash(connectingColor),
      const DeepCollectionEquality().hash(connectedColor),
      const DeepCollectionEquality().hash(serverTypeColor),
      const DeepCollectionEquality().hash(spacing),
      const DeepCollectionEquality().hash(font),
    );
  }
}

mixin _$ConnectionCardIconThemeTailorMixin
    on ThemeExtension<ConnectionCardIconTheme> {
  double get iconSize;
  double get flagBorderSize;
  double get dipIconWidth;
  double get dipIconHeight;
  Color get borderConnectedColor;
  Color get borderConnectingColor;
  Color get disconnectedBackgroundColor;
  String get disconnectedIcon;
  EdgeInsetsGeometry get disconnectedPadding;

  @override
  ConnectionCardIconTheme copyWith({
    double? iconSize,
    double? flagBorderSize,
    double? dipIconWidth,
    double? dipIconHeight,
    Color? borderConnectedColor,
    Color? borderConnectingColor,
    Color? disconnectedBackgroundColor,
    String? disconnectedIcon,
    EdgeInsetsGeometry? disconnectedPadding,
  }) {
    return ConnectionCardIconTheme(
      iconSize: iconSize ?? this.iconSize,
      flagBorderSize: flagBorderSize ?? this.flagBorderSize,
      dipIconWidth: dipIconWidth ?? this.dipIconWidth,
      dipIconHeight: dipIconHeight ?? this.dipIconHeight,
      borderConnectedColor: borderConnectedColor ?? this.borderConnectedColor,
      borderConnectingColor:
          borderConnectingColor ?? this.borderConnectingColor,
      disconnectedBackgroundColor:
          disconnectedBackgroundColor ?? this.disconnectedBackgroundColor,
      disconnectedIcon: disconnectedIcon ?? this.disconnectedIcon,
      disconnectedPadding: disconnectedPadding ?? this.disconnectedPadding,
    );
  }

  @override
  ConnectionCardIconTheme lerp(
    covariant ThemeExtension<ConnectionCardIconTheme>? other,
    double t,
  ) {
    if (other is! ConnectionCardIconTheme)
      return this as ConnectionCardIconTheme;
    return ConnectionCardIconTheme(
      iconSize: t < 0.5 ? iconSize : other.iconSize,
      flagBorderSize: t < 0.5 ? flagBorderSize : other.flagBorderSize,
      dipIconWidth: t < 0.5 ? dipIconWidth : other.dipIconWidth,
      dipIconHeight: t < 0.5 ? dipIconHeight : other.dipIconHeight,
      borderConnectedColor: Color.lerp(
        borderConnectedColor,
        other.borderConnectedColor,
        t,
      )!,
      borderConnectingColor: Color.lerp(
        borderConnectingColor,
        other.borderConnectingColor,
        t,
      )!,
      disconnectedBackgroundColor: Color.lerp(
        disconnectedBackgroundColor,
        other.disconnectedBackgroundColor,
        t,
      )!,
      disconnectedIcon: t < 0.5 ? disconnectedIcon : other.disconnectedIcon,
      disconnectedPadding: t < 0.5
          ? disconnectedPadding
          : other.disconnectedPadding,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is ConnectionCardIconTheme &&
            const DeepCollectionEquality().equals(iconSize, other.iconSize) &&
            const DeepCollectionEquality().equals(
              flagBorderSize,
              other.flagBorderSize,
            ) &&
            const DeepCollectionEquality().equals(
              dipIconWidth,
              other.dipIconWidth,
            ) &&
            const DeepCollectionEquality().equals(
              dipIconHeight,
              other.dipIconHeight,
            ) &&
            const DeepCollectionEquality().equals(
              borderConnectedColor,
              other.borderConnectedColor,
            ) &&
            const DeepCollectionEquality().equals(
              borderConnectingColor,
              other.borderConnectingColor,
            ) &&
            const DeepCollectionEquality().equals(
              disconnectedBackgroundColor,
              other.disconnectedBackgroundColor,
            ) &&
            const DeepCollectionEquality().equals(
              disconnectedIcon,
              other.disconnectedIcon,
            ) &&
            const DeepCollectionEquality().equals(
              disconnectedPadding,
              other.disconnectedPadding,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(iconSize),
      const DeepCollectionEquality().hash(flagBorderSize),
      const DeepCollectionEquality().hash(dipIconWidth),
      const DeepCollectionEquality().hash(dipIconHeight),
      const DeepCollectionEquality().hash(borderConnectedColor),
      const DeepCollectionEquality().hash(borderConnectingColor),
      const DeepCollectionEquality().hash(disconnectedBackgroundColor),
      const DeepCollectionEquality().hash(disconnectedIcon),
      const DeepCollectionEquality().hash(disconnectedPadding),
    );
  }
}

mixin _$ConnectionCardButtonThemeTailorMixin
    on ThemeExtension<ConnectionCardButtonTheme> {
  double get maxConnectButtonWidth;
  ButtonStyle get secureMyConnectionButtonStyle;
  ButtonStyle get cancelButtonStyle;
  ButtonStyle get connectionDetailsButtonStyle;

  @override
  ConnectionCardButtonTheme copyWith({
    double? maxConnectButtonWidth,
    ButtonStyle? secureMyConnectionButtonStyle,
    ButtonStyle? cancelButtonStyle,
    ButtonStyle? connectionDetailsButtonStyle,
  }) {
    return ConnectionCardButtonTheme(
      maxConnectButtonWidth:
          maxConnectButtonWidth ?? this.maxConnectButtonWidth,
      secureMyConnectionButtonStyle:
          secureMyConnectionButtonStyle ?? this.secureMyConnectionButtonStyle,
      cancelButtonStyle: cancelButtonStyle ?? this.cancelButtonStyle,
      connectionDetailsButtonStyle:
          connectionDetailsButtonStyle ?? this.connectionDetailsButtonStyle,
    );
  }

  @override
  ConnectionCardButtonTheme lerp(
    covariant ThemeExtension<ConnectionCardButtonTheme>? other,
    double t,
  ) {
    if (other is! ConnectionCardButtonTheme)
      return this as ConnectionCardButtonTheme;
    return ConnectionCardButtonTheme(
      maxConnectButtonWidth: t < 0.5
          ? maxConnectButtonWidth
          : other.maxConnectButtonWidth,
      secureMyConnectionButtonStyle: t < 0.5
          ? secureMyConnectionButtonStyle
          : other.secureMyConnectionButtonStyle,
      cancelButtonStyle: t < 0.5 ? cancelButtonStyle : other.cancelButtonStyle,
      connectionDetailsButtonStyle: t < 0.5
          ? connectionDetailsButtonStyle
          : other.connectionDetailsButtonStyle,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is ConnectionCardButtonTheme &&
            const DeepCollectionEquality().equals(
              maxConnectButtonWidth,
              other.maxConnectButtonWidth,
            ) &&
            const DeepCollectionEquality().equals(
              secureMyConnectionButtonStyle,
              other.secureMyConnectionButtonStyle,
            ) &&
            const DeepCollectionEquality().equals(
              cancelButtonStyle,
              other.cancelButtonStyle,
            ) &&
            const DeepCollectionEquality().equals(
              connectionDetailsButtonStyle,
              other.connectionDetailsButtonStyle,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(maxConnectButtonWidth),
      const DeepCollectionEquality().hash(secureMyConnectionButtonStyle),
      const DeepCollectionEquality().hash(cancelButtonStyle),
      const DeepCollectionEquality().hash(connectionDetailsButtonStyle),
    );
  }
}

mixin _$ConnectionCardThemeTailorMixin on ThemeExtension<ConnectionCardTheme> {
  TextStyle get primaryFont;
  EdgeInsetsGeometry get mapPadding;
  EdgeInsetsGeometry get connectionCardPadding;
  EdgeInsetsGeometry get margin;
  BorderRadius get borderRadius;
  double get minWidth;
  double get smallSpacing;
  double get mediumSpacing;
  ConnectionCardLabelTheme get labelTheme;
  ConnectionCardIconTheme get iconTheme;
  ConnectionCardButtonTheme get buttonTheme;

  @override
  ConnectionCardTheme copyWith({
    TextStyle? primaryFont,
    EdgeInsetsGeometry? mapPadding,
    EdgeInsetsGeometry? connectionCardPadding,
    EdgeInsetsGeometry? margin,
    BorderRadius? borderRadius,
    double? minWidth,
    double? smallSpacing,
    double? mediumSpacing,
    ConnectionCardLabelTheme? labelTheme,
    ConnectionCardIconTheme? iconTheme,
    ConnectionCardButtonTheme? buttonTheme,
  }) {
    return ConnectionCardTheme(
      primaryFont: primaryFont ?? this.primaryFont,
      mapPadding: mapPadding ?? this.mapPadding,
      connectionCardPadding:
          connectionCardPadding ?? this.connectionCardPadding,
      margin: margin ?? this.margin,
      borderRadius: borderRadius ?? this.borderRadius,
      minWidth: minWidth ?? this.minWidth,
      smallSpacing: smallSpacing ?? this.smallSpacing,
      mediumSpacing: mediumSpacing ?? this.mediumSpacing,
      labelTheme: labelTheme ?? this.labelTheme,
      iconTheme: iconTheme ?? this.iconTheme,
      buttonTheme: buttonTheme ?? this.buttonTheme,
    );
  }

  @override
  ConnectionCardTheme lerp(
    covariant ThemeExtension<ConnectionCardTheme>? other,
    double t,
  ) {
    if (other is! ConnectionCardTheme) return this as ConnectionCardTheme;
    return ConnectionCardTheme(
      primaryFont: TextStyle.lerp(primaryFont, other.primaryFont, t)!,
      mapPadding: t < 0.5 ? mapPadding : other.mapPadding,
      connectionCardPadding: t < 0.5
          ? connectionCardPadding
          : other.connectionCardPadding,
      margin: t < 0.5 ? margin : other.margin,
      borderRadius: t < 0.5 ? borderRadius : other.borderRadius,
      minWidth: t < 0.5 ? minWidth : other.minWidth,
      smallSpacing: t < 0.5 ? smallSpacing : other.smallSpacing,
      mediumSpacing: t < 0.5 ? mediumSpacing : other.mediumSpacing,
      labelTheme: labelTheme.lerp(other.labelTheme, t),
      iconTheme: iconTheme.lerp(other.iconTheme, t),
      buttonTheme: buttonTheme.lerp(other.buttonTheme, t),
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is ConnectionCardTheme &&
            const DeepCollectionEquality().equals(
              primaryFont,
              other.primaryFont,
            ) &&
            const DeepCollectionEquality().equals(
              mapPadding,
              other.mapPadding,
            ) &&
            const DeepCollectionEquality().equals(
              connectionCardPadding,
              other.connectionCardPadding,
            ) &&
            const DeepCollectionEquality().equals(margin, other.margin) &&
            const DeepCollectionEquality().equals(
              borderRadius,
              other.borderRadius,
            ) &&
            const DeepCollectionEquality().equals(minWidth, other.minWidth) &&
            const DeepCollectionEquality().equals(
              smallSpacing,
              other.smallSpacing,
            ) &&
            const DeepCollectionEquality().equals(
              mediumSpacing,
              other.mediumSpacing,
            ) &&
            const DeepCollectionEquality().equals(
              labelTheme,
              other.labelTheme,
            ) &&
            const DeepCollectionEquality().equals(iconTheme, other.iconTheme) &&
            const DeepCollectionEquality().equals(
              buttonTheme,
              other.buttonTheme,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(primaryFont),
      const DeepCollectionEquality().hash(mapPadding),
      const DeepCollectionEquality().hash(connectionCardPadding),
      const DeepCollectionEquality().hash(margin),
      const DeepCollectionEquality().hash(borderRadius),
      const DeepCollectionEquality().hash(minWidth),
      const DeepCollectionEquality().hash(smallSpacing),
      const DeepCollectionEquality().hash(mediumSpacing),
      const DeepCollectionEquality().hash(labelTheme),
      const DeepCollectionEquality().hash(iconTheme),
      const DeepCollectionEquality().hash(buttonTheme),
    );
  }
}

extension ConnectionCardThemeBuildContextProps on BuildContext {
  ConnectionCardTheme get connectionCardTheme =>
      Theme.of(this).extension<ConnectionCardTheme>()!;
  TextStyle get primaryFont => connectionCardTheme.primaryFont;
  EdgeInsetsGeometry get mapPadding => connectionCardTheme.mapPadding;
  EdgeInsetsGeometry get connectionCardPadding =>
      connectionCardTheme.connectionCardPadding;
  EdgeInsetsGeometry get margin => connectionCardTheme.margin;
  BorderRadius get borderRadius => connectionCardTheme.borderRadius;
  double get minWidth => connectionCardTheme.minWidth;
  double get smallSpacing => connectionCardTheme.smallSpacing;
  double get mediumSpacing => connectionCardTheme.mediumSpacing;
  ConnectionCardLabelTheme get labelTheme => connectionCardTheme.labelTheme;
  ConnectionCardIconTheme get iconTheme => connectionCardTheme.iconTheme;
  ConnectionCardButtonTheme get buttonTheme => connectionCardTheme.buttonTheme;
}
