// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'uievent_repository.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(uiEventRepository)
const uiEventRepositoryProvider = UiEventRepositoryProvider._();

final class UiEventRepositoryProvider
    extends
        $FunctionalProvider<
          UiEventRepository,
          UiEventRepository,
          UiEventRepository
        >
    with $Provider<UiEventRepository> {
  const UiEventRepositoryProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'uiEventRepositoryProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$uiEventRepositoryHash();

  @$internal
  @override
  $ProviderElement<UiEventRepository> $createElement(
    $ProviderPointer pointer,
  ) => $ProviderElement(pointer);

  @override
  UiEventRepository create(Ref ref) {
    return uiEventRepository(ref);
  }

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(UiEventRepository value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<UiEventRepository>(value),
    );
  }
}

String _$uiEventRepositoryHash() => r'fd5d8144d24a17b2ece0ade727a5420d9b55cd13';
