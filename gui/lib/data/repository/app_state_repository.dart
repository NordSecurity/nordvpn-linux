import 'dart:async';

import 'package:nordvpn/grpc/grpc_service.dart';
import 'package:nordvpn/pb/daemon/common.pb.dart';
import 'package:nordvpn/pb/daemon/service.pbgrpc.dart';
import 'package:nordvpn/pb/daemon/state.pb.dart';
import 'package:riverpod/riverpod.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'app_state_repository.g.dart';

class AppStateRepository {
  final DaemonClient _client;
  AppStateRepository(DaemonClient client) : _client = client;

  Stream<AppState> get stream {
    return _client.subscribeToStateChanges(Empty());
  }
}

// Registers to the daemon app state changes and forwards them
@Riverpod(keepAlive: true)
AppStateRepository appStateRepository(Ref ref) {
  return AppStateRepository(createDaemonClient());
}
