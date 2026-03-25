// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'vpn_status_card_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$VpnStatusCardThemeTailorMixin on ThemeExtension<VpnStatusCardTheme> {
  double get height;
  double get maxConnectButtonWidth;
  TextStyle get primaryFont;
  TextStyle get secondaryFont;
  double get iconSize;
  ButtonStyle get secureMyConnectionButtonStyle;
  EdgeInsetsGeometry get connectionCardPadding;
  double get smallSpacing;
  double get mediumSpacing;
  Color get iconBackgroundColor;
  String get disconnectedIcon;

  @override
  VpnStatusCardTheme copyWith({
    double? height,
    double? maxConnectButtonWidth,
    TextStyle? primaryFont,
    TextStyle? secondaryFont,
    double? iconSize,
    ButtonStyle? secureMyConnectionButtonStyle,
    EdgeInsetsGeometry? connectionCardPadding,
    double? smallSpacing,
    double? mediumSpacing,
    Color? iconBackgroundColor,
    String? disconnectedIcon,
  }) {
    return VpnStatusCardTheme(
      height: height ?? this.height,
      maxConnectButtonWidth:
          maxConnectButtonWidth ?? this.maxConnectButtonWidth,
      primaryFont: primaryFont ?? this.primaryFont,
      secondaryFont: secondaryFont ?? this.secondaryFont,
      iconSize: iconSize ?? this.iconSize,
      secureMyConnectionButtonStyle:
          secureMyConnectionButtonStyle ?? this.secureMyConnectionButtonStyle,
      connectionCardPadding:
          connectionCardPadding ?? this.connectionCardPadding,
      smallSpacing: smallSpacing ?? this.smallSpacing,
      mediumSpacing: mediumSpacing ?? this.mediumSpacing,
      iconBackgroundColor: iconBackgroundColor ?? this.iconBackgroundColor,
      disconnectedIcon: disconnectedIcon ?? this.disconnectedIcon,
    );
  }

  @override
  VpnStatusCardTheme lerp(
    covariant ThemeExtension<VpnStatusCardTheme>? other,
    double t,
  ) {
    if (other is! VpnStatusCardTheme) return this as VpnStatusCardTheme;
    return VpnStatusCardTheme(
      height: t < 0.5 ? height : other.height,
      maxConnectButtonWidth: t < 0.5
          ? maxConnectButtonWidth
          : other.maxConnectButtonWidth,
      primaryFont: TextStyle.lerp(primaryFont, other.primaryFont, t)!,
      secondaryFont: TextStyle.lerp(secondaryFont, other.secondaryFont, t)!,
      iconSize: t < 0.5 ? iconSize : other.iconSize,
      secureMyConnectionButtonStyle: t < 0.5
          ? secureMyConnectionButtonStyle
          : other.secureMyConnectionButtonStyle,
      connectionCardPadding: t < 0.5
          ? connectionCardPadding
          : other.connectionCardPadding,
      smallSpacing: t < 0.5 ? smallSpacing : other.smallSpacing,
      mediumSpacing: t < 0.5 ? mediumSpacing : other.mediumSpacing,
      iconBackgroundColor: Color.lerp(
        iconBackgroundColor,
        other.iconBackgroundColor,
        t,
      )!,
      disconnectedIcon: t < 0.5 ? disconnectedIcon : other.disconnectedIcon,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is VpnStatusCardTheme &&
            const DeepCollectionEquality().equals(height, other.height) &&
            const DeepCollectionEquality().equals(
              maxConnectButtonWidth,
              other.maxConnectButtonWidth,
            ) &&
            const DeepCollectionEquality().equals(
              primaryFont,
              other.primaryFont,
            ) &&
            const DeepCollectionEquality().equals(
              secondaryFont,
              other.secondaryFont,
            ) &&
            const DeepCollectionEquality().equals(iconSize, other.iconSize) &&
            const DeepCollectionEquality().equals(
              secureMyConnectionButtonStyle,
              other.secureMyConnectionButtonStyle,
            ) &&
            const DeepCollectionEquality().equals(
              connectionCardPadding,
              other.connectionCardPadding,
            ) &&
            const DeepCollectionEquality().equals(
              smallSpacing,
              other.smallSpacing,
            ) &&
            const DeepCollectionEquality().equals(
              mediumSpacing,
              other.mediumSpacing,
            ) &&
            const DeepCollectionEquality().equals(
              iconBackgroundColor,
              other.iconBackgroundColor,
            ) &&
            const DeepCollectionEquality().equals(
              disconnectedIcon,
              other.disconnectedIcon,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(height),
      const DeepCollectionEquality().hash(maxConnectButtonWidth),
      const DeepCollectionEquality().hash(primaryFont),
      const DeepCollectionEquality().hash(secondaryFont),
      const DeepCollectionEquality().hash(iconSize),
      const DeepCollectionEquality().hash(secureMyConnectionButtonStyle),
      const DeepCollectionEquality().hash(connectionCardPadding),
      const DeepCollectionEquality().hash(smallSpacing),
      const DeepCollectionEquality().hash(mediumSpacing),
      const DeepCollectionEquality().hash(iconBackgroundColor),
      const DeepCollectionEquality().hash(disconnectedIcon),
    );
  }
}

extension VpnStatusCardThemeBuildContextProps on BuildContext {
  VpnStatusCardTheme get vpnStatusCardTheme =>
      Theme.of(this).extension<VpnStatusCardTheme>()!;
  double get height => vpnStatusCardTheme.height;
  double get maxConnectButtonWidth => vpnStatusCardTheme.maxConnectButtonWidth;
  TextStyle get primaryFont => vpnStatusCardTheme.primaryFont;
  TextStyle get secondaryFont => vpnStatusCardTheme.secondaryFont;
  double get iconSize => vpnStatusCardTheme.iconSize;
  ButtonStyle get secureMyConnectionButtonStyle =>
      vpnStatusCardTheme.secureMyConnectionButtonStyle;
  EdgeInsetsGeometry get connectionCardPadding =>
      vpnStatusCardTheme.connectionCardPadding;
  double get smallSpacing => vpnStatusCardTheme.smallSpacing;
  double get mediumSpacing => vpnStatusCardTheme.mediumSpacing;
  Color get iconBackgroundColor => vpnStatusCardTheme.iconBackgroundColor;
  String get disconnectedIcon => vpnStatusCardTheme.disconnectedIcon;
}
