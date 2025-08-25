import 'package:flutter/widgets.dart';
import 'package:grpc/grpc.dart';
import 'package:nordvpn/data/mocks/daemon/mock_account_info.dart';
import 'package:nordvpn/data/mocks/daemon/mock_application_settings.dart';
import 'package:nordvpn/data/mocks/daemon/mock_daemon.dart';
import 'package:nordvpn/data/mocks/daemon/mock_servers_list.dart';
import 'package:nordvpn/data/mocks/daemon/mock_vpn_status.dart';
import 'package:nordvpn/logger.dart';

const defaultPortNumber = 50051;

// Controls the lifetime of the mock gRPC server
final class GrpcServer {
  static final GrpcServer instance = GrpcServer._internal();
  factory GrpcServer() {
    return instance;
  }
  GrpcServer._internal();

  final List<VoidCallback> _shutdownCallbacks = [];
  Server? _server;
  MockDaemon? _daemon;

  bool get isRunning => _daemon != null;

  MockDaemon get daemon => _daemon!;
  MockAccountInfo get account => daemon.account;
  MockApplicationSettings get appSettings => daemon.appSettings;
  MockServersList get serversList => daemon.serversList;
  MockVpnStatus get vpnStatus => daemon.vpnStatus;

  /// Starts the gRPC server
  Future<void> start() async {
    if (isRunning) {
      return;
    }
    _daemon = MockDaemon();
    _server = Server.create(services: [_daemon!]);
    await _server!.serve(port: defaultPortNumber, shared: true);
    logger.i('✅ gRPC Server started on port $defaultPortNumber');
  }

  /// Stops the gRPC server
  Future<void> stop({Duration timeout = const Duration(seconds: 5)}) async {
    if (!isRunning) {
      logger.i('Stopping the gRPC server but it is not running');
      return;
    }
    logger.i('🛑 Stopping gRPC Server...');
    _daemon!.dispose();
    _daemon = null;

    await _server!.shutdown();
    _server = null;
    for (final callback in _shutdownCallbacks) {
      callback();
    }
    _shutdownCallbacks.clear();
    logger.i('✅ gRPC Server stopped.');
  }

  // Register a callback when the server will be stopped
  void registerOnShutdown(VoidCallback callback) {
    _shutdownCallbacks.add(callback);
  }
}
