// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'dropdown_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$DropdownThemeTailorMixin on ThemeExtension<DropdownTheme> {
  Color get color;
  double get borderRadius;
  Color get borderColor;
  Color get focusBorderColor;
  Color get errorBorderColor;
  double get borderWidth;
  double get horizontalPadding;
  ButtonStyle get itemStyle;
  MenuStyle get menuStyle;

  @override
  DropdownTheme copyWith({
    Color? color,
    double? borderRadius,
    Color? borderColor,
    Color? focusBorderColor,
    Color? errorBorderColor,
    double? borderWidth,
    double? horizontalPadding,
    ButtonStyle? itemStyle,
    MenuStyle? menuStyle,
  }) {
    return DropdownTheme(
      color: color ?? this.color,
      borderRadius: borderRadius ?? this.borderRadius,
      borderColor: borderColor ?? this.borderColor,
      focusBorderColor: focusBorderColor ?? this.focusBorderColor,
      errorBorderColor: errorBorderColor ?? this.errorBorderColor,
      borderWidth: borderWidth ?? this.borderWidth,
      horizontalPadding: horizontalPadding ?? this.horizontalPadding,
      itemStyle: itemStyle ?? this.itemStyle,
      menuStyle: menuStyle ?? this.menuStyle,
    );
  }

  @override
  DropdownTheme lerp(covariant ThemeExtension<DropdownTheme>? other, double t) {
    if (other is! DropdownTheme) return this as DropdownTheme;
    return DropdownTheme(
      color: Color.lerp(color, other.color, t)!,
      borderRadius: t < 0.5 ? borderRadius : other.borderRadius,
      borderColor: Color.lerp(borderColor, other.borderColor, t)!,
      focusBorderColor: Color.lerp(
        focusBorderColor,
        other.focusBorderColor,
        t,
      )!,
      errorBorderColor: Color.lerp(
        errorBorderColor,
        other.errorBorderColor,
        t,
      )!,
      borderWidth: t < 0.5 ? borderWidth : other.borderWidth,
      horizontalPadding: t < 0.5 ? horizontalPadding : other.horizontalPadding,
      itemStyle: t < 0.5 ? itemStyle : other.itemStyle,
      menuStyle: t < 0.5 ? menuStyle : other.menuStyle,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is DropdownTheme &&
            const DeepCollectionEquality().equals(color, other.color) &&
            const DeepCollectionEquality().equals(
              borderRadius,
              other.borderRadius,
            ) &&
            const DeepCollectionEquality().equals(
              borderColor,
              other.borderColor,
            ) &&
            const DeepCollectionEquality().equals(
              focusBorderColor,
              other.focusBorderColor,
            ) &&
            const DeepCollectionEquality().equals(
              errorBorderColor,
              other.errorBorderColor,
            ) &&
            const DeepCollectionEquality().equals(
              borderWidth,
              other.borderWidth,
            ) &&
            const DeepCollectionEquality().equals(
              horizontalPadding,
              other.horizontalPadding,
            ) &&
            const DeepCollectionEquality().equals(itemStyle, other.itemStyle) &&
            const DeepCollectionEquality().equals(menuStyle, other.menuStyle));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(color),
      const DeepCollectionEquality().hash(borderRadius),
      const DeepCollectionEquality().hash(borderColor),
      const DeepCollectionEquality().hash(focusBorderColor),
      const DeepCollectionEquality().hash(errorBorderColor),
      const DeepCollectionEquality().hash(borderWidth),
      const DeepCollectionEquality().hash(horizontalPadding),
      const DeepCollectionEquality().hash(itemStyle),
      const DeepCollectionEquality().hash(menuStyle),
    );
  }
}

extension DropdownThemeBuildContextProps on BuildContext {
  DropdownTheme get dropdownTheme => Theme.of(this).extension<DropdownTheme>()!;
  Color get color => dropdownTheme.color;
  double get borderRadius => dropdownTheme.borderRadius;
  Color get borderColor => dropdownTheme.borderColor;
  Color get focusBorderColor => dropdownTheme.focusBorderColor;
  Color get errorBorderColor => dropdownTheme.errorBorderColor;
  double get borderWidth => dropdownTheme.borderWidth;
  double get horizontalPadding => dropdownTheme.horizontalPadding;
  ButtonStyle get itemStyle => dropdownTheme.itemStyle;
  MenuStyle get menuStyle => dropdownTheme.menuStyle;
}
