import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/providers/app_state_provider.dart';
import 'package:nordvpn/data/providers/recommended_server_provider.dart';
import 'package:nordvpn/data/repository/vpn_repository.dart';
import 'package:nordvpn/data/repository/vpn_settings_repository.dart';
import 'package:nordvpn/pb/daemon/config/protocol.pbenum.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
// servers.pb.dart exports a different Technology (the per-server one), so hide it here
import 'package:nordvpn/pb/daemon/servers.pb.dart' hide Technology;

final class _FakeVpnRepository implements VpnRepository {
  final List<RecommendedServerLocation> locations;
  int fetchCount = 0;

  _FakeVpnRepository(this.locations);

  @override
  Future<RecommendedServerLocation> fetchRecommendedServerLocation() async {
    final location = locations[fetchCount.clamp(0, locations.length - 1)];
    fetchCount++;
    return location;
  }

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

final class _FakeVpnSettingsRepository implements VpnSettingsRepository {
  final ApplicationSettings settings;

  _FakeVpnSettingsRepository(this.settings);

  @override
  Future<ApplicationSettings> fetchSettings() async => settings;

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

final class _FakeAppStateChange implements AppStateChange {
  VpnSettingsObserver? observer;

  @override
  void addSettingsObserver(VpnSettingsObserver observer) {
    this.observer = observer;
  }

  @override
  void removeSettingsObserver(VpnSettingsObserver observer) {
    this.observer = null;
  }

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

void main() {
  final dallas = RecommendedServerLocation(
    countryCode: "US",
    countryName: "United States",
    cityName: "Dallas",
  );
  final bucharest = RecommendedServerLocation(
    countryCode: "RO",
    countryName: "Romania",
    cityName: "Bucharest",
  );

  ApplicationSettings settingsFor(Technology technology, Protocol protocol) {
    return ApplicationSettings.fromSettings(
      Settings(technology: technology, protocol: protocol),
    );
  }

  // Builds the provider and returns it alongside the fakes it was given.
  Future<(RecommendedServer, _FakeVpnRepository)> build({
    required ApplicationSettings initialSettings,
  }) async {
    final vpnRepository = _FakeVpnRepository([dallas, bucharest]);
    final appState = _FakeAppStateChange();
    final container = ProviderContainer(
      overrides: [
        vpnRepositoryProvider.overrideWithValue(vpnRepository),
        vpnSettingsProvider.overrideWithValue(
          _FakeVpnSettingsRepository(initialSettings),
        ),
        appStateProvider.overrideWithValue(appState),
      ],
    );
    addTearDown(container.dispose);

    await container.read(recommendedServerProvider.future);
    final notifier = container.read(recommendedServerProvider.notifier);

    expect(
      appState.observer,
      same(notifier),
      reason: "the provider registers itself as a settings observer",
    );

    return (notifier, vpnRepository);
  }

  test("the first settings change after startup refetches the location", () async {
    final (notifier, vpnRepository) = await build(
      initialSettings: settingsFor(Technology.NORDLYNX, Protocol.UDP),
    );

    expect(vpnRepository.fetchCount, 1, reason: "the initial fetch on build");
    expect(notifier.state.value, dallas);

    // The very first change the observer ever sees must still be acted on. Before the
    // baseline was seeded in build(), this change only filled it in and was swallowed.
    await notifier.onSettingsChanged(
      settingsFor(Technology.NORDWHISPER, Protocol.Webtunnel),
    );

    expect(vpnRepository.fetchCount, 2, reason: "the protocol changed");
    expect(notifier.state.value, bucharest);
  });

  test(
    "a settings change that leaves the protocol alone does not refetch",
    () async {
      final (notifier, vpnRepository) = await build(
        initialSettings: settingsFor(Technology.NORDLYNX, Protocol.UDP),
      );

      await notifier.onSettingsChanged(
        settingsFor(Technology.NORDLYNX, Protocol.UDP),
      );

      expect(vpnRepository.fetchCount, 1);
      expect(notifier.state.value, dallas);
    },
  );
}
