import 'package:flutter/material.dart';
import 'package:theme_tailor_annotation/theme_tailor_annotation.dart';

part 'custom_dns_theme.tailor.dart';

@tailorMixin
final class CustomDnsTheme extends ThemeExtension<CustomDnsTheme>
    with _$CustomDnsThemeTailorMixin {
  @override
  final Color formBackground;
  @override
  final double dnsInputWidth;
  @override
  final Color dividerColor;

  CustomDnsTheme({
    required this.formBackground,
    required this.dnsInputWidth,
    required this.dividerColor,
  });
}
