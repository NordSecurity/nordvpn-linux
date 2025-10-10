import 'package:nordvpn/data/models/recent_connections.dart';
import 'package:nordvpn/data/providers/app_state_provider.dart';
import 'package:nordvpn/data/repository/recent_connections_repository.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:nordvpn/constants.dart';

part 'recent_connections_controller.g.dart';

@riverpod
class RecentConnectionsController extends _$RecentConnectionsController
    implements RecentConnectionsListObserver {
  @override
  Future<List<RecentConnection>> build() async {
    // Register as an observer to listen for updates
    final appState = ref.read(appStateProvider);
    appState.addRecentConnectionsListObserver(this);
    ref.onDispose(() => appState.removeRecentConnectionsListObserver(this));

    // Fetch the initial list
    return ref
        .watch(recentConnectionsRepositoryProvider)
        .getRecentConnections(maxRecentConnections);
  }

  @override
  void onRecentConnectionsListChanged(
    List<RecentConnection> recentConnections,
  ) {
    state = AsyncData(recentConnections);
  }
}
