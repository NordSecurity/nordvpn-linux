import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'nav_rail_theme.tailor.dart';

@tailorMixin
final class NavRailTheme extends ThemeExtension<NavRailTheme>
    with _$NavRailThemeTailorMixin {
  @override
  final Color railBg;

  @override
  final double railWidth;

  @override
  final double containerWidth;

  @override
  final double containerHeight;

  @override
  final double betweenIconsGap;

  @override
  final double iconsPaddingTop;

  @override
  final double iconsMargin;

  @override
  final BorderRadius radius;

  @override
  final Color selectedItemBg;

  NavRailTheme({
    required this.railBg,
    required this.railWidth,
    required this.containerWidth,
    required this.containerHeight,
    required this.betweenIconsGap,
    required this.iconsPaddingTop,
    required this.iconsMargin,
    required this.radius,
    required this.selectedItemBg,
  });
}
