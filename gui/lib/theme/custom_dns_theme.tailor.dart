// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'custom_dns_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$CustomDnsThemeTailorMixin on ThemeExtension<CustomDnsTheme> {
  Color get formBackground;
  double get dnsInputWidth;
  Color get dividerColor;

  @override
  CustomDnsTheme copyWith({
    Color? formBackground,
    double? dnsInputWidth,
    Color? dividerColor,
  }) {
    return CustomDnsTheme(
      formBackground: formBackground ?? this.formBackground,
      dnsInputWidth: dnsInputWidth ?? this.dnsInputWidth,
      dividerColor: dividerColor ?? this.dividerColor,
    );
  }

  @override
  CustomDnsTheme lerp(
    covariant ThemeExtension<CustomDnsTheme>? other,
    double t,
  ) {
    if (other is! CustomDnsTheme) return this as CustomDnsTheme;
    return CustomDnsTheme(
      formBackground: Color.lerp(formBackground, other.formBackground, t)!,
      dnsInputWidth: t < 0.5 ? dnsInputWidth : other.dnsInputWidth,
      dividerColor: Color.lerp(dividerColor, other.dividerColor, t)!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is CustomDnsTheme &&
            const DeepCollectionEquality().equals(
              formBackground,
              other.formBackground,
            ) &&
            const DeepCollectionEquality().equals(
              dnsInputWidth,
              other.dnsInputWidth,
            ) &&
            const DeepCollectionEquality().equals(
              dividerColor,
              other.dividerColor,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(formBackground),
      const DeepCollectionEquality().hash(dnsInputWidth),
      const DeepCollectionEquality().hash(dividerColor),
    );
  }
}

extension CustomDnsThemeBuildContextProps on BuildContext {
  CustomDnsTheme get customDnsTheme =>
      Theme.of(this).extension<CustomDnsTheme>()!;
  Color get formBackground => customDnsTheme.formBackground;
  double get dnsInputWidth => customDnsTheme.dnsInputWidth;
  Color get dividerColor => customDnsTheme.dividerColor;
}
