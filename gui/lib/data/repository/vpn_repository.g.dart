// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'vpn_repository.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(vpnRepository)
final vpnRepositoryProvider = VpnRepositoryProvider._();

final class VpnRepositoryProvider
    extends $FunctionalProvider<VpnRepository, VpnRepository, VpnRepository>
    with $Provider<VpnRepository> {
  VpnRepositoryProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'vpnRepositoryProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$vpnRepositoryHash();

  @$internal
  @override
  $ProviderElement<VpnRepository> $createElement($ProviderPointer pointer) =>
      $ProviderElement(pointer);

  @override
  VpnRepository create(Ref ref) {
    return vpnRepository(ref);
  }

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(VpnRepository value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<VpnRepository>(value),
    );
  }
}

String _$vpnRepositoryHash() => r'd811a6cab546f1cc60ce16f58648a52a1f983210';
