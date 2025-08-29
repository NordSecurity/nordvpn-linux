import 'dart:async';

import 'package:fixnum/fixnum.dart';
import 'package:nordvpn/constants.dart';
import 'package:nordvpn/data/mocks/daemon/cancelable_delayed.dart';
import 'package:nordvpn/data/mocks/daemon/mock_servers_list.dart';

import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/internal/dates_util.dart';
import 'package:nordvpn/pb/daemon/account.pb.dart';
import 'package:nordvpn/pb/daemon/common.pb.dart';
import 'package:nordvpn/pb/daemon/login.pb.dart';
import 'package:nordvpn/pb/daemon/state.pb.dart';

// Store information about the user account for the mocked
final class MockAccountInfo extends CancelableDelayed {
  final StreamController<AppState> stream;
  final MockServersList serversList;

  MockAccountInfo(this.stream, this.serversList);

  bool get _isLoggedIn => _account?.hasEmail() ?? false;

  AccountResponse? _account;
  AccountResponse get account {
    if (error != null) {
      throw error!;
    }

    if (!_isLoggedIn) {
      throw error ?? "you are not logged in";
    }
    return _account!;
  }

  LoginStatus? statusCode;
  String? error;

  // When null, has no DIP subscription.
  // For 0 it has subscription, but no servers selected
  // For bigger values, selectes the specified number of servers from DIP
  // servers list, normally 1 or 2
  int? _numberOfDipServers;
  int? get numberOfDipServers => _numberOfDipServers;
  set numberOfDipServers(int? numberOfDipServers) {
    _numberOfDipServers = numberOfDipServers;
    if (_account != null) {
      final dip = (_numberOfDipServers != null)
          ? serversList.dipServers
                .map((e) => e.id)
                .toList()
                .sublist(0, _numberOfDipServers!)
          : null;

      setAccount(dedicatedIpServers: dip);
    }
  }

  // delay added for each call
  Duration delayDuration = Duration(seconds: 1);
  // Login URL
  String loginUrl = "<invalid url not to open browser>";

  void replaceAccount(AccountResponse? newAccount) {
    _account = newAccount;
    if (newAccount == null) {
      stream.add(AppState(loginEvent: LoginEvent(type: LoginEventType.LOGOUT)));
    } else {
      stream.add(AppState(loginEvent: LoginEvent(type: LoginEventType.LOGIN)));
    }
  }

  void setAccount({
    Int64? type,
    String? email,
    DateTime? expiresAt,
    String? username,
    List<Int64>? dedicatedIpServers,
  }) {
    final wasLoggedIn = _isLoggedIn;
    final oldAccount = _account ?? AccountResponse();

    final date = (expiresAt != null)
        ? daemonDateFormat.format(expiresAt)
        : null;

    List<DedidcatedIPService>? dip;
    if (dedicatedIpServers != null) {
      if (dedicatedIpServers.isEmpty) {
        dip = [];
      } else {
        final service = DedidcatedIPService(
          serverIds: dedicatedIpServers,
          dedicatedIpExpiresAt: daemonDateFormat.format(DateTime(2030)),
        );

        dip = [service];
      }
    }

    // update the account info
    _account = AccountResponse(
      type: type ?? oldAccount.type,
      email: email ?? oldAccount.email,
      expiresAt: date ?? oldAccount.expiresAt,
      username: username ?? oldAccount.username,
      dedicatedIpServices: dip,
      dedicatedIpStatus: (dip != null) ? Int64(DaemonStatusCode.success) : null,
    );

    // old account was logged in and new account is logged in so it's just modification
    if (wasLoggedIn && _isLoggedIn) {
      stream.add(
        AppState(
          accountModification: AccountModification(
            expiresAt: _account!.expiresAt,
          ),
        ),
      );
      return;
    }

    final eventType = !_isLoggedIn
        ? LoginEventType.LOGOUT
        : LoginEventType.LOGIN;

    stream.add(AppState(loginEvent: LoginEvent(type: eventType)));
  }

  Future<LoginOAuth2Response> _updateAccount({
    Int64? type,
    String? email,
    DateTime? expiresAt,
    String? username,
    List<Int64>? dedicatedIpServers,
  }) async {
    if (error != null) {
      throw error!;
    }

    if (statusCode != null) {
      return LoginOAuth2Response(status: statusCode);
    }

    await delayed(delayDuration);

    setAccount(
      type: type,
      email: email,
      expiresAt: expiresAt,
      username: username,
      dedicatedIpServers: dedicatedIpServers,
    );

    return LoginOAuth2Response(status: LoginStatus.SUCCESS);
  }

  Future<LoginOAuth2Response> login() async {
    final dip = (_numberOfDipServers != null)
        ? serversList.dipServers
              .map((e) => e.id)
              .toList()
              .sublist(0, _numberOfDipServers!)
        : null;

    return await _updateAccount(
      type: Int64(DaemonStatusCode.success),
      email: "fake@fake.com",
      expiresAt: DateTime(2030),
      username: "fake",
      dedicatedIpServers: dip,
    );
  }

  Future<Payload> logout() async {
    if (error != null) {
      throw error!;
    }

    await delayed(delayDuration);
    replaceAccount(null);
    return Payload(type: Int64(DaemonStatusCode.success));
  }

  Future<bool> isExpired() async {
    if (!_isLoggedIn) {
      return true;
    }

    final date = parseDate(_account!.expiresAt);
    if ((date == null) || date.isBefore(DateTime.now())) {
      return true;
    }

    return false;
  }

  Future<bool> isLoggedIn() async {
    await delayed(delayDuration);

    if (error != null) {
      throw error!;
    }

    return _isLoggedIn;
  }
}
