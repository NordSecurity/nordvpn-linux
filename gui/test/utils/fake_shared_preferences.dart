import 'package:shared_preferences_platform_interface/in_memory_shared_preferences_async.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';
import 'package:shared_preferences_platform_interface/types.dart';

base class FakeSharedPreferencesAsync extends SharedPreferencesAsyncPlatform {
  final InMemorySharedPreferencesAsync backend =
      InMemorySharedPreferencesAsync.empty();

  @override
  Future<bool> clear(
    ClearPreferencesParameters parameters,
    SharedPreferencesOptions options,
  ) {
    return backend.clear(parameters, options);
  }

  @override
  Future<bool?> getBool(String key, SharedPreferencesOptions options) {
    return backend.getBool(key, options);
  }

  @override
  Future<double?> getDouble(String key, SharedPreferencesOptions options) {
    return backend.getDouble(key, options);
  }

  @override
  Future<int?> getInt(String key, SharedPreferencesOptions options) {
    return backend.getInt(key, options);
  }

  @override
  Future<Set<String>> getKeys(
    GetPreferencesParameters parameters,
    SharedPreferencesOptions options,
  ) {
    return backend.getKeys(parameters, options);
  }

  @override
  Future<Map<String, Object>> getPreferences(
    GetPreferencesParameters parameters,
    SharedPreferencesOptions options,
  ) {
    return backend.getPreferences(parameters, options);
  }

  @override
  Future<String?> getString(String key, SharedPreferencesOptions options) {
    return backend.getString(key, options);
  }

  @override
  Future<List<String>?> getStringList(
    String key,
    SharedPreferencesOptions options,
  ) {
    return backend.getStringList(key, options);
  }

  @override
  Future<bool> setBool(
    String key,
    bool value,
    SharedPreferencesOptions options,
  ) {
    return backend.setBool(key, value, options);
  }

  @override
  Future<bool> setDouble(
    String key,
    double value,
    SharedPreferencesOptions options,
  ) {
    return backend.setDouble(key, value, options);
  }

  @override
  Future<bool> setInt(String key, int value, SharedPreferencesOptions options) {
    return backend.setInt(key, value, options);
  }

  @override
  Future<bool> setString(
    String key,
    String value,
    SharedPreferencesOptions options,
  ) {
    return backend.setString(key, value, options);
  }

  @override
  Future<bool> setStringList(
    String key,
    List<String> value,
    SharedPreferencesOptions options,
  ) {
    return backend.setStringList(key, value, options);
  }
}
