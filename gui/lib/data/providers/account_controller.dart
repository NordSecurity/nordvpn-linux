import 'dart:async';

import 'package:grpc/grpc.dart';
import 'package:nordvpn/config.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/data/models/servers_list.dart';
import 'package:nordvpn/data/models/user_account.dart';
import 'package:nordvpn/data/providers/app_state_provider.dart';
import 'package:nordvpn/data/providers/grpc_connection_controller.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/providers/servers_list_controller.dart';
import 'package:nordvpn/data/repository/account_repository.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/internal/uri_launch_extension.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/account.pb.dart';
import 'package:nordvpn/pb/daemon/login.pb.dart';
import 'package:nordvpn/pb/daemon/state.pb.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'account_controller.g.dart';

const _ignoredStatusCodes = {
  DaemonStatusCode.success,
  DaemonStatusCode.alreadyLoggedIn,
  DaemonStatusCode.notLoggedIn,
};

// Handles the user account operations
@riverpod
final class AccountController extends _$AccountController
    implements AccountObserver {
  ServersList? _serversList;
  // Timer used to display the loading indicator for the login button
  Timer? _loginTimer;

  @override
  FutureOr<UserAccount?> build() async {
    _loginTimer?.cancel();
    final isConnected = ref.watch(grpcConnectionControllerProvider);
    if (isConnected is! AsyncData) {
      return null;
    }

    _registerNotifications();

    return await _fetchAccount();
  }

  void _dispose(AppStateChange notification) {
    notification.removeAccountObserver(this);
    _loginTimer?.cancel();
  }

  Future<void> register() async {
    final vpnProvider = ref.read(accountRepositoryProvider);
    int status = DaemonStatusCode.success;
    try {
      final registerData = await vpnProvider.register();

      switch (registerData.status) {
        case LoginStatus.SUCCESS:
          final uri = Uri.parse(registerData.url);

          if (!await uri.launch()) {
            if (!useMockDaemon) {
              logger.e('Could not launch $uri');
              status = DaemonStatusCode.failedToOpenBrowserToCreateAccount;
            }
          }
          break;
        case LoginStatus.ALREADY_LOGGED_IN:
          assert(false, "no account creation allowed when logged in");
          break;
        case LoginStatus.NO_NET:
          status = DaemonStatusCode.offline;
          break;
        case LoginStatus.UNKNOWN_OAUTH2_ERROR:
          status = DaemonStatusCode.failure;
          break;
        case LoginStatus.CONSENT_MISSING:
          // this should not happen because user needs to go through
          // consent flow before login screen will be available
          logger.e(
            "user consent is still missing when registering, this should not happen",
          );
          break;
      }
    } on GrpcError catch (e) {
      status = DaemonStatusCode.fromGrpcError(e);
    } catch (e) {
      status = DaemonStatusCode.failure;
      logger.e("register thrown unknown error: $e");
    }
    _showError(status);
  }

  Future<void> login() async {
    state = const AsyncLoading();
    final vpnProvider = ref.read(accountRepositoryProvider);
    int status = DaemonStatusCode.success;
    try {
      Config config = sl();
      // This gRPC signals errors using LoginStatus codes
      // The only exception it should throw is gRPC timeout
      final loginData = await vpnProvider.login(config.loginTimeout);

      switch (loginData.status) {
        case LoginStatus.SUCCESS:
          final uri = Uri.parse(loginData.url);
          if (await uri.launch()) {
            _startLoginTimer(config.loginTimeout);
            return;
          } else {
            if (!useMockDaemon) {
              status = DaemonStatusCode.failedToOpenBrowserToLogin;
            }
          }
          break;
        case LoginStatus.ALREADY_LOGGED_IN:
          status = DaemonStatusCode.alreadyLoggedIn;
          await _refreshAccount();
          break;
        case LoginStatus.NO_NET:
          status = DaemonStatusCode.offline;
          break;
        case LoginStatus.UNKNOWN_OAUTH2_ERROR:
          status = DaemonStatusCode.failure;
          break;
        case LoginStatus.CONSENT_MISSING:
          // this should not happen because user needs to go through
          // consent flow before login screen will be available
          logger.e(
            "user consent is still missing when logging in, this should not happen",
          );
          break;
      }
    } on GrpcError catch (e) {
      status = DaemonStatusCode.fromGrpcError(e);
    } catch (e) {
      status = DaemonStatusCode.failure;
      logger.e("login thrown unknown error: $e");
    }
    cancelLoading();
    _showError(status);
  }

  Future<void> logout() async {
    final vpnProvider = ref.read(accountRepositoryProvider);
    int status = DaemonStatusCode.success;
    try {
      // This gRPC signals errors mostly using status codes
      // The only exception it should throw is NotLoggedIn
      status = await vpnProvider.logout();
    } on GrpcError catch (e) {
      status = DaemonStatusCode.fromGrpcError(e);
    } catch (e) {
      status = DaemonStatusCode.failure;
      logger.e("login thrown unknown error: $e");
    }

    _showError(status);
  }

  void cancelLoading() {
    _loginTimer?.cancel();
    if (state is AsyncLoading && state.value == null) {
      state = const AsyncData(null);
    }
  }

  void _showError(int code) {
    if (!_ignoredStatusCodes.contains(code)) {
      ref.read(popupsProvider.notifier).show(code);
    }
  }

  void _registerNotifications() {
    final notification = ref.read(appStateProvider);
    notification.addAccountObserver(this);
    ref.listen(serversListControllerProvider, (_, next) async {
      if (next is AsyncData && next.value != null) {
        _serversList = next.value!;
        // servers list changed then refresh user info
        _refreshAccount();
      }
    });

    ref.onDispose(() => _dispose(notification));
  }

  @override
  void onAccountChanged(LoginEventType type) async {
    if (type == LoginEventType.LOGOUT) {
      _updateState(null);
    } else {
      _refreshAccount();
    }
  }

  Future<void> _refreshAccount() async {
    try {
      final account = await _fetchAccount();
      _updateState(account);
    } catch (error, stackTrace) {
      state = AsyncError(error, stackTrace);
    }
  }

  void _startLoginTimer(Duration loginTimeout) {
    _loginTimer?.cancel();
    _loginTimer = Timer(loginTimeout, () => cancelLoading());
  }

  Future<UserAccount?> _fetchAccount() async {
    int status = DaemonStatusCode.success;

    try {
      // This gRPC signals errors mostly using status codes
      // But the expected exceptions are NotLoggedIn and ErrUnhandled
      final accountInfo = await ref
          .read(accountRepositoryProvider)
          .accountInfo();

      status = accountInfo.type.toInt();
      switch (status) {
        case DaemonStatusCode.success:
        case DaemonStatusCode.noService:
          // Success flow will return here
          return _buildUserAccount(accountInfo);
        default:
          // We want to popup the error
          break;
      }
    } on GrpcError catch (e) {
      status = DaemonStatusCode.fromGrpcError(e);
    } catch (e) {
      status = DaemonStatusCode.failure;
      logger.e("failed to fetch the account info $e");
    }

    if (_ignoredStatusCodes.contains(status)) {
      return null;
    }

    throw status;
  }

  UserAccount? _buildUserAccount(AccountResponse? response) {
    if (response == null) {
      return null;
    }

    if (response.dedicatedIpServices.isEmpty) {
      return UserAccount.fromResponse(response);
    }

    final dipIds = response.dedicatedIpServices
        .expand((services) => services.serverIds)
        .map((id) => id.toInt())
        .toSet();

    final servers = _serversList?.findServers(dipIds, ServerType.dedicatedIP);
    if (servers == null) {
      logger.e("failed to match DIP servers ids");
    }

    // if the application doesn't have yet servers list or there is an error
    // fetching it, return the user account without any dedicated IP servers
    return UserAccount.from(response, servers);
  }

  @override
  void onAccountModified(AccountModification _) {
    // NOTE: This method is supposed to update the account with the
    // [AccountModification] passed as an argument. I'm refreshing the whole
    // account here instead because of LVPN-7955 - the account cache is not
    // invalidated properly, so the first [AccountController._refreshAccount]
    // call made in [AccountController.onAccountChanged] is insufficient.
    // When this method is triggered by account modification event, the
    // cache is already refreshed and we get correct account data.
    // The downside of this hack is that this method is called with some delay
    // relative to the act of logging in, so for few seconds, the account
    // information displayed in GUI is incorrect.
    _refreshAccount();
  }

  void _updateState(UserAccount? userAccount) {
    _loginTimer?.cancel();
    if (state is AsyncData && state.value == userAccount) {
      return;
    }

    state = AsyncData(userAccount);
  }
}
