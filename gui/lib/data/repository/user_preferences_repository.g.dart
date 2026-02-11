// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'user_preferences_repository.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(userPreferences)
final userPreferencesProvider = UserPreferencesProvider._();

final class UserPreferencesProvider
    extends
        $FunctionalProvider<
          UserPreferencesRepository,
          UserPreferencesRepository,
          UserPreferencesRepository
        >
    with $Provider<UserPreferencesRepository> {
  UserPreferencesProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'userPreferencesProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$userPreferencesHash();

  @$internal
  @override
  $ProviderElement<UserPreferencesRepository> $createElement(
    $ProviderPointer pointer,
  ) => $ProviderElement(pointer);

  @override
  UserPreferencesRepository create(Ref ref) {
    return userPreferences(ref);
  }

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(UserPreferencesRepository value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<UserPreferencesRepository>(value),
    );
  }
}

String _$userPreferencesHash() => r'b7d59190a6480505b86ec562a95aea140c56d867';
