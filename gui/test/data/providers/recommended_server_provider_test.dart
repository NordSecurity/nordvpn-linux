import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/providers/recommended_server_provider.dart';
import 'package:nordvpn/pb/daemon/config/vpn_protocol.pbenum.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
// servers.pb.dart exports a different Technology (the per-server one), so hide it here
import 'package:nordvpn/pb/daemon/servers.pb.dart' hide Technology;

import '../../utils/provider_fakes.dart';

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

  ApplicationSettings settingsFor(VPNProtocol vpnProtocol) {
    return ApplicationSettings.fromSettings(Settings(vpnProtocol: vpnProtocol));
  }

  // Builds the provider and returns it alongside the fakes it was given.
  Future<(RecommendedServer, FakeVpnRepository)> build({
    required ApplicationSettings initialSettings,
  }) async {
    final vpnRepository = FakeVpnRepository(locations: [dallas, bucharest]);
    final appState = FakeAppStateChange();
    final container = createContainer(
      vpnRepository: vpnRepository,
      vpnSettingsRepository: FakeVpnSettingsRepository(
        settings: initialSettings,
      ),
      appState: appState,
    );

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
      initialSettings: settingsFor(VPNProtocol.VPN_PROTOCOL_NORDLYNX),
    );

    expect(vpnRepository.fetchCount, 1, reason: "the initial fetch on build");
    expect(notifier.state.value, dallas);

    // The very first change the observer ever sees must still be acted on. Before the
    // baseline was seeded in build(), this change only filled it in and was swallowed.
    await notifier.onSettingsChanged(
      settingsFor(VPNProtocol.VPN_PROTOCOL_NORDWHISPER),
    );

    expect(vpnRepository.fetchCount, 2, reason: "the protocol changed");
    expect(notifier.state.value, bucharest);
  });

  test(
    "a settings change that leaves the protocol alone does not refetch",
    () async {
      final (notifier, vpnRepository) = await build(
        initialSettings: settingsFor(VPNProtocol.VPN_PROTOCOL_NORDLYNX),
      );

      await notifier.onSettingsChanged(
        settingsFor(VPNProtocol.VPN_PROTOCOL_NORDLYNX),
      );

      expect(vpnRepository.fetchCount, 1);
      expect(notifier.state.value, dallas);
    },
  );
}
