// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'pending_settings_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$pendingVPNProtocolHash() =>
    r'86c35594f0755112222a6d9667e38d05a0bb13e6';

/// A notifier for storing a pending VPN protocol value that awaits user confirmation.
/// Used when a VPN protocol change requires user confirmation (e.g., reconnect popup).
///
/// Copied from [PendingVPNProtocol].
@ProviderFor(PendingVPNProtocol)
final pendingVPNProtocolProvider =
    NotifierProvider<PendingVPNProtocol, VpnProtocol?>.internal(
      PendingVPNProtocol.new,
      name: r'pendingVPNProtocolProvider',
      debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
          ? null
          : _$pendingVPNProtocolHash,
      dependencies: null,
      allTransitiveDependencies: null,
    );

typedef _$PendingVPNProtocol = Notifier<VpnProtocol?>;
// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member, deprecated_member_use_from_same_package
