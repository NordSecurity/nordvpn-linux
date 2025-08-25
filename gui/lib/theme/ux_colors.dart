import 'package:flutter/material.dart';

// UX defined colors
final class UXColors {
  final ThemeMode mode;

  UXColors(this.mode);

  bool get _isLight => mode == ThemeMode.light;

  // ============================ [ text ] ============================

  Color get textPrimary => _isLight
      ? Color.fromRGBO(27, 27, 27, 1.0)
      : Color.fromRGBO(255, 255, 255, 1.0);

  Color get textSecondary => _isLight
      ? Color.fromRGBO(94, 94, 94, 1.0)
      : Color.fromRGBO(207, 207, 207, 1.0);

  Color get textDisabled => _isLight
      ? Color.fromRGBO(27, 27, 27, 0.36)
      : Color.fromRGBO(255, 255, 255, 0.36);

  Color get textOnAccent => _isLight
      ? Color.fromRGBO(255, 255, 255, 1.0)
      : Color.fromRGBO(255, 255, 255, 1.0);

  Color get textSuccess => _isLight
      ? Color.fromRGBO(6, 121, 72, 1.0)
      : Color.fromRGBO(51, 204, 112, 1.0);

  Color get textCaution => _isLight
      ? Color.fromRGBO(191, 34, 19, 1.0)
      : Color.fromRGBO(255, 120, 107, 1.0);

  Color get textAccentPrimary => _isLight
      ? Color.fromRGBO(62, 95, 255, 1.0)
      : Color.fromRGBO(97, 124, 255, 1.0);

  // ============================ [ icon ] ============================

  Color get iconPrimary => _isLight
      ? Color.fromRGBO(27, 27, 27, 1)
      : Color.fromRGBO(255, 255, 255, 1.0);

  // ============================ [ fill ] ============================

  Color get fillWhiteFixed => Color.fromRGBO(255, 255, 255, 1.0);

  Color get fillAccentPrimary => _isLight
      ? Color.fromRGBO(62, 95, 255, 1.0)
      : Color.fromRGBO(62, 95, 255, 1.0);

  Color get fillPrimaryGrey => _isLight
      ? Color.fromRGBO(0, 0, 0, 0.44)
      : Color.fromRGBO(255, 255, 255, 0.44);

  Color get fillGreyPrimary => _isLight
      ? Color.fromRGBO(255, 255, 255, 1.0)
      : Color.fromRGBO(27, 27, 27, 1.0);

  Color get fillGreySecondary => _isLight
      ? Color.fromRGBO(111, 111, 111, 0.16)
      : Color.fromRGBO(117, 117, 117, 0.25);

  Color get fillGreyTertiary => _isLight
      ? Color.fromRGBO(111, 111, 111, 0.06)
      : Color.fromRGBO(117, 117, 117, 0.8);

  Color get fillGreyDisabled => _isLight
      ? Color.fromRGBO(27, 27, 27, 0.2)
      : Color.fromRGBO(255, 255, 255, 0.15);

  Color get fillGreyQuaternary => _isLight
      ? Color.fromRGBO(111, 111, 111, 0.02)
      : Color.fromRGBO(117, 117, 117, 0.04);

  // ============================ [ background ] ============================

  Color get backgroundPrimary => _isLight
      ? Color.fromRGBO(255, 255, 255, 1.0)
      : Color.fromRGBO(27, 27, 27, 1.0);

  Color get backgroundSecondary => _isLight
      ? Color.fromRGBO(245, 245, 245, 1.0)
      : Color.fromRGBO(41, 41, 41, 1.0);

  Color get backgroundOverlay => _isLight
      ? Color.fromRGBO(27, 27, 27, 0.5)
      : Color.fromRGBO(51, 51, 51, 0.5);

  // ============================ [ stroke ] ============================

  Color get strokeMedium => _isLight
      ? Color.fromRGBO(225, 225, 225, 1.0)
      : Color.fromRGBO(237, 237, 237, 0.1);

  Color get strokeSoft => _isLight
      ? Color.fromRGBO(237, 237, 237, 1.0)
      : Color.fromRGBO(45, 45, 45, 1.0);

  Color get strokeControlPrimary => _isLight
      ? Color.fromRGBO(27, 27, 27, 1.0)
      : Color.fromRGBO(255, 255, 255, 1.0);

  Color get strokeDivider => _isLight
      ? Color.fromRGBO(227, 227, 227, 0.06)
      : Color.fromRGBO(111, 111, 111, 0.07);

  Color get strokeAccent => _isLight
      ? Color.fromRGBO(62, 95, 255, 1.0)
      : Color.fromRGBO(62, 95, 255, 1.0);

  Color get strokeCaution => _isLight
      ? Color.fromRGBO(222, 39, 23, 1.0)
      : Color.fromRGBO(255, 120, 170, 1.0);

  Color get strokeDisabled => _isLight
      ? Color.fromRGBO(27, 27, 27, 0.2)
      : Color.fromRGBO(255, 255, 255, 0.15);
}
