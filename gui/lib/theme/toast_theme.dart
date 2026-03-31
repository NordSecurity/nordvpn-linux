import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'toast_theme.tailor.dart';

@tailorMixin
final class ToastTheme extends ThemeExtension<ToastTheme>
    with _$ToastThemeTailorMixin {
  @override
  final TextStyle toastMessageTextStyle;

  @override
  final Color toastBackgroundColor;

  @override
  final double toastSpacing;

  @override
  final BorderRadius toastBorderRadius;

  @override
  final double widgetWidth;

  @override
  final double widgetHeight;

  @override
  final EdgeInsets toastCloseButtonPadding;

  @override
  final double toastBorderWidth;

  @override
  final Color toastBorderColor;

  ToastTheme({
    required this.toastMessageTextStyle,
    required this.toastBackgroundColor,
    required this.toastSpacing,
    required this.toastBorderRadius,
    required this.widgetWidth,
    required this.widgetHeight,
    required this.toastCloseButtonPadding,
    required this.toastBorderWidth,
    required this.toastBorderColor,
  });
}
