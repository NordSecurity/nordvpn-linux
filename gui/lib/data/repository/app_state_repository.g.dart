// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'app_state_repository.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(appStateRepository)
final appStateRepositoryProvider = AppStateRepositoryProvider._();

final class AppStateRepositoryProvider
    extends
        $FunctionalProvider<
          AppStateRepository,
          AppStateRepository,
          AppStateRepository
        >
    with $Provider<AppStateRepository> {
  AppStateRepositoryProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'appStateRepositoryProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$appStateRepositoryHash();

  @$internal
  @override
  $ProviderElement<AppStateRepository> $createElement(
    $ProviderPointer pointer,
  ) => $ProviderElement(pointer);

  @override
  AppStateRepository create(Ref ref) {
    return appStateRepository(ref);
  }

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(AppStateRepository value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<AppStateRepository>(value),
    );
  }
}

String _$appStateRepositoryHash() =>
    r'c9e92c240bdb361eaeac15f5c7361055d9e4af6d';
