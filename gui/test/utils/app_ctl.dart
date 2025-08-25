import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nordvpn/data/mocks/daemon/grpc_server.dart';
import 'package:nordvpn/data/mocks/daemon/mock_account_info.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart' as pbstatus;
import 'package:nordvpn/data/mocks/daemon/mock_application_settings.dart';
import 'package:nordvpn/data/mocks/daemon/mock_vpn_status.dart';
import 'package:nordvpn/data/models/server_info.dart';
import 'package:nordvpn/pb/daemon/account.pb.dart';
import 'package:nordvpn/pb/daemon/config/technology.pbenum.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';
import 'package:nordvpn/router/router.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/i18n/strings.g.dart';

import 'account_screen_handle.dart';
import 'auto_connect_settings_screen_handle.dart';
import 'connect_settings_screen_handle.dart';
import 'custom_dns_settings_handle.dart';
import 'fakes.dart';
import 'finders.dart';
import 'login_screen_handle.dart';
import 'test_helpers.dart';
import 'vpn_screen_handle.dart';

// Convenience class to easier navigate and run the integration tests
final class AppCtl {
  final WidgetTester tester;
  MockAccountInfo get appAccount => GrpcServer.instance.account;
  MockApplicationSettings get appSettings => GrpcServer.instance.appSettings;
  MockVpnStatus get vpnStatus => GrpcServer.instance.vpnStatus;

  AppCtl({required this.tester});

  Future<void> refreshAppState() async {
    await waitForUiUpdates(duration: Duration(milliseconds: 800));
  }

  Future<void> expireSubscription() async {
    appAccount.setAccount(expiresAt: DateTime(2010));
    await tester.pumpAndSettle();
  }

  Future<void> renewSubscription() async {
    appAccount.setAccount(
      expiresAt: DateTime.now().add(const Duration(days: 1)),
    );
    await tester.pumpAndSettle();
  }

  Future<void> logIn({AccountResponse? account}) async {
    await acceptConsent();
    appAccount.replaceAccount(account ?? fakeAccount());
    await tester.pumpAndSettle();
  }

  Future<void> acceptConsent({AccountResponse? account}) async {
    // TODO: use riverpod provider
    await tester.pumpUntilFound(find.text(t.ui.rejectNonEssential));
    await tester.tap(find.text(t.ui.rejectNonEssential));
    await tester.pumpAndSettle();
  }

  AppCtl goTo(AppRoute route) {
    final context = goRouterKey.currentContext;
    context!.go(route.toString());
    return this;
  }

  Future<void> connect({
    String? country,
    String? city,
    String? hostname,
    bool? isVirtualLocation,
    ServerType? group,
  }) async {
    final status = StatusResponse(
      country: country,
      city: city,
      state: pbstatus.ConnectionState.CONNECTED,
      hostname: hostname,
      virtualLocation: isVirtualLocation,
      parameters: ConnectionParameters(
        country: country,
        city: city,
        group: group?.toServerGroup(),
      ),
    );
    vpnStatus.setStatus(status);
  }

  Future<void> setThreatProtection(bool enabled) async {
    await appSettings.setSettings(threatProtectionLite: enabled);
    await refreshAppState();
  }

  Future<void> setObfuscatedServers(bool enabled) async {
    await appSettings.setSettings(technology: Technology.OPENVPN);
    await appSettings.setSettings(obfuscate: enabled);

    await refreshAppState();
  }

  Future<void> waitForUiUpdates({
    Duration duration = const Duration(milliseconds: 100),
    Duration timeout = const Duration(seconds: 5),
  }) {
    return tester.pumpAndSettleWithTimeout(
      duration: duration,
      timeout: timeout,
    );
  }

  Future<LoginScreenHandle> goToLoginScreen() async {
    final loginScreenHandle = LoginScreenHandle(goTo(AppRoute.login));
    await loginScreenHandle.waitUntilFound(loginForm());
    return loginScreenHandle;
  }

  Future<VpnScreenHandle> goToVpnScreen({AccountResponse? account}) async {
    await logIn(account: account);
    await goTo(AppRoute.vpn).waitForUiUpdates();
    final vpnScreenHandle = VpnScreenHandle(this);
    await vpnScreenHandle.waitUntilFound(vpnStatusCard());
    return vpnScreenHandle;
  }

  Future<AccountScreenHandle> goToAccountScreen({
    AccountResponse? account,
  }) async {
    await logIn(account: account);
    await goTo(AppRoute.settingsAccount).waitForUiUpdates();
    final accountScreenHandle = AccountScreenHandle(this);
    await accountScreenHandle.waitUntilFound(userInfo());
    return accountScreenHandle;
  }

  Future<void> changeVirtualServers(bool enabled) async {
    await appSettings.setSettings(virtualLocation: enabled);
    await refreshAppState();
  }

  Future<ConnectionSettingsScreenHandle> goToConnectionSettingsScreen({
    AccountResponse? account,
  }) async {
    await logIn(account: account);
    await goTo(AppRoute.settingsVpnConnection).waitForUiUpdates();
    final handle = ConnectionSettingsScreenHandle(this);
    await handle.waitUntilFound(vpnConnectionBreadcrumb());
    return handle;
  }

  Future<AutoConnectSettingsScreenHandle> goToAutoConnectSettingsScreen({
    AccountResponse? account,
  }) async {
    await logIn(account: account);
    await goTo(AppRoute.settingsAutoconnect).waitForUiUpdates();
    final handle = AutoConnectSettingsScreenHandle(this);
    await handle.waitUntilFound(autoConnectPanel());
    return handle;
  }

  Future<CustomDnsSettingsHandle> goToCustomDnsSettingsScreen({
    AccountResponse? account,
  }) async {
    await logIn(account: account);
    await goTo(AppRoute.settingsCustomDns).waitForUiUpdates();
    final handle = CustomDnsSettingsHandle(this);
    await handle.waitUntilVisible();
    return handle;
  }
}
