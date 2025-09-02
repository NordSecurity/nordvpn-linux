// ignore_for_file: annotate_overrides

import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'input_theme.tailor.dart';

@tailorMixin
final class InputTheme extends ThemeExtension<InputTheme>
    with _$InputThemeTailorMixin {
  final double height;
  final TextStyle textStyle;
  final TextStyle errorStyle;
  final EnabledStyle enabled;
  final FocusedStyle focused;
  final ErrorStyle error;
  final FocusedErrorStyle focusedError;
  final IconStyle icon;

  InputTheme({
    required this.height,
    required this.textStyle,
    required this.errorStyle,
    required this.enabled,
    required this.focused,
    required this.error,
    required this.focusedError,
    required this.icon,
  });
}

@tailorMixinComponent
final class EnabledStyle extends ThemeExtension<EnabledStyle>
    with _$EnabledStyleTailorMixin {
  final Color borderColor;
  final double borderWidth;

  EnabledStyle({required this.borderColor, required this.borderWidth});
}

@tailorMixinComponent
final class FocusedStyle extends ThemeExtension<FocusedStyle>
    with _$FocusedStyleTailorMixin {
  final Color borderColor;
  final double borderWidth;

  FocusedStyle({required this.borderColor, required this.borderWidth});
}

@tailorMixinComponent
final class ErrorStyle extends ThemeExtension<ErrorStyle>
    with _$ErrorStyleTailorMixin {
  final Color borderColor;
  final double borderWidth;

  ErrorStyle({required this.borderColor, required this.borderWidth});
}

@tailorMixinComponent
final class FocusedErrorStyle extends ThemeExtension<FocusedErrorStyle>
    with _$FocusedErrorStyleTailorMixin {
  final Color borderColor;
  final double borderWidth;

  FocusedErrorStyle({required this.borderColor, required this.borderWidth});
}

@tailorMixinComponent
final class IconStyle extends ThemeExtension<IconStyle>
    with _$IconStyleTailorMixin {
  final Color color;
  final Color hoverColor;

  IconStyle({required this.color, required this.hoverColor});
}
