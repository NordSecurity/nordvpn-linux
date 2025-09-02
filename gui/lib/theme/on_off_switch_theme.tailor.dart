// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'on_off_switch_theme.dart';

// **************************************************************************
// TailorAnnotationsGenerator
// **************************************************************************

mixin _$OnOffSwitchThemeTailorMixin on ThemeExtension<OnOffSwitchTheme> {
  OnOffLabelTheme get label;
  OnOffSliderTheme get slider;

  @override
  OnOffSwitchTheme copyWith({
    OnOffLabelTheme? label,
    OnOffSliderTheme? slider,
  }) {
    return OnOffSwitchTheme(
      label: label ?? this.label,
      slider: slider ?? this.slider,
    );
  }

  @override
  OnOffSwitchTheme lerp(
    covariant ThemeExtension<OnOffSwitchTheme>? other,
    double t,
  ) {
    if (other is! OnOffSwitchTheme) return this as OnOffSwitchTheme;
    return OnOffSwitchTheme(
      label: label.lerp(other.label, t),
      slider: slider.lerp(other.slider, t),
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is OnOffSwitchTheme &&
            const DeepCollectionEquality().equals(label, other.label) &&
            const DeepCollectionEquality().equals(slider, other.slider));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(label),
      const DeepCollectionEquality().hash(slider),
    );
  }
}

extension OnOffSwitchThemeBuildContextProps on BuildContext {
  OnOffSwitchTheme get onOffSwitchTheme =>
      Theme.of(this).extension<OnOffSwitchTheme>()!;
  OnOffLabelTheme get label => onOffSwitchTheme.label;
  OnOffSliderTheme get slider => onOffSwitchTheme.slider;
}

