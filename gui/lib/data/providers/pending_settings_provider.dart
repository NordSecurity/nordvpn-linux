import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'pending_settings_provider.g.dart';

/// A notifier for storing a pending VPN protocol value that awaits user confirmation.
/// Used when a VPN protocol change requires user confirmation (e.g., reconnect popup).
@Riverpod(keepAlive: true)
class PendingVPNProtocol extends _$PendingVPNProtocol {
  @override
  VpnProtocol? build() => null;

  /// Sets the pending value.
  void set(VpnProtocol value) {
    state = value;
  }

  /// Clears the pending value.
  void clear() {
    state = null;
  }

  /// Returns the pending value and clears it.
  /// Use this when applying the pending change.
  VpnProtocol? consume() {
    final value = state;
    clear();
    return value;
  }
}
