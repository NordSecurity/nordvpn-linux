import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'on_off_switch_theme.tailor.dart';

@tailorMixin
final class OnOffSwitchTheme extends ThemeExtension<OnOffSwitchTheme>
    with _$OnOffSwitchThemeTailorMixin {
  @override
  final OnOffLabelTheme label;

  @override
  final OnOffSliderTheme slider;

  OnOffSwitchTheme({required this.label, required this.slider});
}

@tailorMixinComponent
final class OnOffLabelTheme extends ThemeExtension<OnOffLabelTheme>
    with _$OnOffLabelThemeTailorMixin {
  @override
  final double width;

  @override
  final double paddingRight;

  @override
  final TextStyle textStyle;

  @override
  final TextStyle disabledTextStyle;

  OnOffLabelTheme({
    required this.width,
    required this.paddingRight,
    required this.textStyle,
    required this.disabledTextStyle,
  });
}

@tailorMixinComponent
final class OnOffSliderTheme extends ThemeExtension<OnOffSliderTheme>
    with _$OnOffSliderThemeTailorMixin {
  @override
  final double width;

  @override
  final double height;

  @override
  final double bottomOffset;

  @override
  final double topOffset;

  @override
  final double borderRadius;

  @override
  final SwitchOnOffProps on;

  @override
  final SwitchOnOffProps off;

  @override
  final SwitchOnOffProps disabledOn;

  @override
  final SwitchOnOffProps disabledOff;

  OnOffSliderTheme({
    required this.width,
    required this.height,
    required this.on,
    required this.off,
    required this.disabledOn,
    required this.disabledOff,
    required this.bottomOffset,
    required this.topOffset,
    required this.borderRadius,
  });
}

@tailorMixinComponent
final class SwitchOnOffProps extends ThemeExtension<SwitchOnOffProps>
    with _$SwitchOnOffPropsTailorMixin {
  @override
  final double leftOffset;

  @override
  final double rightOffset;

  @override
  final Color color;

  @override
  final Color borderColor;

  @override
  final Color backgroundColor;

  SwitchOnOffProps({
    required this.leftOffset,
    required this.rightOffset,
    required this.color,
    required this.borderColor,
    required this.backgroundColor,
  });
}
