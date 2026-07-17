import 'package:fixnum/fixnum.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/mocks/daemon/grpc_server.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/pb/daemon/config/group.pbenum.dart' as pb;
import 'package:nordvpn/pb/daemon/servers.pb.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';

import '../../utils/fakes.dart';

// Server list that contains only standard servers, so any Double VPN request
// finds no matching server (mirrors "no Double VPN server for this technology").
ServersResponse _standardServersOnly() => ServersResponse(
  servers: ServersMap(
    serversByCountry: [
      ServerCountry(
        countryCode: "FR",
        countryName: "France",
        cities: [
          ServerCity(
            cityName: "Paris",
            servers: [
              Server(
                id: Int64(1),
                hostName: "fr1.nordvpn.com",
                serverGroups: [pb.ServerGroup.STANDARD_VPN_SERVERS],
              ),
            ],
          ),
        ],
      ),
    ],
  ),
);

void main() {
  late ProviderContainer container;

  setUp(() async {
    await GrpcServer.instance.start();
    GrpcServer.instance.account.replaceAccount(fakeAccount());
    GrpcServer.instance.vpnStatus.delayDuration = Duration.zero;
    GrpcServer.instance.serversList.setServersList = _standardServersOnly();

    container = ProviderContainer();
    addTearDown(() async {
      container.dispose();
      await GrpcServer.instance.stop();
    });

    // wait for the controller to finish its initial build (fetchStatus)
    await container.read(vpnStatusControllerProvider.future);
  });

  test(
    "reconnect falls back to quick connect when no suitable server is found",
    () async {
      await container
          .read(vpnStatusControllerProvider.notifier)
          .reconnectOrQuickConnect(
            ConnectionParameters(group: pb.ServerGroup.DOUBLE_VPN),
          );

      // The "server not available" (3032) result of the reconnect must NOT be
      // surfaced; instead the quick connect fallback should have run.
      expect(
        container.read(popupsProvider)?.id,
        isNot(DaemonStatusCode.serverUnavailable),
      );
      // The quick connect for the current technology connected to a server.
      expect(
        GrpcServer.instance.vpnStatus.status.state,
        ConnectionState.CONNECTED,
      );
    },
  );

  test("reconnect connects normally when a suitable server exists", () async {
    await container
        .read(vpnStatusControllerProvider.notifier)
        .reconnectOrQuickConnect(ConnectionParameters(countryCode: "fr"));

    expect(container.read(popupsProvider), isNull);
    expect(
      GrpcServer.instance.vpnStatus.status.state,
      ConnectionState.CONNECTED,
    );
  });

  test(
    "reconnect surfaces non-server-unavailable failures without falling back",
    () async {
      GrpcServer.instance.vpnStatus.connectingErrorStatusCode =
          DaemonStatusCode.failedToConnectToVpn;

      await container
          .read(vpnStatusControllerProvider.notifier)
          .reconnectOrQuickConnect(
            ConnectionParameters(group: pb.ServerGroup.DOUBLE_VPN),
          );

      // The failure is shown as-is and no quick connect fallback happens.
      expect(
        container.read(popupsProvider)?.id,
        DaemonStatusCode.failedToConnectToVpn,
      );
      expect(
        GrpcServer.instance.vpnStatus.status.state,
        isNot(ConnectionState.CONNECTED),
      );
    },
  );
}
