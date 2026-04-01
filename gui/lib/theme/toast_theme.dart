import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'toast_theme.tailor.dart';

@tailorMixin
final class ToastTheme extends ThemeExtension<ToastTheme>
    with _$ToastThemeTailorMixin {
  @override
  final TextStyle messageTextStyle;

  @override
  final Color backgroundColor;

  @override
  final double spacing;

  @override
  final BorderRadius borderRadius;

  @override
  final double widgetWidth;

  @override
  final double widgetHeight;

  @override
  final double widgetPositionRight;

  @override
  final double widgetPositionBottom;

  @override
  final EdgeInsets closeButtonPadding;

  @override
  final double borderWidth;

  @override
  final Color borderColor;

  ToastTheme({
    required this.messageTextStyle,
    required this.backgroundColor,
    required this.spacing,
    required this.borderRadius,
    required this.widgetWidth,
    required this.widgetHeight,
    required this.closeButtonPadding,
    required this.borderWidth,
    required this.borderColor,
    required this.widgetPositionRight,
    required this.widgetPositionBottom,
  });
}
