import 'package:flutter/material.dart';
import 'package:nordvpn/theme/ux_colors.dart';

final class UXFonts {
  final UXColors uxColors;

  UXFonts(ThemeMode mode) : uxColors = UXColors(mode);

  TextStyle get caption => TextStyle(
    fontSize: 12,
    color: uxColors.textSecondary,
    fontWeight: FontWeight.w400,
  );

  TextStyle get captionTransparent_60 => TextStyle(
    fontSize: 12,
    color: uxColors.textSecondary,
    fontWeight: FontWeight.w400,
  );

  TextStyle get body => TextStyle(
    fontSize: 14,
    color: uxColors.textPrimary,
    fontWeight: FontWeight.w400,
  );

  TextStyle get captionStrong => TextStyle(
    fontSize: 12,
    color: uxColors.textPrimary,
    fontWeight: FontWeight.w600,
  );

  TextStyle get bodyStrong => TextStyle(
    fontSize: 14,
    fontWeight: FontWeight.w500,
    color: uxColors.textPrimary,
  );

  TextStyle get title => TextStyle(
    fontSize: 24,
    fontWeight: FontWeight.w700,
    color: uxColors.textPrimary,
  );

  TextStyle get bodyCaution => TextStyle(
    fontSize: 14,
    color: uxColors.textCaution,
    fontWeight: FontWeight.w400,
  );

  TextStyle get subtitle => TextStyle(
    fontSize: 18,
    color: uxColors.textPrimary,
    fontWeight: FontWeight.w600,
  );

  TextStyle get textDisabled => TextStyle(
    fontSize: 14,
    color: uxColors.textDisabled,
    fontWeight: FontWeight.w400,
  );

  TextStyle get linkNormal => TextStyle(
    fontSize: 14,
    color: uxColors.fillAccentPrimary,
    fontWeight: FontWeight.w400,
  );

  TextStyle get linkSmall => TextStyle(
    fontSize: 12,
    color: uxColors.fillAccentPrimary,
    fontWeight: FontWeight.w400,
  );
}
