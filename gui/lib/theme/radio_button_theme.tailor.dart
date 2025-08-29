// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'radio_button_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$RadioButtonThemeTailorMixin on ThemeExtension<RadioButtonTheme> {
  RadioLabelTheme get label;
  RadioStyle get radio;
  double get padding;

  @override
  RadioButtonTheme copyWith({
    RadioLabelTheme? label,
    RadioStyle? radio,
    double? padding,
  }) {
    return RadioButtonTheme(
      label: label ?? this.label,
      radio: radio ?? this.radio,
      padding: padding ?? this.padding,
    );
  }

  @override
  RadioButtonTheme lerp(
    covariant ThemeExtension<RadioButtonTheme>? other,
    double t,
  ) {
    if (other is! RadioButtonTheme) return this as RadioButtonTheme;
    return RadioButtonTheme(
      label: label.lerp(other.label, t),
      radio: radio.lerp(other.radio, t),
      padding: t < 0.5 ? padding : other.padding,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is RadioButtonTheme &&
            const DeepCollectionEquality().equals(label, other.label) &&
            const DeepCollectionEquality().equals(radio, other.radio) &&
            const DeepCollectionEquality().equals(padding, other.padding));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(label),
      const DeepCollectionEquality().hash(radio),
      const DeepCollectionEquality().hash(padding),
    );
  }
}

extension RadioButtonThemeBuildContextProps on BuildContext {
  RadioButtonTheme get radioButtonTheme =>
      Theme.of(this).extension<RadioButtonTheme>()!;
  RadioLabelTheme get label => radioButtonTheme.label;
  RadioStyle get radio => radioButtonTheme.radio;
  double get padding => radioButtonTheme.padding;
}

mixin _$RadioLabelThemeTailorMixin on ThemeExtension<RadioLabelTheme> {
  double get width;
  double get paddingLeft;

  @override
  RadioLabelTheme copyWith({double? width, double? paddingLeft}) {
    return RadioLabelTheme(
      width: width ?? this.width,
      paddingLeft: paddingLeft ?? this.paddingLeft,
    );
  }

  @override
  RadioLabelTheme lerp(
    covariant ThemeExtension<RadioLabelTheme>? other,
    double t,
  ) {
    if (other is! RadioLabelTheme) return this as RadioLabelTheme;
    return RadioLabelTheme(
      width: t < 0.5 ? width : other.width,
      paddingLeft: t < 0.5 ? paddingLeft : other.paddingLeft,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is RadioLabelTheme &&
            const DeepCollectionEquality().equals(width, other.width) &&
            const DeepCollectionEquality().equals(
              paddingLeft,
              other.paddingLeft,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(width),
      const DeepCollectionEquality().hash(paddingLeft),
    );
  }
}

mixin _$RadioStyleTailorMixin on ThemeExtension<RadioStyle> {
  double get width;
  double get height;
  RadioOnOffProps get on;
  RadioOnOffProps get off;
  double get borderWidth;

  @override
  RadioStyle copyWith({
    double? width,
    double? height,
    RadioOnOffProps? on,
    RadioOnOffProps? off,
    double? borderWidth,
  }) {
    return RadioStyle(
      width: width ?? this.width,
      height: height ?? this.height,
      on: on ?? this.on,
      off: off ?? this.off,
      borderWidth: borderWidth ?? this.borderWidth,
    );
  }

  @override
  RadioStyle lerp(covariant ThemeExtension<RadioStyle>? other, double t) {
    if (other is! RadioStyle) return this as RadioStyle;
    return RadioStyle(
      width: t < 0.5 ? width : other.width,
      height: t < 0.5 ? height : other.height,
      on: on.lerp(other.on, t),
      off: off.lerp(other.off, t),
      borderWidth: t < 0.5 ? borderWidth : other.borderWidth,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is RadioStyle &&
            const DeepCollectionEquality().equals(width, other.width) &&
            const DeepCollectionEquality().equals(height, other.height) &&
            const DeepCollectionEquality().equals(on, other.on) &&
            const DeepCollectionEquality().equals(off, other.off) &&
            const DeepCollectionEquality().equals(
              borderWidth,
              other.borderWidth,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(width),
      const DeepCollectionEquality().hash(height),
      const DeepCollectionEquality().hash(on),
      const DeepCollectionEquality().hash(off),
      const DeepCollectionEquality().hash(borderWidth),
    );
  }
}

mixin _$RadioOnOffPropsTailorMixin on ThemeExtension<RadioOnOffProps> {
  Color get fillColor;
  Color get borderColor;
  Color get dotColor;
  double get dotWidth;
  double get dotHeight;
  double get borderWidth;

  @override
  RadioOnOffProps copyWith({
    Color? fillColor,
    Color? borderColor,
    Color? dotColor,
    double? dotWidth,
    double? dotHeight,
    double? borderWidth,
  }) {
    return RadioOnOffProps(
      fillColor: fillColor ?? this.fillColor,
      borderColor: borderColor ?? this.borderColor,
      dotColor: dotColor ?? this.dotColor,
      dotWidth: dotWidth ?? this.dotWidth,
      dotHeight: dotHeight ?? this.dotHeight,
      borderWidth: borderWidth ?? this.borderWidth,
    );
  }

  @override
  RadioOnOffProps lerp(
    covariant ThemeExtension<RadioOnOffProps>? other,
    double t,
  ) {
    if (other is! RadioOnOffProps) return this as RadioOnOffProps;
    return RadioOnOffProps(
      fillColor: Color.lerp(fillColor, other.fillColor, t)!,
      borderColor: Color.lerp(borderColor, other.borderColor, t)!,
      dotColor: Color.lerp(dotColor, other.dotColor, t)!,
      dotWidth: t < 0.5 ? dotWidth : other.dotWidth,
      dotHeight: t < 0.5 ? dotHeight : other.dotHeight,
      borderWidth: t < 0.5 ? borderWidth : other.borderWidth,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is RadioOnOffProps &&
            const DeepCollectionEquality().equals(fillColor, other.fillColor) &&
            const DeepCollectionEquality().equals(
              borderColor,
              other.borderColor,
            ) &&
            const DeepCollectionEquality().equals(dotColor, other.dotColor) &&
            const DeepCollectionEquality().equals(dotWidth, other.dotWidth) &&
            const DeepCollectionEquality().equals(dotHeight, other.dotHeight) &&
            const DeepCollectionEquality().equals(
              borderWidth,
              other.borderWidth,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(fillColor),
      const DeepCollectionEquality().hash(borderColor),
      const DeepCollectionEquality().hash(dotColor),
      const DeepCollectionEquality().hash(dotWidth),
      const DeepCollectionEquality().hash(dotHeight),
      const DeepCollectionEquality().hash(borderWidth),
    );
  }
}
