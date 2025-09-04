import 'dart:io';

import 'package:grpc/grpc.dart';
import 'package:grpc/grpc_or_grpcweb.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/mocks/daemon/grpc_server.dart';
import 'package:nordvpn/grpc/error_handling_interceptor.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/service.pbgrpc.dart';
import 'package:nordvpn/service_locator.dart';

const _socketEnvVar = 'NORDVPND_SOCKET';
const String _defaultSocketPath = "/run/nordvpn/nordvpnd.sock";

String _socketPath() {
  final socketPath =
      (Platform.environment[_socketEnvVar]?.trim().isNotEmpty ?? false)
      ? Platform.environment[_socketEnvVar]!.trim()
      : _defaultSocketPath;
  logger.i("daemon socket path: $socketPath");
  return socketPath;
}

/// Creates a new channel to the daemon.
/// Preferably is to use the shared instance from sl() instead of creating a new one
ClientChannel createNewChannel() {
  if (useMockDaemon) {
    // when mocked server is used, then create a connection to the local host
    return GrpcOrGrpcWebClientChannel.grpc(
      "localhost",
      port: defaultPortNumber,
      options: const ChannelOptions(credentials: ChannelCredentials.insecure()),
    );
  }
  final channel = GrpcOrGrpcWebClientChannel.grpc(
    InternetAddress(_socketPath(), type: InternetAddressType.unix),
    port: 0,
    options: const ChannelOptions(credentials: ChannelCredentials.insecure()),
  );

  return channel;
}

// Creates a new DaemonClient and register the error interceptor for it.
// It uses the shared channel from sl() when channel is null.
// Recommended to use this instead of manually creating a new DaemonClient
// because the ErrorHandlingInterceptor will automatically be added to the client.
DaemonClient createDaemonClient([ClientChannel? channel]) {
  final ch = channel ?? sl();
  ErrorHandlingInterceptor errorInterceptor = sl();
  return DaemonClient(ch, interceptors: [errorInterceptor]);
}
