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

  @override
  VpnStatusCardTheme copyWith({
    double? height,
    double? maxConnectButtonWidth,
    TextStyle? primaryFont,
    TextStyle? secondaryFont,
    double? iconSize,
  }) {
    return VpnStatusCardTheme(
      height: height ?? this.height,
      maxConnectButtonWidth:
          maxConnectButtonWidth ?? this.maxConnectButtonWidth,
      primaryFont: primaryFont ?? this.primaryFont,
      secondaryFont: secondaryFont ?? this.secondaryFont,
      iconSize: iconSize ?? this.iconSize,
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
            const DeepCollectionEquality().equals(iconSize, other.iconSize));
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
}
