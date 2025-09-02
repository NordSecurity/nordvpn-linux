// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'login_form_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$LoginFormThemeTailorMixin on ThemeExtension<LoginFormTheme> {
  TextStyle get titleStyle;
  TextStyle get checkboxDescStyle;
  double get height;
  double get width;
  LoginButtonProgressIndicatorTheme get progressIndicator;

  @override
  LoginFormTheme copyWith({
    TextStyle? titleStyle,
    TextStyle? checkboxDescStyle,
    double? height,
    double? width,
    LoginButtonProgressIndicatorTheme? progressIndicator,
  }) {
    return LoginFormTheme(
      titleStyle: titleStyle ?? this.titleStyle,
      checkboxDescStyle: checkboxDescStyle ?? this.checkboxDescStyle,
      height: height ?? this.height,
      width: width ?? this.width,
      progressIndicator: progressIndicator ?? this.progressIndicator,
    );
  }

  @override
  LoginFormTheme lerp(
    covariant ThemeExtension<LoginFormTheme>? other,
    double t,
  ) {
    if (other is! LoginFormTheme) return this as LoginFormTheme;
    return LoginFormTheme(
      titleStyle: TextStyle.lerp(titleStyle, other.titleStyle, t)!,
      checkboxDescStyle: TextStyle.lerp(
        checkboxDescStyle,
        other.checkboxDescStyle,
        t,
      )!,
      height: t < 0.5 ? height : other.height,
      width: t < 0.5 ? width : other.width,
      progressIndicator: progressIndicator.lerp(other.progressIndicator, t),
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is LoginFormTheme &&
            const DeepCollectionEquality().equals(
              titleStyle,
              other.titleStyle,
            ) &&
            const DeepCollectionEquality().equals(
              checkboxDescStyle,
              other.checkboxDescStyle,
            ) &&
            const DeepCollectionEquality().equals(height, other.height) &&
            const DeepCollectionEquality().equals(width, other.width) &&
            const DeepCollectionEquality().equals(
              progressIndicator,
              other.progressIndicator,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(titleStyle),
      const DeepCollectionEquality().hash(checkboxDescStyle),
      const DeepCollectionEquality().hash(height),
      const DeepCollectionEquality().hash(width),
      const DeepCollectionEquality().hash(progressIndicator),
    );
  }
}

extension LoginFormThemeBuildContextProps on BuildContext {
  LoginFormTheme get loginFormTheme =>
      Theme.of(this).extension<LoginFormTheme>()!;
  TextStyle get titleStyle => loginFormTheme.titleStyle;
  TextStyle get checkboxDescStyle => loginFormTheme.checkboxDescStyle;
  double get height => loginFormTheme.height;
  double get width => loginFormTheme.width;
  LoginButtonProgressIndicatorTheme get progressIndicator =>
      loginFormTheme.progressIndicator;
}

mixin _$LoginButtonProgressIndicatorThemeTailorMixin
    on ThemeExtension<LoginButtonProgressIndicatorTheme> {
  double get height;
  double get width;
  double get stroke;
  Color get color;

  @override
  LoginButtonProgressIndicatorTheme copyWith({
    double? height,
    double? width,
    double? stroke,
    Color? color,
  }) {
    return LoginButtonProgressIndicatorTheme(
      height: height ?? this.height,
      width: width ?? this.width,
      stroke: stroke ?? this.stroke,
      color: color ?? this.color,
    );
  }

  @override
  LoginButtonProgressIndicatorTheme lerp(
    covariant ThemeExtension<LoginButtonProgressIndicatorTheme>? other,
    double t,
  ) {
    if (other is! LoginButtonProgressIndicatorTheme)
      return this as LoginButtonProgressIndicatorTheme;
    return LoginButtonProgressIndicatorTheme(
      height: t < 0.5 ? height : other.height,
      width: t < 0.5 ? width : other.width,
      stroke: t < 0.5 ? stroke : other.stroke,
      color: Color.lerp(color, other.color, t)!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is LoginButtonProgressIndicatorTheme &&
            const DeepCollectionEquality().equals(height, other.height) &&
            const DeepCollectionEquality().equals(width, other.width) &&
            const DeepCollectionEquality().equals(stroke, other.stroke) &&
            const DeepCollectionEquality().equals(color, other.color));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(height),
      const DeepCollectionEquality().hash(width),
      const DeepCollectionEquality().hash(stroke),
      const DeepCollectionEquality().hash(color),
    );
  }
}
