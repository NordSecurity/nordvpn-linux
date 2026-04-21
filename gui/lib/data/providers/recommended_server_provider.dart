import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/providers/app_state_provider.dart';
import 'package:nordvpn/data/repository/vpn_repository.dart';
import 'package:nordvpn/pb/daemon/servers.pb.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'recommended_server_provider.g.dart';

@Riverpod(keepAlive: true)
class RecommendedServer extends _$RecommendedServer
    implements VpnSettingsObserver {
  ApplicationSettings? _appSettings;

  @override
  FutureOr<RecommendedServerLocation> build() async {
    _registerNotifications();
    return await ref
        .read(vpnRepositoryProvider)
        .fetchRecommendedServerLocation();
  }

  void _registerNotifications() {
    final notification = ref.read(appStateProvider);
    notification.addSettingsObserver(this);
    ref.onDispose(() {
      notification.removeSettingsObserver(this);
    });
  }

  @override
  Future<void> onSettingsChanged(ApplicationSettings settings) async {
    if (!_shouldRefetch(settings)) {
      return;
    }

    state = const AsyncValue.loading();

    state = await AsyncValue.guard(() async {
      return await ref
          .read(vpnRepositoryProvider)
          .fetchRecommendedServerLocation();
    });
  }

  bool _shouldRefetch(ApplicationSettings newSettings) {
    final currentSettings = _appSettings;
    _appSettings = newSettings;

    if (currentSettings == null) {
      return false;
    }

    return currentSettings.obfuscatedServers != newSettings.obfuscatedServers ||
        currentSettings.protocol != newSettings.protocol;
  }
}
