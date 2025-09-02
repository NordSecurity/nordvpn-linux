// ignore_for_file: annotate_overrides

import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'radio_button_theme.tailor.dart';

@tailorMixin
final class RadioButtonTheme extends ThemeExtension<RadioButtonTheme>
    with _$RadioButtonThemeTailorMixin {
  final RadioLabelTheme label;
  final RadioStyle radio;
  final double padding;

  RadioButtonTheme({
    required this.label,
    required this.radio,
    required this.padding,
  });
}

@tailorMixinComponent
final class RadioLabelTheme extends ThemeExtension<RadioLabelTheme>
    with _$RadioLabelThemeTailorMixin {
  final double width;
  final double paddingLeft;

  RadioLabelTheme({required this.width, required this.paddingLeft});
}

@tailorMixinComponent
final class RadioStyle extends ThemeExtension<RadioStyle>
    with _$RadioStyleTailorMixin {
  final double width;
  final double height;

  final RadioOnOffProps on;
  final RadioOnOffProps off;

  final double borderWidth;

  RadioStyle({
    required this.width,
    required this.height,
    required this.on,
    required this.off,
    required this.borderWidth,
  });
}

@tailorMixinComponent
final class RadioOnOffProps extends ThemeExtension<RadioOnOffProps>
    with _$RadioOnOffPropsTailorMixin {
  final Color fillColor;
  final Color borderColor;
  final Color dotColor;
  final double dotWidth;
  final double dotHeight;
  final double borderWidth;

  RadioOnOffProps({
    required this.fillColor,
    required this.borderColor,
    required this.dotColor,
    required this.dotWidth,
    required this.dotHeight,
    required this.borderWidth,
  });
}
