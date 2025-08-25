import 'package:nordvpn/data/providers/app_state_provider.dart';
import 'package:nordvpn/data/providers/grpc_connection_controller.dart';
import 'package:nordvpn/data/repository/account_repository.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/state.pb.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'login_status_provider.g.dart';

// Monitors the the logged in status
// There are 3 possible states:
//  * loading - data is being fetched
//  * error   - something went wrong
// *  data    - store the login status reported by the daemon (true/false)
@riverpod
final class LoginStatus extends _$LoginStatus implements AccountObserver {
  @override
  FutureOr<bool> build() async {
    final isConnected = ref.watch(grpcConnectionControllerProvider);
    if (isConnected is! AsyncData) {
      throw "grpc connect not established";
    }
    _registerNotifications();

    return await _fetchLoginState();
  }

  @override
  Future<void> onAccountChanged(LoginEventType type) async {
    if (type == LoginEventType.LOGOUT) {
      state = AsyncData(false);
    } else {
      _refreshState();
    }
  }

  @override
  Future<void> onAccountModified(AccountModification _) async {
    _refreshState();
  }

  Future<bool> _fetchLoginState() async {
    // Returns only true or false
    // No error status code or exception thrown
    // that should be handled here
    final status = await ref.read(accountRepositoryProvider).isLoggedIn();
    return status;
  }

  Future<void> _refreshState() async {
    try {
      state = AsyncData(await _fetchLoginState());
    } catch (error, stack) {
      logger.e("failed to get login status $error");
      state = AsyncError(error, stack);
    }
  }

  void _registerNotifications() {
    final notification = ref.read(appStateProvider);
    notification.addAccountObserver(this);
    ref.onDispose(() => _dispose(notification));
  }

  void _dispose(AppStateChange notification) {
    notification.removeAccountObserver(this);
  }

  Future<void> retry() async {
    await _refreshState();
  }
}
