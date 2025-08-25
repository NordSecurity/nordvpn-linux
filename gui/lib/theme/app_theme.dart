import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'app_theme.tailor.dart';

// This class will contain all the other custom theme elements that cannot
// be set into the ThemeData
@tailorMixin
final class AppTheme extends ThemeExtension<AppTheme>
    with _$AppThemeTailorMixin {
  AppTheme({
    required this.borderRadiusLarge,
    required this.borderRadiusMedium,
    required this.borderRadiusSmall,
    required this.padding,
    required this.margin,
    required this.outerPadding,
    required this.borderColor,
    required this.verticalSpaceSmall,
    required this.verticalSpaceMedium,
    required this.verticalSpaceLarge,
    required this.horizontalSpaceSmall,
    required this.horizontalSpace,
    required this.textErrorColor,
    required this.successColor,
    required this.flagsBorderSize,
    required this.overlayBackgroundColor,
    required this.caption,
    required this.captionStrong,
    required this.bodyStrong,
    required this.subtitleStrong,
    required this.body,
    required this.captionRegularGray171,
    required this.linkButton,
    required this.title,
    required this.trailingIconSize,
    required this.backgroundColor,
    required this.areaBackgroundColor,
    required this.dividerColor,
    required this.disabledOpacity,
    required this.linkNormal,
    required this.linkSmall,
    required this.textDisabled,
    required this.area,
  });

  @override
  final double borderRadiusLarge;

  @override
  final double borderRadiusMedium;

  @override
  final double borderRadiusSmall;

  @override
  final double padding;

  @override
  final double margin;

  @override
  final double outerPadding;

  @override
  final Color borderColor;

  @override
  final double verticalSpaceSmall;

  @override
  final double verticalSpaceMedium;

  @override
  final double verticalSpaceLarge;

  @override
  final double horizontalSpaceSmall;

  @override
  final double horizontalSpace;

  @override
  final Color textErrorColor;

  @override
  final Color successColor;

  @override
  final double flagsBorderSize;

  @override
  final Color overlayBackgroundColor;

  @override
  final double trailingIconSize;

  @override
  final Color backgroundColor;

  @override
  final Color areaBackgroundColor;

  @override
  final Color dividerColor;

  @override
  final double disabledOpacity;
  // fonts
  @override
  final TextStyle captionStrong;

  @override
  final TextStyle caption;

  @override
  final TextStyle captionRegularGray171;

  @override
  final TextStyle bodyStrong;

  @override
  final TextStyle body;

  @override
  final TextStyle subtitleStrong;

  @override
  final TextStyle linkButton;

  @override
  final TextStyle title;

  @override
  final TextStyle linkNormal;

  @override
  final TextStyle linkSmall;

  @override
  final TextStyle textDisabled;

  // background color for area containing some elements,
  // like for [AutoconnectPanel]
  @override
  final Color area;
}
