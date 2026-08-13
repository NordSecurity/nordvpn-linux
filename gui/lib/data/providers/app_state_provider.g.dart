// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'app_state_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(appState)
final appStateProvider = AppStateProvider._();

final class AppStateProvider
    extends $FunctionalProvider<AppStateChange, AppStateChange, AppStateChange>
    with $Provider<AppStateChange> {
  AppStateProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'appStateProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$appStateHash();

  @$internal
  @override
  $ProviderElement<AppStateChange> $createElement($ProviderPointer pointer) =>
      $ProviderElement(pointer);

  @override
  AppStateChange create(Ref ref) {
    return appState(ref);
  }

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(AppStateChange value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<AppStateChange>(value),
    );
  }
}

String _$appStateHash() => r'b54466029523e85ecb03e71bffa6a92a35fb767c';
