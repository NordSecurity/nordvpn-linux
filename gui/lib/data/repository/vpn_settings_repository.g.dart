// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'vpn_settings_repository.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(vpnSettings)
final vpnSettingsProvider = VpnSettingsProvider._();

final class VpnSettingsProvider
    extends
        $FunctionalProvider<
          VpnSettingsRepository,
          VpnSettingsRepository,
          VpnSettingsRepository
        >
    with $Provider<VpnSettingsRepository> {
  VpnSettingsProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'vpnSettingsProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$vpnSettingsHash();

  @$internal
  @override
  $ProviderElement<VpnSettingsRepository> $createElement(
    $ProviderPointer pointer,
  ) => $ProviderElement(pointer);

  @override
  VpnSettingsRepository create(Ref ref) {
    return vpnSettings(ref);
  }

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(VpnSettingsRepository value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<VpnSettingsRepository>(value),
    );
  }
}

String _$vpnSettingsHash() => r'c4ee594879dba1d6bad0717c5eca821b255a38d6';
