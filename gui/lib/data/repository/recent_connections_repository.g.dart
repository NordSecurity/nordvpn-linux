// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'recent_connections_repository.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(recentConnectionsRepository)
final recentConnectionsRepositoryProvider =
    RecentConnectionsRepositoryProvider._();

final class RecentConnectionsRepositoryProvider
    extends
        $FunctionalProvider<
          RecentConnectionsRepository,
          RecentConnectionsRepository,
          RecentConnectionsRepository
        >
    with $Provider<RecentConnectionsRepository> {
  RecentConnectionsRepositoryProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'recentConnectionsRepositoryProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$recentConnectionsRepositoryHash();

  @$internal
  @override
  $ProviderElement<RecentConnectionsRepository> $createElement(
    $ProviderPointer pointer,
  ) => $ProviderElement(pointer);

  @override
  RecentConnectionsRepository create(Ref ref) {
    return recentConnectionsRepository(ref);
  }

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(RecentConnectionsRepository value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<RecentConnectionsRepository>(value),
    );
  }
}

String _$recentConnectionsRepositoryHash() =>
    r'69afdeb0fbd89f718eedb23cf998f7a7726f1f8b';
