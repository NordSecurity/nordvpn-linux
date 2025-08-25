// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'input_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$InputThemeTailorMixin on ThemeExtension<InputTheme> {
  double get height;
  TextStyle get textStyle;
  TextStyle get errorStyle;
  EnabledStyle get enabled;
  FocusedStyle get focused;
  ErrorStyle get error;
  FocusedErrorStyle get focusedError;
  IconStyle get icon;

  @override
  InputTheme copyWith({
    double? height,
    TextStyle? textStyle,
    TextStyle? errorStyle,
    EnabledStyle? enabled,
    FocusedStyle? focused,
    ErrorStyle? error,
    FocusedErrorStyle? focusedError,
    IconStyle? icon,
  }) {
    return InputTheme(
      height: height ?? this.height,
      textStyle: textStyle ?? this.textStyle,
      errorStyle: errorStyle ?? this.errorStyle,
      enabled: enabled ?? this.enabled,
      focused: focused ?? this.focused,
      error: error ?? this.error,
      focusedError: focusedError ?? this.focusedError,
      icon: icon ?? this.icon,
    );
  }

  @override
  InputTheme lerp(covariant ThemeExtension<InputTheme>? other, double t) {
    if (other is! InputTheme) return this as InputTheme;
    return InputTheme(
      height: t < 0.5 ? height : other.height,
      textStyle: TextStyle.lerp(textStyle, other.textStyle, t)!,
      errorStyle: TextStyle.lerp(errorStyle, other.errorStyle, t)!,
      enabled: enabled.lerp(other.enabled, t),
      focused: focused.lerp(other.focused, t),
      error: error.lerp(other.error, t),
      focusedError: focusedError.lerp(other.focusedError, t),
      icon: icon.lerp(other.icon, t),
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is InputTheme &&
            const DeepCollectionEquality().equals(height, other.height) &&
            const DeepCollectionEquality().equals(textStyle, other.textStyle) &&
            const DeepCollectionEquality().equals(
              errorStyle,
              other.errorStyle,
            ) &&
            const DeepCollectionEquality().equals(enabled, other.enabled) &&
            const DeepCollectionEquality().equals(focused, other.focused) &&
            const DeepCollectionEquality().equals(error, other.error) &&
            const DeepCollectionEquality().equals(
              focusedError,
              other.focusedError,
            ) &&
            const DeepCollectionEquality().equals(icon, other.icon));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(height),
      const DeepCollectionEquality().hash(textStyle),
      const DeepCollectionEquality().hash(errorStyle),
      const DeepCollectionEquality().hash(enabled),
      const DeepCollectionEquality().hash(focused),
      const DeepCollectionEquality().hash(error),
      const DeepCollectionEquality().hash(focusedError),
      const DeepCollectionEquality().hash(icon),
    );
  }
}

extension InputThemeBuildContextProps on BuildContext {
  InputTheme get inputTheme => Theme.of(this).extension<InputTheme>()!;
  double get height => inputTheme.height;
  TextStyle get textStyle => inputTheme.textStyle;
  TextStyle get errorStyle => inputTheme.errorStyle;
  EnabledStyle get enabled => inputTheme.enabled;
  FocusedStyle get focused => inputTheme.focused;
  ErrorStyle get error => inputTheme.error;
  FocusedErrorStyle get focusedError => inputTheme.focusedError;
  IconStyle get icon => inputTheme.icon;
}

