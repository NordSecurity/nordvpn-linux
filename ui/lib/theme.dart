import 'package:flutter/material.dart';

enum AppTheme {
  light,
  dark,
}

class ThemeManager {
  ThemeManager._();

  static AppTheme _currentTheme = AppTheme.light;
  static void toggleTheme() {
    _currentTheme =
        _currentTheme == AppTheme.light ? AppTheme.dark : AppTheme.light;
  }

  static Color get navBarBackgroundColor {
    return _currentTheme == AppTheme.light
        ? const Color.fromRGBO(240, 240, 240, 1)
        : const Color.fromRGBO(240, 240, 240, 1);
  }

  static Color get navBarSelectedItemBgColor {
    return _currentTheme == AppTheme.light
        ? const Color.fromRGBO(0, 0, 0, 0.1)
        : const Color.fromRGBO(0, 0, 0, 0.1);
  }

  static Color get appHeaderBgColor {
    return navBarBackgroundColor;
  }

  static double get appHeaderHeight => 70;

  static Color get accentColor {
    return _currentTheme == AppTheme.light ? Colors.blue : Colors.amber;
  }

  static Color get accentTextColor {
    return _currentTheme == AppTheme.light ? Colors.white : Colors.amber;
  }

  static Color get borderColor {
    return _currentTheme == AppTheme.light ? Colors.black26 : Colors.white10;
  }

  static double get borderRadius => 8;
}
