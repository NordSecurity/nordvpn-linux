import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/data/providers/app_state_provider.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/data/repository/vpn_repository.dart';
import 'package:nordvpn/i18n/country_names_service.dart';
import 'package:nordvpn/pb/daemon/config/protocol.pbenum.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import '../../utils/fake_shared_preferences.dart';

final class _FakeVpnRepository implements VpnRepository {
  final StatusResponse status;

  _FakeVpnRepository(this.status);

  @override
  Future<StatusResponse> fetchStatus() async => status;

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

final class _FakeAppStateChange implements AppStateChange {
  @override
  void addVpnStatusObserver(VpnStatusObserver observer) {}

  @override
  void removeVpnStatusObserver(VpnStatusObserver observer) {}

  @override
  void addPauseEventsObserver(PauseEventsObserver observer) {}

  @override
  void removePauseEventsObserver(PauseEventsObserver observer) {}

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUpAll(() async {
    SharedPreferencesAsyncPlatform.instance = FakeSharedPreferencesAsync();
    await initServiceLocator();
    sl<CountryNamesService>().register(code: "LT", name: "Lithuania");
  });

  StatusResponse statusResponse({
    required Technology technology,
    required Protocol protocol,
  }) {
    return StatusResponse(
      state: ConnectionState.CONNECTED,
      technology: technology,
      protocol: protocol,
      ip: "127.0.0.1",
      hostname: "lt123.nordvpn.com",
      country: "Lithuania",
      city: "Vilnius",
      parameters: ConnectionParameters(source: ConnectionSource.MANUAL),
    );
  }

  final nordLynx = statusResponse(
    technology: Technology.NORDLYNX,
    protocol: Protocol.UDP,
  );
  final nordWhisper = statusResponse(
    technology: Technology.NORDWHISPER,
    protocol: Protocol.Webtunnel,
  );

  Future<ProviderContainer> buildController() async {
    final container = ProviderContainer(
      overrides: [
        vpnRepositoryProvider.overrideWithValue(_FakeVpnRepository(nordLynx)),
        appStateProvider.overrideWithValue(_FakeAppStateChange()),
      ],
    );
    addTearDown(container.dispose);
    await container.read(vpnStatusControllerProvider.future);
    return container;
  }

  test("a status change refreshes the protocol", () async {
    final container = await buildController();
    expect(
      container.read(vpnStatusControllerProvider).value!.protocol,
      VpnProtocol.nordlynx,
    );

    container
        .read(vpnStatusControllerProvider.notifier)
        .onVpnStatusChanged(nordWhisper);

    expect(
      container.read(vpnStatusControllerProvider).value!.protocol,
      VpnProtocol.nordWhisper,
    );
  });

  test("the state converges so a repeated status is ignored", () async {
    final container = await buildController();

    container
        .read(vpnStatusControllerProvider.notifier)
        .onVpnStatusChanged(nordWhisper);

    final state = container.read(vpnStatusControllerProvider).value!;
    expect(state.isEqualToStatusResponse(nordWhisper), isTrue);
  });
}
