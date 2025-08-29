import 'package:flutter/material.dart';
import 'package:nordvpn/data/models/user_preferences.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/data/repository/user_preferences_repository.dart';
import 'package:nordvpn/logger.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'preferences_controller.g.dart';

typedef _SetCallback = Future<void> Function(UserPreferencesRepository repo);
typedef _GetCallback<T> = Future<T> Function(UserPreferencesRepository repo);

// Controls local, GUI-specific settings.
@riverpod
class PreferencesController extends _$PreferencesController {
  @override
  Future<UserPreferences> build() async {
    final repo = ref.watch(userPreferencesProvider);
    return await repo.loadPreferences();
  }

  Future<void> setAppearance(ThemeMode value) async {
    if (!state.hasValue) {
      logger.e("no state value when setting appearance");
      return;
    }
    if (!await _setValue((repo) => repo.setAppearance(value))) return;
    state = AsyncData(state.value!.copyWith(appearance: value));
  }

  Future<void> resetToDefaults() async {
    if (!await _setValue((repo) => repo.reset())) return;
    final preferences = await _getValue((repo) => repo.loadPreferences());
    if (preferences == null) return;
    state = AsyncData(preferences);
  }

  Future<bool> _setValue(_SetCallback callback) async {
    // Not a gRPC repository
    // This error handling should be good enough
    final repo = ref.read(userPreferencesProvider);
    try {
      await callback(repo);
      return true;
    } catch (e) {
      logger.e("failed to set preferences: $e");
      ref.read(popupsProvider.notifier).show(DaemonStatusCode.failure);
    }
    return false;
  }

  Future<T?> _getValue<T>(_GetCallback<T> callback) async {
    // Not a gRPC repository
    // This error handling should be good enough
    final repo = ref.read(userPreferencesProvider);
    try {
      return await callback(repo);
    } catch (e) {
      logger.e("failed to get preferences: $e");
      ref.read(popupsProvider.notifier).show(DaemonStatusCode.failure);
    }
    return null;
  }
}