mixin _$EnabledStyleTailorMixin on ThemeExtension<EnabledStyle> {
  Color get borderColor;
  double get borderWidth;

  @override
  EnabledStyle copyWith({Color? borderColor, double? borderWidth}) {
    return EnabledStyle(
      borderColor: borderColor ?? this.borderColor,
      borderWidth: borderWidth ?? this.borderWidth,
    );
  }

  @override
  EnabledStyle lerp(covariant ThemeExtension<EnabledStyle>? other, double t) {
    if (other is! EnabledStyle) return this as EnabledStyle;
    return EnabledStyle(
      borderColor: Color.lerp(borderColor, other.borderColor, t)!,
      borderWidth: t < 0.5 ? borderWidth : other.borderWidth,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is EnabledStyle &&
            const DeepCollectionEquality().equals(
              borderColor,
              other.borderColor,
            ) &&
            const DeepCollectionEquality().equals(
              borderWidth,
              other.borderWidth,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(borderColor),
      const DeepCollectionEquality().hash(borderWidth),
    );
  }
}

mixin _$FocusedStyleTailorMixin on ThemeExtension<FocusedStyle> {
  Color get borderColor;
  double get borderWidth;

  @override
  FocusedStyle copyWith({Color? borderColor, double? borderWidth}) {
    return FocusedStyle(
      borderColor: borderColor ?? this.borderColor,
      borderWidth: borderWidth ?? this.borderWidth,
    );
  }

  @override
  FocusedStyle lerp(covariant ThemeExtension<FocusedStyle>? other, double t) {
    if (other is! FocusedStyle) return this as FocusedStyle;
    return FocusedStyle(
      borderColor: Color.lerp(borderColor, other.borderColor, t)!,
      borderWidth: t < 0.5 ? borderWidth : other.borderWidth,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is FocusedStyle &&
            const DeepCollectionEquality().equals(
              borderColor,
              other.borderColor,
            ) &&
            const DeepCollectionEquality().equals(
              borderWidth,
              other.borderWidth,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(borderColor),
      const DeepCollectionEquality().hash(borderWidth),
    );
  }
}

mixin _$ErrorStyleTailorMixin on ThemeExtension<ErrorStyle> {
  Color get borderColor;
  double get borderWidth;

  @override
  ErrorStyle copyWith({Color? borderColor, double? borderWidth}) {
    return ErrorStyle(
      borderColor: borderColor ?? this.borderColor,
      borderWidth: borderWidth ?? this.borderWidth,
    );
  }

  @override
  ErrorStyle lerp(covariant ThemeExtension<ErrorStyle>? other, double t) {
    if (other is! ErrorStyle) return this as ErrorStyle;
    return ErrorStyle(
      borderColor: Color.lerp(borderColor, other.borderColor, t)!,
      borderWidth: t < 0.5 ? borderWidth : other.borderWidth,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is ErrorStyle &&
            const DeepCollectionEquality().equals(
              borderColor,
              other.borderColor,
            ) &&
            const DeepCollectionEquality().equals(
              borderWidth,
              other.borderWidth,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(borderColor),
      const DeepCollectionEquality().hash(borderWidth),
    );
  }
}

mixin _$FocusedErrorStyleTailorMixin on ThemeExtension<FocusedErrorStyle> {
  Color get borderColor;
  double get borderWidth;

  @override
  FocusedErrorStyle copyWith({Color? borderColor, double? borderWidth}) {
    return FocusedErrorStyle(
      borderColor: borderColor ?? this.borderColor,
      borderWidth: borderWidth ?? this.borderWidth,
    );
  }

  @override
  FocusedErrorStyle lerp(
    covariant ThemeExtension<FocusedErrorStyle>? other,
    double t,
  ) {
    if (other is! FocusedErrorStyle) return this as FocusedErrorStyle;
    return FocusedErrorStyle(
      borderColor: Color.lerp(borderColor, other.borderColor, t)!,
      borderWidth: t < 0.5 ? borderWidth : other.borderWidth,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is FocusedErrorStyle &&
            const DeepCollectionEquality().equals(
              borderColor,
              other.borderColor,
            ) &&
            const DeepCollectionEquality().equals(
              borderWidth,
              other.borderWidth,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(borderColor),
      const DeepCollectionEquality().hash(borderWidth),
    );
  }
}

mixin _$IconStyleTailorMixin on ThemeExtension<IconStyle> {
  Color get color;
  Color get hoverColor;

  @override
  IconStyle copyWith({Color? color, Color? hoverColor}) {
    return IconStyle(
      color: color ?? this.color,
      hoverColor: hoverColor ?? this.hoverColor,
    );
  }

  @override
  IconStyle lerp(covariant ThemeExtension<IconStyle>? other, double t) {
    if (other is! IconStyle) return this as IconStyle;
    return IconStyle(
      color: Color.lerp(color, other.color, t)!,
      hoverColor: Color.lerp(hoverColor, other.hoverColor, t)!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is IconStyle &&
            const DeepCollectionEquality().equals(color, other.color) &&
            const DeepCollectionEquality().equals(
              hoverColor,
              other.hoverColor,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(color),
      const DeepCollectionEquality().hash(hoverColor),
    );
  }
}
