// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'settings_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$SettingsThemeTailorMixin on ThemeExtension<SettingsTheme> {
  TextStyle get currentPageNameStyle;
  TextStyle get parentPageStyle;
  TextStyle get itemTitleStyle;
  TextStyle get itemSubtitleStyle;
  TextStyle get vpnStatusStyle;
  double get textInputWidth;
  TextStyle get otherProductsTitle;
  TextStyle get otherProductsSubtitle;
  double get fwMarkInputSize;
  EdgeInsets get itemPadding;

  @override
  SettingsTheme copyWith({
    TextStyle? currentPageNameStyle,
    TextStyle? parentPageStyle,
    TextStyle? itemTitleStyle,
    TextStyle? itemSubtitleStyle,
    TextStyle? vpnStatusStyle,
    double? textInputWidth,
    TextStyle? otherProductsTitle,
    TextStyle? otherProductsSubtitle,
    double? fwMarkInputSize,
    EdgeInsets? itemPadding,
  }) {
    return SettingsTheme(
      currentPageNameStyle: currentPageNameStyle ?? this.currentPageNameStyle,
      parentPageStyle: parentPageStyle ?? this.parentPageStyle,
      itemTitleStyle: itemTitleStyle ?? this.itemTitleStyle,
      itemSubtitleStyle: itemSubtitleStyle ?? this.itemSubtitleStyle,
      vpnStatusStyle: vpnStatusStyle ?? this.vpnStatusStyle,
      textInputWidth: textInputWidth ?? this.textInputWidth,
      otherProductsTitle: otherProductsTitle ?? this.otherProductsTitle,
      otherProductsSubtitle:
          otherProductsSubtitle ?? this.otherProductsSubtitle,
      fwMarkInputSize: fwMarkInputSize ?? this.fwMarkInputSize,
      itemPadding: itemPadding ?? this.itemPadding,
    );
  }

  @override
  SettingsTheme lerp(covariant ThemeExtension<SettingsTheme>? other, double t) {
    if (other is! SettingsTheme) return this as SettingsTheme;
    return SettingsTheme(
      currentPageNameStyle: TextStyle.lerp(
        currentPageNameStyle,
        other.currentPageNameStyle,
        t,
      )!,
      parentPageStyle: TextStyle.lerp(
        parentPageStyle,
        other.parentPageStyle,
        t,
      )!,
      itemTitleStyle: TextStyle.lerp(itemTitleStyle, other.itemTitleStyle, t)!,
      itemSubtitleStyle: TextStyle.lerp(
        itemSubtitleStyle,
        other.itemSubtitleStyle,
        t,
      )!,
      vpnStatusStyle: TextStyle.lerp(vpnStatusStyle, other.vpnStatusStyle, t)!,
      textInputWidth: t < 0.5 ? textInputWidth : other.textInputWidth,
      otherProductsTitle: TextStyle.lerp(
        otherProductsTitle,
        other.otherProductsTitle,
        t,
      )!,
      otherProductsSubtitle: TextStyle.lerp(
        otherProductsSubtitle,
        other.otherProductsSubtitle,
        t,
      )!,
      fwMarkInputSize: t < 0.5 ? fwMarkInputSize : other.fwMarkInputSize,
      itemPadding: t < 0.5 ? itemPadding : other.itemPadding,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is SettingsTheme &&
            const DeepCollectionEquality().equals(
              currentPageNameStyle,
              other.currentPageNameStyle,
            ) &&
            const DeepCollectionEquality().equals(
              parentPageStyle,
              other.parentPageStyle,
            ) &&
            const DeepCollectionEquality().equals(
              itemTitleStyle,
              other.itemTitleStyle,
            ) &&
            const DeepCollectionEquality().equals(
              itemSubtitleStyle,
              other.itemSubtitleStyle,
            ) &&
            const DeepCollectionEquality().equals(
              vpnStatusStyle,
              other.vpnStatusStyle,
            ) &&
            const DeepCollectionEquality().equals(
              textInputWidth,
              other.textInputWidth,
            ) &&
            const DeepCollectionEquality().equals(
              otherProductsTitle,
              other.otherProductsTitle,
            ) &&
            const DeepCollectionEquality().equals(
              otherProductsSubtitle,
              other.otherProductsSubtitle,
            ) &&
            const DeepCollectionEquality().equals(
              fwMarkInputSize,
              other.fwMarkInputSize,
            ) &&
            const DeepCollectionEquality().equals(
              itemPadding,
              other.itemPadding,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(currentPageNameStyle),
      const DeepCollectionEquality().hash(parentPageStyle),
      const DeepCollectionEquality().hash(itemTitleStyle),
      const DeepCollectionEquality().hash(itemSubtitleStyle),
      const DeepCollectionEquality().hash(vpnStatusStyle),
      const DeepCollectionEquality().hash(textInputWidth),
      const DeepCollectionEquality().hash(otherProductsTitle),
      const DeepCollectionEquality().hash(otherProductsSubtitle),
      const DeepCollectionEquality().hash(fwMarkInputSize),
      const DeepCollectionEquality().hash(itemPadding),
    );
  }
}

extension SettingsThemeBuildContextProps on BuildContext {
  SettingsTheme get settingsTheme => Theme.of(this).extension<SettingsTheme>()!;
  TextStyle get currentPageNameStyle => settingsTheme.currentPageNameStyle;
  TextStyle get parentPageStyle => settingsTheme.parentPageStyle;
  TextStyle get itemTitleStyle => settingsTheme.itemTitleStyle;
  TextStyle get itemSubtitleStyle => settingsTheme.itemSubtitleStyle;
  TextStyle get vpnStatusStyle => settingsTheme.vpnStatusStyle;
  double get textInputWidth => settingsTheme.textInputWidth;
  TextStyle get otherProductsTitle => settingsTheme.otherProductsTitle;
  TextStyle get otherProductsSubtitle => settingsTheme.otherProductsSubtitle;
  double get fwMarkInputSize => settingsTheme.fwMarkInputSize;
  EdgeInsets get itemPadding => settingsTheme.itemPadding;
}
