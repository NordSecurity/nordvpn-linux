import 'dart:async';

import 'package:fixnum/fixnum.dart';
import 'package:nordvpn/data/mocks/daemon/cancelable_delayed.dart';
import 'package:nordvpn/data/mocks/daemon/connect_arguments_extension.dart';
import 'package:nordvpn/data/mocks/daemon/mock_servers_list.dart';
import 'package:nordvpn/data/mocks/daemon/mock_application_settings.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/pb/daemon/common.pb.dart';
import 'package:nordvpn/pb/daemon/connect.pb.dart';
import 'package:nordvpn/pb/daemon/state.pb.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';
import 'package:nordvpn/pb/daemon/config/group.pbenum.dart' as config;

// Store information about the VPN status for the mocked daemon
final class MockVpnStatus extends CancelableDelayed {
  final StreamController<AppState> stream;
  final MockApplicationSettings appSettings;
  final MockServersList serversList;

  MockVpnStatus(this.stream, this.appSettings, this.serversList);

  var delayDuration = Duration(milliseconds: 500);

  StatusResponse _status = StatusResponse(state: ConnectionState.DISCONNECTED);
  String? error;
  int? connectingErrorStatusCode;
  int? disconnectErrorStatusCode;
  bool _shouldCancel = false;

  StatusResponse get status => _status;
  void setStatus(StatusResponse newStatus) {
    _status = newStatus;
    sendNotification(newStatus);
  }

  void sendNotification(StatusResponse newStatus) {
    stream.add(AppState(connectionStatus: newStatus));
  }

  Stream<Payload> findServerAndConnect(ConnectRequest args) async* {
    _shouldCancel = false;
    await delayed(delayDuration);
    if (error != null) {
      throw error!;
    }

    if (connectingErrorStatusCode != null) {
      yield Payload(type: Int64(connectingErrorStatusCode!));

      return;
    }

    final server = serversList.findServer(args);
    if (server == null) {
      throw "server not found";
    }

    String? country;
    String? city;

    if (args.hasServerTag()) {
      country = server.countryCode;
      if (server.countryCode.toLowerCase() != args.serverTag) {
        city = server.cityName;
      }
    }

    final settings = appSettings.settings.data;
    final group = args.serverGroup.isNotEmpty
        ? args.toServerGroup()
        : appSettings.settings.data.obfuscate
        ? config.ServerGroup.OBFUSCATED
        : null;

    StatusResponse newStatus = StatusResponse(
      state: ConnectionState.CONNECTING,
      city: server.cityName,
      country: server.countryCode,
      hostname: server.server.hostName,
      virtualLocation: server.server.virtual,
      name: "NOT SET",
      ip: "NOT IP",
      download: Int64(0),
      postQuantum: settings.postquantumVpn,
      protocol: settings.protocol,
      upload: Int64(0),
      uptime: Int64(0),
      technology: settings.technology,
      parameters: ConnectionParameters(
        country: country,
        city: city,
        group: group,
      ),
    );

    yield Payload(type: Int64(DaemonStatusCode.connecting));
    sendNotification(newStatus);

    await delayed(delayDuration + delayDuration);

    if (_shouldCancel) {
      _shouldCancel = false;
      yield Payload(type: Int64(DaemonStatusCode.disconnected));
      newStatus.state = ConnectionState.DISCONNECTED;
      sendNotification(newStatus);

      return;
    }

    newStatus.state = ConnectionState.CONNECTED;
    yield Payload(type: Int64(DaemonStatusCode.connected));
    setStatus(newStatus);

    if (_shouldCancel) {
      _shouldCancel = false;
      yield Payload(type: Int64(DaemonStatusCode.disconnected));
      newStatus.state = ConnectionState.DISCONNECTED;
      sendNotification(newStatus);

      return;
    }
  }

  Stream<Payload> disconnect() async* {
    if (error != null) {
      throw error!;
    }

    await delayed(delayDuration);
    if (disconnectErrorStatusCode != null) {
      yield Payload(type: Int64(disconnectErrorStatusCode!));

      return;
    }

    await delayed(delayDuration);

    yield Payload(type: Int64(DaemonStatusCode.disconnected));
    setStatus(StatusResponse(state: ConnectionState.DISCONNECTED));

    await delayed(delayDuration);
  }

  Future<Payload> cancel() async {
    _shouldCancel = true;
    return Payload(type: Int64(DaemonStatusCode.success));
  }
}
