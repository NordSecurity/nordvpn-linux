import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/models/user_preferences.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:shared_preferences/shared_preferences.dart';

part 'user_preferences_repository.g.dart';

const _themeMode = "themeMode";

// Loads user preferences using `shared_preferences`.
final class UserPreferencesRepository {
  final SharedPreferencesAsync preferences;

  UserPreferencesRepository(this.preferences);

  Future<UserPreferences> loadPreferences() async {
    return UserPreferences(
      appearance: await _getThemeModeOrDefault(defaultTheme),
    );
  }

  Future<ThemeMode> _getThemeModeOrDefault(ThemeMode defaultMode) async {
    final savedTheme = await preferences.getString(_themeMode);
    final themeModeValue = savedTheme ?? defaultMode.toString();
    return themeModeValue.toThemeMode();
  }

  Future<void> setAppearance(ThemeMode appearance) async {
    await preferences.setString(_themeMode, appearance.name);
  }

  Future<void> reset() async {
    await preferences.clear();
  }
}

@Riverpod(keepAlive: true)
UserPreferencesRepository userPreferences(Ref ref) {
  return UserPreferencesRepository(sl());
}

extension _StringThemeModeExt on String {
  ThemeMode toThemeMode() {
    return ThemeMode.values.firstWhere(
      (e) => e.name == this,
      orElse: () => defaultTheme,
    );
  }
}
