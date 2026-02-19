import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/pb/daemon/account.pb.dart';
import 'package:nordvpn/pb/daemon/common.pb.dart';
import 'package:nordvpn/pb/daemon/login.pb.dart';
import 'package:nordvpn/pb/daemon/logout.pb.dart';
import 'package:nordvpn/grpc/grpc_service.dart';
import 'package:nordvpn/grpc/ui_event_interceptor.dart';
import 'package:nordvpn/pb/daemon/service.pbgrpc.dart';
import 'package:nordvpn/pb/daemon/token.pb.dart';
import 'package:nordvpn/pb/daemon/uievent.pbenum.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'account_repository.g.dart';

final class AccountRepository {
  final DaemonClient _client;
  AccountRepository([DaemonClient? client])
    : _client = client ?? createDaemonClient();

  Future<LoginOAuth2Response> register() =>
      _doLogin(LoginType.LoginType_SIGNUP);

  Future<LoginOAuth2Response> _doLogin(LoginType type) async {
    final options = createUiEventCallOptions(
      formReference: UIEvent_FormReference.GUI,
      itemName: UIEvent_ItemName.LOGIN,
    );
    return await _client.loginOAuth2(
      LoginOAuth2Request(type: type),
      options: options,
    );
  }

  Future<bool> isLoggedIn() async {
    final result = await _client.isLoggedIn(Empty());
    return result.isLoggedIn;
  }

  Future<LoginOAuth2Response> login(Duration timeout) =>
      _doLogin(LoginType.LoginType_LOGIN);

  /// Logs out the current user.
  Future<int> logout() async {
    final options = createUiEventCallOptions(
      formReference: UIEvent_FormReference.GUI,
      itemName: UIEvent_ItemName.LOGOUT,
    );
    final result = await _client.logout(
      LogoutRequest(persistToken: false),
      options: options,
    );
    return result.type.toInt();
  }

  Future<AccountResponse> accountInfo() async {
    return await _client.accountInfo(AccountRequest(full: true));
  }

  Future<TokenInfoResponse> tokenInfo() async {
    return await _client.tokenInfo(Empty());
  }
}

@Riverpod(keepAlive: true)
AccountRepository accountRepository(Ref ref) {
  return AccountRepository();
}
