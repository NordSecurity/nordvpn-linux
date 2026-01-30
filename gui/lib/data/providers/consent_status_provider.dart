import 'package:nordvpn/data/models/app_settings.dart';
import 'package:nordvpn/data/providers/app_state_provider.dart';
import 'package:nordvpn/data/providers/grpc_connection_controller.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/repository/vpn_settings_repository.dart';
import 'package:nordvpn/logger.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'consent_status_provider.g.dart';

enum ConsentLevel { none, acceptedAll, essentialOnly }

// ConsentStatus is a provider that controls the ConsentScreen
@riverpod
final class ConsentStatus extends _$ConsentStatus
    implements VpnSettingsObserver {
  @override
  FutureOr<ConsentLevel> build() async {
    // If grpcConnectionControllerProvider will have AsyncError
    // this call will throw an error which will set this provider's
    // status to AsyncError using the error thrown by the grpc
    await ref.watch(grpcConnectionControllerProvider.future);

    _registerNotifications();
    return await _fetchConsentStatus();
  }

  void _registerNotifications() {
    final appNotification = ref.read(appStateProvider);
    appNotification.addSettingsObserver(this);
    ref.onDispose(() => _dispose(appNotification));
  }

  void _dispose(AppStateChange appNotification) {
    appNotification.removeSettingsObserver(this);
  }

  Future<ConsentLevel> _fetchConsentStatus() async {
    final settings = await ref.read(vpnSettingsProvider).fetchSettings();
    return settings.analyticsConsent;
  }

  @override
  Future<void> onSettingsChanged(ApplicationSettings settings) async {
    if (state is AsyncData && state.value == settings.analyticsConsent) {
      return;
    }
    state = AsyncData(settings.analyticsConsent);
  }

  Future<void> setLevel(ConsentLevel level) async {
    try {
      await ref
          .read(vpnSettingsControllerProvider.notifier)
          .setAnalytics(level == ConsentLevel.acceptedAll);
      state = AsyncData(level);
    } catch (error, stackTrace) {
      logger.e("failed to set consent level: $error");
      state = AsyncError(error, stackTrace);
    }
  }

  Future<void> retry() async {
    try {
      final consent = await _fetchConsentStatus();
      state = AsyncData(consent);
    } catch (error, stack) {
      logger.e("failed to retry consent status: $error");
      state = AsyncError(error, stack);
    }
  }
}
