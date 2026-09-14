import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/providers/app_state_provider.dart';
import 'package:nordvpn/data/repository/vpn_repository.dart';
import 'package:nordvpn/data/repository/vpn_settings_repository.dart';
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
    // Load the current settings once, so that when the first settings change arrives
    // _shouldRefetch has something to compare it against. Without this the first change
    // would only be stored and the location would not be refetched.

    _appSettings ??= await ref.read(vpnSettingsProvider).fetchSettings();
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

    return currentSettings.protocol != newSettings.protocol;
  }
}
