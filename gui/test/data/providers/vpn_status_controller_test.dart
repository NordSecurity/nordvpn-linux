import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/i18n/country_names_service.dart';
import 'package:nordvpn/pb/daemon/config/vpn_protocol.pbenum.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import '../../utils/fake_shared_preferences.dart';
import '../../utils/provider_fakes.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUpAll(() async {
    SharedPreferencesAsyncPlatform.instance = FakeSharedPreferencesAsync();
    await initServiceLocator();
    sl<CountryNamesService>().register(code: "LT", name: "Lithuania");
  });

  StatusResponse statusResponse({required VPNProtocol vpnProtocol}) {
    return StatusResponse(
      state: ConnectionState.CONNECTED,
      vpnProtocol: vpnProtocol,
      ip: "127.0.0.1",
      hostname: "lt123.nordvpn.com",
      country: "Lithuania",
      city: "Vilnius",
      parameters: ConnectionParameters(source: ConnectionSource.MANUAL),
    );
  }

  final nordLynx = statusResponse(
    vpnProtocol: VPNProtocol.VPN_PROTOCOL_NORDLYNX,
  );
  final nordWhisper = statusResponse(
    vpnProtocol: VPNProtocol.VPN_PROTOCOL_NORDWHISPER,
  );

  Future<ProviderContainer> buildController() async {
    final container = createContainer(
      vpnRepository: FakeVpnRepository(status: nordLynx),
    );
    await container.read(vpnStatusControllerProvider.future);
    return container;
  }

  test("the state converges so a repeated status is ignored", () async {
    final container = await buildController();

    container
        .read(vpnStatusControllerProvider.notifier)
        .onVpnStatusChanged(nordWhisper);

    final state = container.read(vpnStatusControllerProvider).value!;
    expect(state.isEqualToStatusResponse(nordWhisper), isTrue);
  });
}
