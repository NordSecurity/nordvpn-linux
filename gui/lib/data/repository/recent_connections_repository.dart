import 'package:fixnum/fixnum.dart';
import 'package:nordvpn/data/models/recent_connections.dart';
import 'package:nordvpn/grpc/grpc_service.dart';
import 'package:nordvpn/pb/daemon/service.pbgrpc.dart';
import 'package:nordvpn/pb/daemon/recent_connections.pb.dart';
import 'package:riverpod/riverpod.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'recent_connections_repository.g.dart';

@riverpod
RecentConnectionsRepository recentConnectionsRepository(Ref ref) {
  return RecentConnectionsRepository();
}

class RecentConnectionsRepository {
  final DaemonClient _client;

  RecentConnectionsRepository([DaemonClient? client])
    : _client = client ?? createDaemonClient();

  Future<List<RecentConnection>> getRecentConnections(int limit) async {
    final request = RecentConnectionsRequest(limit: Int64(limit));
    final response = await _client.getRecentConnections(request);
    return response.connections
        .map((pb) => RecentConnection.fromPb(pb))
        .toList();
  }
}
