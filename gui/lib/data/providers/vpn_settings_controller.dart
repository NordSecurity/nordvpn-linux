import 'package:nordvpn/data/models/allow_list.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/data/providers/app_state_provider.dart';
import 'package:nordvpn/data/providers/pending_settings_provider.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/data/repository/vpn_settings_repository.dart';
import 'package:nordvpn/internal/popup_codes.dart';
import 'package:nordvpn/logger.dart';
import 'package:grpc/grpc.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'vpn_settings_controller.g.dart';

// Don't show popups for those codes:
// - success
// - dnsListModified - error that happens when enabling Threat Protection
//   and Custom DNS is set, if user allows resetting Custom DNS, this "error"
//   will happen
// - tpLiteDisabled - error that happens when enabling Custom DNS and Threat
//   Protection is enabled, if user allows disabling TP, this "error" will
//   happen
const _popupIgnoreCodes = [
  DaemonStatusCode.success,
  DaemonStatusCode.dnsListModified,
  DaemonStatusCode.tpLiteDisabled,
  DaemonStatusCode.allowListModified,
];

@riverpod
class VpnSettingsController extends _$VpnSettingsController
    implements VpnSettingsObserver {
  @override
  FutureOr<ApplicationSettings> build() async {
    _registerNotifications();
    return await _fetchSettings();
  }

  /// Sets the VPN protocol.
  /// If VPN is connected, stores the pending protocol and shows a popup.
  /// The change will only be applied if user confirms in the popup.
  /// Otherwise sets it immediately.
  Future<int> setVpnProtocol(VpnProtocol protocol) async {
    // We need to check VPN status here because protocol change requires
    // a different flow: store pending protocol and show confirmation popup.
    // The daemon doesn't return vpnIsRunning for protocol changes.
    final vpnStatus = ref.read(vpnStatusControllerProvider).asData?.value;
    if (vpnStatus != null && vpnStatus.isConnected()) {
      // VPN is connected - store pending protocol and show popup
      // The change will be applied only if user confirms
      ref.read(pendingVPNProtocolProvider.notifier).set(protocol);
      ref
          .read(popupsProvider.notifier)
          .show(PopupCodes.reconnectToChangeProtocol);
      return DaemonStatusCode.success;
    }

    // VPN is not connected - apply the change directly
    return await _setValue((repository) => repository.setVpnProtocol(protocol));
  }

  /// Applies the pending protocol change.
  /// Returns true if the protocol was applied successfully.
  Future<bool> applyPendingVPNProtocol() async {
    // Consume the pending protocol change
    final pendingVPNProtocol = ref
        .read(pendingVPNProtocolProvider.notifier)
        .consume();
    if (pendingVPNProtocol == null) {
      logger.e('Cannot apply protocol change: no pending vpn protocol');
      return false;
    }

    // Apply the protocol change
    final status = await _setValue(
      (repository) => repository.setVpnProtocol(pendingVPNProtocol),
      popupCodeOverrides: {
        // virtualLocationsDisabled needs to show a specific popup.
        DaemonStatusCode.virtualLocationsDisabled:
            PopupCodes.reconnectToChangeVirtualLocation,
        // Ignore vpnIsRunning here - we already showed the reconnect popup
        DaemonStatusCode.vpnIsRunning: DaemonStatusCode.success,
      },
    );
    // Accept success, nothingToDo, or vpnIsRunning as valid statuses
    if (status != DaemonStatusCode.success &&
        status != DaemonStatusCode.nothingToDo &&
        status != DaemonStatusCode.vpnIsRunning) {
      logger.e('Failed to apply protocol change: $status');
      return false;
    }

    return true;
  }

  Future<int> resetToDefaults() async {
    return await ref.read(vpnSettingsProvider).resetToDefaults();
  }

  Future<int> setObfuscated(bool value) async {
    return await _setValue(
      (repository) => repository.setObfuscated(value),
      popupCodeOverrides: {
        DaemonStatusCode.vpnIsRunning: PopupCodes.reconnectToChangeObfuscation,
      },
    );
  }

  Future<int> setAnalytics(bool value) async {
    return await _setValue((repository) => repository.setAnalytics(value));
  }

  Future<int> setFirewall(bool value) async {
    return await _setValue((repository) => repository.setFirewall(value));
  }

  Future<int> setNotifications(bool value) async {
    return await _setValue((repository) => repository.setNotifications(value));
  }

  Future<int> setFirewallMark(int value) async {
    final status = await _setValue(
      (repository) => repository.setFirewallMark(value),
    );
    if (status == DaemonStatusCode.success) {
      ref
          .read(popupsProvider.notifier)
          .show(DaemonStatusCode.restartDaemonRequiredForFwMark);
    }

    return status;
  }

  Future<int> setKillSwitch(bool value) async {
    return await _setValue((repository) => repository.setKillSwitch(value));
  }

  void setAllowList(bool value) async {
    if (!state.hasValue) {
      logger.e("no state value when setting Allow List");
      return;
    }
    state = AsyncData(state.value!.copyWith(allowList: value));
  }

  Future<int> addToAllowList({PortInterval? port, Subnet? subnet}) async {
    assert((port != null) || (subnet != null), " port or subnet must be valid");
    return await _setValue(
      (repository) =>
          repository.addToAllowList(port: port, subnet: subnet?.value),
    );
  }

  Future<int> removeFromAllowList({PortInterval? port, Subnet? subnet}) async {
    assert((port != null) || (subnet != null), " port or subnet must be valid");
    return await _setValue(
      (repository) =>
          repository.removeFromAllowList(port: port, subnet: subnet?.value),
    );
  }

  Future<int> disableAllowList() async {
    final res = await _setValue((repository) => repository.disableAllowList());
    if (res == DaemonStatusCode.success) {
      setAllowList(false);
    }
    return res;
  }

  Future<int> setThreatProtection(bool value) async {
    return await _setValue(
      (repository) => repository.setThreatProtection(value),
    );
  }

  void setCustomDns(bool value) {
    if (!state.hasValue) {
      logger.e("no state value when setting custom DNS");
      return;
    }
    state = AsyncData(state.value!.copyWith(customDns: value));
  }

  Future<int> setCustomDnsServers(List<String> dnsServers) async {
    return await _setValue((repository) => repository.setDns(dnsServers));
  }

  Future<int> addCustomDns(String address) async {
    if (address.isEmpty) {
      return DaemonStatusCode.invalidDnsAddress;
    }
    List<String> dnsServers = [...?state.asData?.value.customDnsServers];
    if (dnsServers.contains(address)) {
      return DaemonStatusCode.alreadyExists;
    }

    dnsServers.add(address);
    return await setCustomDnsServers(dnsServers);
  }

  Future<int> removeCustomDns(String address) async {
    if (address.isEmpty) {
      return DaemonStatusCode.invalidDnsAddress;
    }
    final dnsServers = [...?state.asData?.value.customDnsServers];
    if (!dnsServers.contains(address)) {
      return DaemonStatusCode.nothingToDo;
    }

    dnsServers.remove(address);
    return await setCustomDnsServers(dnsServers);
  }

  Future<int> clearCustomDns() async {
    final res = await _setValue(
      (repository) => repository.setDns(List.empty()),
    );
    if (res == DaemonStatusCode.success) {
      setCustomDns(false);
    }
    return res;
  }

  Future<int> setAutoConnect(bool value, ConnectArguments? args) async {
    assert(args?.server == null, "Cannot set a server to auto-connect");

    return await _setValue(
      (repository) => repository.setAutoConnect(value, args),
    );
  }

  Future<int> setRouting(bool value) async {
    return await _setValue((repository) => repository.setRouting(value));
  }

  Future<int> setLanDiscovery(bool value) async {
    return await _setValue((repository) => repository.setLocalNetwork(value));
  }

  Future<int> setPostQuantum(bool value) async {
    return await _setValue(
      (repository) => repository.setPostQuantum(value),
      popupCodeOverrides: {
        DaemonStatusCode.vpnIsRunning: PopupCodes.reconnectToChangePostQuantum,
      },
    );
  }

  Future<int> useVirtualServers(bool value) async {
    return await _setValue(
      (repository) => repository.useVirtualServers(value),
      popupCodeOverrides: {
        DaemonStatusCode.vpnIsRunning:
            PopupCodes.reconnectToChangeVirtualLocation,
      },
    );
  }

  @override
  void onSettingsChanged(ApplicationSettings settings) {
    if (state.value == settings) {
      return;
    }
    // NOTE: The daemon only stores a list of DNS servers and does not track
    // whether Custom DNS is enabled. However, our GUI requires this information
    // to determine when to enable the Custom DNS input.
    // To work around this, we simulate the Custom DNS state in the global
    // state rather than in the widget's state. This design allows popups
    // (which have access to `WidgetRef`) to control the toggle state.
    // As a result, we prevent any state received from the daemon from
    // overwriting the simulated Custom DNS setting. If the user previously
    // enabled Custom DNS in the GUI, we ensure it remains enabled.
    // The same principle applies to Allow List.
    settings = settings.copyWith(
      customDns: state.value?.customDns ?? false,
      allowList: state.value?.allowList ?? false,
    );
    state = AsyncData(settings);
  }

  Future<ApplicationSettings> _fetchSettings() async {
    final repository = ref.read(vpnSettingsProvider);
    return await repository.fetchSettings();
  }

  /// Executes a repository callback and handles status codes and popups.
  ///
  /// The popupCodeOverrides map allows overriding daemon status codes with custom popup codes.
  /// This is useful when different settings need specific popup messages for the same daemon code.
  ///
  /// Example: Both obfuscation and post-quantum changes return vpnIsRunning, but we want
  /// different popup messages for each:
  /// ```dart
  /// _setValue(
  ///   (repo) => repo.setObfuscated(value),
  ///   popupCodeOverrides: {
  ///     DaemonStatusCode.vpnIsRunning: PopupCodes.reconnectToChangeObfuscation,
  ///   },
  /// );
  /// ```
  Future<int> _setValue(
    Future<int> Function(VpnSettingsRepository repository) callback, {
    Map<int, int>? popupCodeOverrides,
  }) async {
    final repository = ref.read(vpnSettingsProvider);
    int status = DaemonStatusCode.failure;

    try {
      status = await callback(repository);
    } on GrpcError catch (e) {
      // Convert gRPC Error into DaemonStatusCode
      status = DaemonStatusCode.fromGrpcError(e);
    } catch (e) {
      // Unexpected error during the gRPC call
      status = DaemonStatusCode.failure;
      // Log the error
      logger.e("Unexpected error: $e");
    }

    // Use overridden popup code if provided, otherwise use the daemon status code
    final popupCode = popupCodeOverrides?[status] ?? status;

    // don't show popup when code is on ignore list
    if (_popupIgnoreCodes.contains(popupCode)) {
      return status;
    }

    // We do that to avoid the toggle of on/off button in case of failure
    state = state;

    ref.read(popupsProvider.notifier).show(popupCode);

    return status;
  }

  void _registerNotifications() {
    final notification = ref.read(appStateProvider);
    notification.addSettingsObserver(this);
    ref.onDispose(() {
      notification.removeSettingsObserver(this);
    });
  }
}
