// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'support_link_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$SupportLinkThemeTailorMixin on ThemeExtension<SupportLinkTheme> {
  TextStyle get textStyle;
  Color get urlColor;

  @override
  SupportLinkTheme copyWith({TextStyle? textStyle, Color? urlColor}) {
    return SupportLinkTheme(
      textStyle: textStyle ?? this.textStyle,
      urlColor: urlColor ?? this.urlColor,
    );
  }

  @override
  SupportLinkTheme lerp(
    covariant ThemeExtension<SupportLinkTheme>? other,
    double t,
  ) {
    if (other is! SupportLinkTheme) return this as SupportLinkTheme;
    return SupportLinkTheme(
      textStyle: TextStyle.lerp(textStyle, other.textStyle, t)!,
      urlColor: Color.lerp(urlColor, other.urlColor, t)!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is SupportLinkTheme &&
            const DeepCollectionEquality().equals(textStyle, other.textStyle) &&
            const DeepCollectionEquality().equals(urlColor, other.urlColor));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(textStyle),
      const DeepCollectionEquality().hash(urlColor),
    );
  }
}

extension SupportLinkThemeBuildContextProps on BuildContext {
  SupportLinkTheme get supportLinkTheme =>
      Theme.of(this).extension<SupportLinkTheme>()!;
  TextStyle get textStyle => supportLinkTheme.textStyle;
  Color get urlColor => supportLinkTheme.urlColor;
}
