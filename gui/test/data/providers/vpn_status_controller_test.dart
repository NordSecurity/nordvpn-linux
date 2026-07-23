import 'package:fixnum/fixnum.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grpc/grpc.dart';
import 'package:nordvpn/data/mocks/daemon/grpc_server.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/providers/vpn_status_controller.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/grpc/error_handling_interceptor.dart';
import 'package:nordvpn/grpc/grpc_service.dart';
import 'package:nordvpn/pb/daemon/config/group.pbenum.dart' as pb;
import 'package:nordvpn/pb/daemon/servers.pb.dart' as pb_servers;
import 'package:nordvpn/pb/daemon/status.pb.dart' as pb_status;
import 'package:nordvpn/service_locator.dart';

import '../../utils/fakes.dart';

// Server list that contains only standard servers, so any Double VPN request
// finds no matching server (mirrors "no Double VPN server for this technology").
pb_servers.ServersResponse _standardServersOnly() => pb_servers.ServersResponse(
  servers: pb_servers.ServersMap(
    serversByCountry: [
      pb_servers.ServerCountry(
        countryCode: "FR",
        countryName: "France",
        cities: [
          pb_servers.ServerCity(
            cityName: "Paris",
            servers: [
              pb_servers.Server(
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

  setUpAll(() {
    // The VPN repository builds its daemon client from the service locator,
    // so the channel and interceptor must be registered before any provider
    // is read. These tests are plain unit tests and do not run the full
    // initServiceLocator() (which also needs a widget binding for assets).
    if (!sl.isRegistered<ClientChannel>()) {
      sl.registerSingleton<ClientChannel>(createNewChannel());
    }
    if (!sl.isRegistered<ErrorHandlingInterceptor>()) {
      sl.registerSingleton(ErrorHandlingInterceptor());
    }
  });

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

    // Keep a subscription alive so the autoDispose provider is not disposed
    // mid-build (its async build() calls ref.onDispose after awaiting the
    // initial fetchStatus).
    container.listen(vpnStatusControllerProvider, (_, _) {});

    // wait for the controller to finish its initial build (fetchStatus)
    await container.read(vpnStatusControllerProvider.future);
  });

  test(
    "reconnect falls back to quick connect when no suitable server is found",
    () async {
      await container
          .read(vpnStatusControllerProvider.notifier)
          .reconnectOrQuickConnect(
            pb_status.ConnectionParameters(group: pb.ServerGroup.DOUBLE_VPN),
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
        pb_status.ConnectionState.CONNECTED,
      );
    },
  );

  test("reconnect connects normally when a suitable server exists", () async {
    await container
        .read(vpnStatusControllerProvider.notifier)
        .reconnectOrQuickConnect(
          pb_status.ConnectionParameters(countryCode: "fr"),
        );

    expect(container.read(popupsProvider), isNull);
    expect(
      GrpcServer.instance.vpnStatus.status.state,
      pb_status.ConnectionState.CONNECTED,
    );
  });
}