mixin _$OnOffLabelThemeTailorMixin on ThemeExtension<OnOffLabelTheme> {
  double get width;
  double get paddingRight;
  TextStyle get textStyle;
  TextStyle get disabledTextStyle;

  @override
  OnOffLabelTheme copyWith({
    double? width,
    double? paddingRight,
    TextStyle? textStyle,
    TextStyle? disabledTextStyle,
  }) {
    return OnOffLabelTheme(
      width: width ?? this.width,
      paddingRight: paddingRight ?? this.paddingRight,
      textStyle: textStyle ?? this.textStyle,
      disabledTextStyle: disabledTextStyle ?? this.disabledTextStyle,
    );
  }

  @override
  OnOffLabelTheme lerp(
    covariant ThemeExtension<OnOffLabelTheme>? other,
    double t,
  ) {
    if (other is! OnOffLabelTheme) return this as OnOffLabelTheme;
    return OnOffLabelTheme(
      width: t < 0.5 ? width : other.width,
      paddingRight: t < 0.5 ? paddingRight : other.paddingRight,
      textStyle: TextStyle.lerp(textStyle, other.textStyle, t)!,
      disabledTextStyle: TextStyle.lerp(
        disabledTextStyle,
        other.disabledTextStyle,
        t,
      )!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is OnOffLabelTheme &&
            const DeepCollectionEquality().equals(width, other.width) &&
            const DeepCollectionEquality().equals(
              paddingRight,
              other.paddingRight,
            ) &&
            const DeepCollectionEquality().equals(textStyle, other.textStyle) &&
            const DeepCollectionEquality().equals(
              disabledTextStyle,
              other.disabledTextStyle,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(width),
      const DeepCollectionEquality().hash(paddingRight),
      const DeepCollectionEquality().hash(textStyle),
      const DeepCollectionEquality().hash(disabledTextStyle),
    );
  }
}

mixin _$OnOffSliderThemeTailorMixin on ThemeExtension<OnOffSliderTheme> {
  double get width;
  double get height;
  double get bottomOffset;
  double get topOffset;
  double get borderRadius;
  SwitchOnOffProps get on;
  SwitchOnOffProps get off;
  SwitchOnOffProps get disabledOn;
  SwitchOnOffProps get disabledOff;

  @override
  OnOffSliderTheme copyWith({
    double? width,
    double? height,
    double? bottomOffset,
    double? topOffset,
    double? borderRadius,
    SwitchOnOffProps? on,
    SwitchOnOffProps? off,
    SwitchOnOffProps? disabledOn,
    SwitchOnOffProps? disabledOff,
  }) {
    return OnOffSliderTheme(
      width: width ?? this.width,
      height: height ?? this.height,
      bottomOffset: bottomOffset ?? this.bottomOffset,
      topOffset: topOffset ?? this.topOffset,
      borderRadius: borderRadius ?? this.borderRadius,
      on: on ?? this.on,
      off: off ?? this.off,
      disabledOn: disabledOn ?? this.disabledOn,
      disabledOff: disabledOff ?? this.disabledOff,
    );
  }

  @override
  OnOffSliderTheme lerp(
    covariant ThemeExtension<OnOffSliderTheme>? other,
    double t,
  ) {
    if (other is! OnOffSliderTheme) return this as OnOffSliderTheme;
    return OnOffSliderTheme(
      width: t < 0.5 ? width : other.width,
      height: t < 0.5 ? height : other.height,
      bottomOffset: t < 0.5 ? bottomOffset : other.bottomOffset,
      topOffset: t < 0.5 ? topOffset : other.topOffset,
      borderRadius: t < 0.5 ? borderRadius : other.borderRadius,
      on: on.lerp(other.on, t),
      off: off.lerp(other.off, t),
      disabledOn: disabledOn.lerp(other.disabledOn, t),
      disabledOff: disabledOff.lerp(other.disabledOff, t),
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is OnOffSliderTheme &&
            const DeepCollectionEquality().equals(width, other.width) &&
            const DeepCollectionEquality().equals(height, other.height) &&
            const DeepCollectionEquality().equals(
              bottomOffset,
              other.bottomOffset,
            ) &&
            const DeepCollectionEquality().equals(topOffset, other.topOffset) &&
            const DeepCollectionEquality().equals(
              borderRadius,
              other.borderRadius,
            ) &&
            const DeepCollectionEquality().equals(on, other.on) &&
            const DeepCollectionEquality().equals(off, other.off) &&
            const DeepCollectionEquality().equals(
              disabledOn,
              other.disabledOn,
            ) &&
            const DeepCollectionEquality().equals(
              disabledOff,
              other.disabledOff,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(width),
      const DeepCollectionEquality().hash(height),
      const DeepCollectionEquality().hash(bottomOffset),
      const DeepCollectionEquality().hash(topOffset),
      const DeepCollectionEquality().hash(borderRadius),
      const DeepCollectionEquality().hash(on),
      const DeepCollectionEquality().hash(off),
      const DeepCollectionEquality().hash(disabledOn),
      const DeepCollectionEquality().hash(disabledOff),
    );
  }
}

mixin _$SwitchOnOffPropsTailorMixin on ThemeExtension<SwitchOnOffProps> {
  double get leftOffset;
  double get rightOffset;
  Color get color;
  Color get borderColor;
  Color get backgroundColor;

  @override
  SwitchOnOffProps copyWith({
    double? leftOffset,
    double? rightOffset,
    Color? color,
    Color? borderColor,
    Color? backgroundColor,
  }) {
    return SwitchOnOffProps(
      leftOffset: leftOffset ?? this.leftOffset,
      rightOffset: rightOffset ?? this.rightOffset,
      color: color ?? this.color,
      borderColor: borderColor ?? this.borderColor,
      backgroundColor: backgroundColor ?? this.backgroundColor,
    );
  }

  @override
  SwitchOnOffProps lerp(
    covariant ThemeExtension<SwitchOnOffProps>? other,
    double t,
  ) {
    if (other is! SwitchOnOffProps) return this as SwitchOnOffProps;
    return SwitchOnOffProps(
      leftOffset: t < 0.5 ? leftOffset : other.leftOffset,
      rightOffset: t < 0.5 ? rightOffset : other.rightOffset,
      color: Color.lerp(color, other.color, t)!,
      borderColor: Color.lerp(borderColor, other.borderColor, t)!,
      backgroundColor: Color.lerp(backgroundColor, other.backgroundColor, t)!,
    );
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is SwitchOnOffProps &&
            const DeepCollectionEquality().equals(
              leftOffset,
              other.leftOffset,
            ) &&
            const DeepCollectionEquality().equals(
              rightOffset,
              other.rightOffset,
            ) &&
            const DeepCollectionEquality().equals(color, other.color) &&
            const DeepCollectionEquality().equals(
              borderColor,
              other.borderColor,
            ) &&
            const DeepCollectionEquality().equals(
              backgroundColor,
              other.backgroundColor,
            ));
  }

  @override
  int get hashCode {
    return Object.hash(
      runtimeType.hashCode,
      const DeepCollectionEquality().hash(leftOffset),
      const DeepCollectionEquality().hash(rightOffset),
      const DeepCollectionEquality().hash(color),
      const DeepCollectionEquality().hash(borderColor),
      const DeepCollectionEquality().hash(backgroundColor),
    );
  }
}
