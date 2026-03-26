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
  ButtonStyle get secureMyConnectionButtonStyle;
  ButtonStyle get cancelButtonStyle;
  EdgeInsetsGeometry get connectionCardPadding;
  double get smallSpacing;
  double get mediumSpacing;
  ConnectionCardLabelThemeStyle get labelStyle;
  ConnectionCardIconThemeStyle get iconStyle;

  @override
  VpnStatusCardTheme copyWith({
    double? height,
    double? maxConnectButtonWidth,
    TextStyle? primaryFont,
    TextStyle? secondaryFont,
    ButtonStyle? secureMyConnectionButtonStyle,
    ButtonStyle? cancelButtonStyle,
    EdgeInsetsGeometry? connectionCardPadding,
    double? smallSpacing,
    double? mediumSpacing,
    ConnectionCardLabelThemeStyle? labelStyle,
    ConnectionCardIconThemeStyle? iconStyle,
  }) {
    return VpnStatusCardTheme(
      height: height ?? this.height,
      maxConnectButtonWidth:
          maxConnectButtonWidth ?? this.maxConnectButtonWidth,
      primaryFont: primaryFont ?? this.primaryFont,
      secondaryFont: secondaryFont ?? this.secondaryFont,
      secureMyConnectionButtonStyle:
          secureMyConnectionButtonStyle ?? this.secureMyConnectionButtonStyle,
      cancelButtonStyle: cancelButtonStyle ?? this.cancelButtonStyle,
      connectionCardPadding:
          connectionCardPadding ?? this.connectionCardPadding,
      smallSpacing: smallSpacing ?? this.smallSpacing,
      mediumSpacing: mediumSpacing ?? this.mediumSpacing,
      labelStyle: labelStyle ?? this.labelStyle,
      iconStyle: iconStyle ?? this.iconStyle,
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
      secureMyConnectionButtonStyle: t < 0.5
          ? secureMyConnectionButtonStyle
          : other.secureMyConnectionButtonStyle,
      cancelButtonStyle: t < 0.5 ? cancelButtonStyle : other.cancelButtonStyle,
      connectionCardPadding: t < 0.5
          ? connectionCardPadding
          : other.connectionCardPadding,
      smallSpacing: t < 0.5 ? smallSpacing : other.smallSpacing,
      mediumSpacing: t < 0.5 ? mediumSpacing : other.mediumSpacing,
      labelStyle: t < 0.5 ? labelStyle : other.labelStyle,
      iconStyle: t < 0.5 ? iconStyle : other.iconStyle,
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
            const DeepCollectionEquality().equals(
              secureMyConnectionButtonStyle,
              other.secureMyConnectionButtonStyle,
            ) &&
            const DeepCollectionEquality().equals(
              cancelButtonStyle,
              other.cancelButtonStyle,
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
              labelStyle,
              other.labelStyle,
            ) &&
            const DeepCollectionEquality().equals(iconStyle, other.iconStyle));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(height),
      const DeepCollectionEquality().hash(maxConnectButtonWidth),
      const DeepCollectionEquality().hash(primaryFont),
      const DeepCollectionEquality().hash(secondaryFont),
      const DeepCollectionEquality().hash(secureMyConnectionButtonStyle),
      const DeepCollectionEquality().hash(cancelButtonStyle),
      const DeepCollectionEquality().hash(connectionCardPadding),
      const DeepCollectionEquality().hash(smallSpacing),
      const DeepCollectionEquality().hash(mediumSpacing),
      const DeepCollectionEquality().hash(labelStyle),
      const DeepCollectionEquality().hash(iconStyle),
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
  ButtonStyle get secureMyConnectionButtonStyle =>
      vpnStatusCardTheme.secureMyConnectionButtonStyle;
  ButtonStyle get cancelButtonStyle => vpnStatusCardTheme.cancelButtonStyle;
  EdgeInsetsGeometry get connectionCardPadding =>
      vpnStatusCardTheme.connectionCardPadding;
  double get smallSpacing => vpnStatusCardTheme.smallSpacing;
  double get mediumSpacing => vpnStatusCardTheme.mediumSpacing;
  ConnectionCardLabelThemeStyle get labelStyle => vpnStatusCardTheme.labelStyle;
  ConnectionCardIconThemeStyle get iconStyle => vpnStatusCardTheme.iconStyle;
}
