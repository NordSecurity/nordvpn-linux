// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'autoconnect_panel_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$AutoconnectPanelThemeTailorMixin
    on ThemeExtension<AutoconnectPanelTheme> {
  TextStyle get primaryFont;
  TextStyle get secondaryFont;
  double get iconSize;
  double get loaderSize;

  @override
  AutoconnectPanelTheme copyWith({
    TextStyle? primaryFont,
    TextStyle? secondaryFont,
    double? iconSize,
    double? loaderSize,
  }) {
    return AutoconnectPanelTheme(
      primaryFont: primaryFont ?? this.primaryFont,
      secondaryFont: secondaryFont ?? this.secondaryFont,
      iconSize: iconSize ?? this.iconSize,
      loaderSize: loaderSize ?? this.loaderSize,
    );
  }

  @override
  AutoconnectPanelTheme lerp(
    covariant ThemeExtension<AutoconnectPanelTheme>? other,
    double t,
  ) {
    if (other is! AutoconnectPanelTheme) return this as AutoconnectPanelTheme;
    return AutoconnectPanelTheme(
      primaryFont: TextStyle.lerp(primaryFont, other.primaryFont, t)!,
      secondaryFont: TextStyle.lerp(secondaryFont, other.secondaryFont, t)!,
      iconSize: t < 0.5 ? iconSize : other.iconSize,
      loaderSize: t < 0.5 ? loaderSize : other.loaderSize,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is AutoconnectPanelTheme &&
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
              loaderSize,
              other.loaderSize,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(primaryFont),
      const DeepCollectionEquality().hash(secondaryFont),
      const DeepCollectionEquality().hash(iconSize),
      const DeepCollectionEquality().hash(loaderSize),
    );
  }
}

extension AutoconnectPanelThemeBuildContextProps on BuildContext {
  AutoconnectPanelTheme get autoconnectPanelTheme =>
      Theme.of(this).extension<AutoconnectPanelTheme>()!;
  TextStyle get primaryFont => autoconnectPanelTheme.primaryFont;
  TextStyle get secondaryFont => autoconnectPanelTheme.secondaryFont;
  double get iconSize => autoconnectPanelTheme.iconSize;
  double get loaderSize => autoconnectPanelTheme.loaderSize;
}
