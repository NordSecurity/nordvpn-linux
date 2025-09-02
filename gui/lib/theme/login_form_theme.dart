// ignore_for_file: annotate_overrides

import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'login_form_theme.tailor.dart';

@tailorMixin
final class LoginFormTheme extends ThemeExtension<LoginFormTheme>
    with _$LoginFormThemeTailorMixin {
  final TextStyle titleStyle;
  final TextStyle checkboxDescStyle;
  final double height;
  final double width;
  final LoginButtonProgressIndicatorTheme progressIndicator;

  LoginFormTheme({
    required this.titleStyle,
    required this.checkboxDescStyle,
    required this.height,
    required this.width,
    required this.progressIndicator,
  });
}

@tailorMixinComponent
final class LoginButtonProgressIndicatorTheme
    extends ThemeExtension<LoginButtonProgressIndicatorTheme>
    with _$LoginButtonProgressIndicatorThemeTailorMixin {
  final double height;
  final double width;
  final double stroke;
  final Color color;

  LoginButtonProgressIndicatorTheme({
    required this.height,
    required this.width,
    required this.stroke,
    required this.color,
  });
}
