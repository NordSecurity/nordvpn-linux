import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:grpc/grpc.dart';
import 'package:nordvpn/config.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/pb/daemon/recent_connections.pb.dart';
import 'package:nordvpn/grpc/grpc_service.dart';
import 'package:nordvpn/grpc/protobuf_utils.dart';
import 'package:nordvpn/grpc/ui_event_interceptor.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/common.pb.dart';
import 'package:nordvpn/pb/daemon/connect.pb.dart';
import 'package:nordvpn/pb/daemon/servers.pb.dart';
import 'package:nordvpn/pb/daemon/service.pbgrpc.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';
import 'package:nordvpn/pb/daemon/uievent.pbenum.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:fixnum/fixnum.dart' as $fixnum;

part 'vpn_repository.g.dart';

/// Provides functionality to communicate over gRPC with the daemon
/// for VPN related functionality.
/// This behaves like a singleton.
class VpnRepository {
  final DaemonClient _client;

  VpnRepository([DaemonClient? client, Config? config])
    : _client = client ?? createDaemonClient();

  /// Connects to a VPN server.
  ///
  /// [itemName] specifies the UI event type for analytics
  Future<int> connect(
    ConnectArguments args, {
    required UIEvent_ItemName itemName,
  }) {
    final options = createUiEventCallOptions(
      formReference: UIEvent_FormReference.HOME_SCREEN,
      itemName: itemName,
      itemValue: args.toUIEventItemValue(),
    );
    return _connect(args.toConnectRequest(), options: options);
  }

  Future<int> _connect(ConnectRequest req, {CallOptions? options}) async {
    final stream = _client.connect(req, options: options);

    await for (var data in stream) {
      final status = data.type.toInt();
      switch (status) {
        case DaemonStatusCode.connecting:
          continue;
        default:
          return status;
      }
    }
    return DaemonStatusCode.failure;
  }

  Future<int> reconnect(ConnectionParameters args) {
    return _connect(args.toConnectRequest());
  }

  /// Disconnects from the current VPN server.
  Future<int> disconnect() async {
    final options = createUiEventCallOptions(
      formReference: UIEvent_FormReference.HOME_SCREEN,
      itemName: UIEvent_ItemName.DISCONNECT,
    );
    final stream = _client.disconnect(Empty(), options: options);

    try {
      await for (var data in stream) {
        final status = data.type.toInt();
        switch (status) {
          case DaemonStatusCode.connecting:
            continue;

          default:
            return status;
        }
      }
    } catch (error) {
      logger.f("disconnect thrown error: $error");
    }

    return DaemonStatusCode.failure;
  }

  Future<int> cancelConnect() async {
    final response = await _client.connectCancel(Empty());
    return response.type.toInt();
  }

  Future<StatusResponse> fetchStatus() async {
    return await _client.status(Empty());
  }

  Future<ServersResponse> fetchServers() async {
    final response = await _client.getServers(Empty());
    return response;
  }

  Future<RecentConnectionsResponse> fetchRecentConnections(int limit) async {
    final request = RecentConnectionsRequest(limit: $fixnum.Int64(limit));
    return await _client.getRecentConnections(request);
  }
}

@Riverpod(keepAlive: true)
VpnRepository vpnRepository(Ref ref) {
  return VpnRepository();
}
