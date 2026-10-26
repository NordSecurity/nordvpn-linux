import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/data/providers/app_state_provider.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/data/repository/vpn_repository.dart';
import 'package:nordvpn/data/repository/vpn_settings_repository.dart';
import 'package:nordvpn/pb/daemon/servers.pb.dart'
    show RecommendedServerLocation;
import 'package:nordvpn/pb/daemon/settings.pb.dart' show Settings;
import 'package:nordvpn/pb/daemon/status.pb.dart' show StatusResponse;

class FakeVpnRepository implements VpnRepository {
  FakeVpnRepository({this.status, this.locations = const []});

  final StatusResponse? status;
  final List<RecommendedServerLocation> locations;
  int fetchCount = 0;

  @override
  Future<StatusResponse> fetchStatus() async => status ?? StatusResponse();

  @override
  Future<RecommendedServerLocation> fetchRecommendedServerLocation() async {
    final location = locations[fetchCount.clamp(0, locations.length - 1)];
    fetchCount++;
    return location;
  }

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

class FakeVpnSettingsRepository implements VpnSettingsRepository {
  FakeVpnSettingsRepository({
    this.settings,
    this.setVpnProtocolStatus = DaemonStatusCode.success,
  });

  final ApplicationSettings? settings;
  final int setVpnProtocolStatus;
  VpnProtocol? requested;

  @override
  Future<ApplicationSettings> fetchSettings() async =>
      settings ?? ApplicationSettings.fromSettings(Settings());

  @override
  Future<int> setVpnProtocol(VpnProtocol vpnProtocol) async {
    requested = vpnProtocol;
    return setVpnProtocolStatus;
  }

  // other methods not relevant
  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

class FakeAppStateChange implements AppStateChange {
  VpnSettingsObserver? observer;

  @override
  void addSettingsObserver(VpnSettingsObserver observer) =>
      this.observer = observer;

  @override
  void removeSettingsObserver(VpnSettingsObserver observer) =>
      this.observer = null;

  // other methods not relevant
  @override
  dynamic noSuchMethod(Invocation invocation) => null;
}

/// Creates a container with the fakes and disposes them on teardown.
ProviderContainer createContainer({
  VpnRepository? vpnRepository,
  VpnSettingsRepository? vpnSettingsRepository,
  AppStateChange? appState,
}) {
  final container = ProviderContainer(
    overrides: [
      vpnRepositoryProvider.overrideWithValue(
        vpnRepository ?? FakeVpnRepository(),
      ),
      vpnSettingsProvider.overrideWithValue(
        vpnSettingsRepository ?? FakeVpnSettingsRepository(),
      ),
      appStateProvider.overrideWithValue(appState ?? FakeAppStateChange()),
    ],
    retry: (retryCount, error) => null,
  );
  addTearDown(container.dispose);
  return container;
}
