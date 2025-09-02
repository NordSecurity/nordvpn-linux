// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'error_screen_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$ErrorScreenThemeTailorMixin on ThemeExtension<ErrorScreenTheme> {
  TextStyle get titleTextStyle;
  TextStyle get descriptionTextStyle;

  @override
  ErrorScreenTheme copyWith({
    TextStyle? titleTextStyle,
    TextStyle? descriptionTextStyle,
  }) {
    return ErrorScreenTheme(
      titleTextStyle: titleTextStyle ?? this.titleTextStyle,
      descriptionTextStyle: descriptionTextStyle ?? this.descriptionTextStyle,
    );
  }

  @override
  ErrorScreenTheme lerp(
    covariant ThemeExtension<ErrorScreenTheme>? other,
    double t,
  ) {
    if (other is! ErrorScreenTheme) return this as ErrorScreenTheme;
    return ErrorScreenTheme(
      titleTextStyle: TextStyle.lerp(titleTextStyle, other.titleTextStyle, t)!,
      descriptionTextStyle: TextStyle.lerp(
        descriptionTextStyle,
        other.descriptionTextStyle,
        t,
      )!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is ErrorScreenTheme &&
            const DeepCollectionEquality().equals(
              titleTextStyle,
              other.titleTextStyle,
            ) &&
            const DeepCollectionEquality().equals(
              descriptionTextStyle,
              other.descriptionTextStyle,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(titleTextStyle),
      const DeepCollectionEquality().hash(descriptionTextStyle),
    );
  }
}

extension ErrorScreenThemeBuildContextProps on BuildContext {
  ErrorScreenTheme get errorScreenTheme =>
      Theme.of(this).extension<ErrorScreenTheme>()!;
  TextStyle get titleTextStyle => errorScreenTheme.titleTextStyle;
  TextStyle get descriptionTextStyle => errorScreenTheme.descriptionTextStyle;
}
