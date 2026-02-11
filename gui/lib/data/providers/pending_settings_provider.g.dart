// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'pending_settings_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning
/// A notifier for storing a pending VPN protocol value that awaits user confirmation.
/// Used when a VPN protocol change requires user confirmation (e.g., reconnect popup).

@ProviderFor(PendingVPNProtocol)
final pendingVPNProtocolProvider = PendingVPNProtocolProvider._();

/// A notifier for storing a pending VPN protocol value that awaits user confirmation.
/// Used when a VPN protocol change requires user confirmation (e.g., reconnect popup).
final class PendingVPNProtocolProvider
    extends $NotifierProvider<PendingVPNProtocol, VpnProtocol?> {
  /// A notifier for storing a pending VPN protocol value that awaits user confirmation.
  /// Used when a VPN protocol change requires user confirmation (e.g., reconnect popup).
  PendingVPNProtocolProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'pendingVPNProtocolProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$pendingVPNProtocolHash();

  @$internal
  @override
  PendingVPNProtocol create() => PendingVPNProtocol();

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(VpnProtocol? value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<VpnProtocol?>(value),
    );
  }
}

String _$pendingVPNProtocolHash() =>
    r'86c35594f0755112222a6d9667e38d05a0bb13e6';

/// A notifier for storing a pending VPN protocol value that awaits user confirmation.
/// Used when a VPN protocol change requires user confirmation (e.g., reconnect popup).

abstract class _$PendingVPNProtocol extends $Notifier<VpnProtocol?> {
  VpnProtocol? build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref = this.ref as $Ref<VpnProtocol?, VpnProtocol?>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<VpnProtocol?, VpnProtocol?>,
              VpnProtocol?,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
