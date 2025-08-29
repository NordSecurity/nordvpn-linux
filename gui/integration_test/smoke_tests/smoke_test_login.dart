import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/pb/daemon/settings.pb.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/config.dart';

import '../../test/utils/account_screen_handle.dart';
import '../../test/utils/connect_settings_screen_handle.dart';
import '../../test/utils/finders.dart';
import '../../test/utils/safe_test_widget.dart';
import '../../test/utils/test_helpers.dart';

void main() {
  WidgetController.hitTestWarningShouldBeFatal = true;

  setUp(() async => await initServiceLocator());
  tearDown(() async => await sl.reset(dispose: true));

  // Call your existing test function
  runLoginSmokeTests();
}

void runLoginSmokeTests() {
  group("Login Smoke Tests", () {
    // use safeTestWidget because login will timeout and generate an exception
    // of the Future value returned by the gRPC call
    // Manual TCID: LVPN-6163
    safeTestWidgets(" - normal login ", (tester) async {
      final app = await tester.setupIntegrationTests(
        config: ConfigImpl(loginTimeoutDuration: Duration(milliseconds: 500)),
      );

      // accept consent
      await app.acceptConsent();
      //
      // go to login screen, login button is enabled
      final loginScreen = await app.goToLoginScreen();
      app.appAccount.delayDuration = Duration(seconds: 1);
      expect(loginScreen.isLoginButtonEnabled(), isTrue);

      // login button is disabled for the duration of login timeout
      await loginScreen.clickLogin();
      await loginScreen.waitUntilFound(
        loginButtonLoadingIndicator(),
        // finished animation == button is enabled again
        finishAnimations: false,
      );
      expect(loginScreen.isLoginButtonEnabled(), isFalse);

      await tester.pumpUntilFound(
        find.text(t.ui.quickConnect),
        timeout: Duration(seconds: 10),
      );
      await app.goTo(AppRoute.settingsAccount).waitForUiUpdates();
      final accountScreenHandle = AccountScreenHandle(app);
      await accountScreenHandle.waitUntilFound(userInfo());
      expect(accountScreenHandle.hasUserInfo(), isTrue);
    });
  });

  group("Login Smoke Test with Kill Switch", () {
    // Manual TCID: LVPN-6478
    testWidgets(" - killswitch login", (tester) async {
      final app = await tester.setupIntegrationTests(
        appSettings: Settings(killSwitch: true),
      );

      // accept consent
      await app.acceptConsent();
      //
      // go to login screen, login button is enabled
      final loginScreen = await app.goToLoginScreen();
      app.appAccount.delayDuration = Duration(seconds: 1);

      await loginScreen.clickTurnOffKillSwitch();

      await loginScreen.waitForLoginButton();
      expect(loginScreen.isLoginButtonEnabled(), isTrue);
      await loginScreen.clickLogin();

      await tester.pumpUntilFound(
        find.text(t.ui.quickConnect),
        timeout: Duration(seconds: 10),
      );
      await app.goTo(AppRoute.settingsVpnConnection).waitForUiUpdates();
      final vpnConnectionHandle = ConnectionSettingsScreenHandle(app);
      await vpnConnectionHandle.waitUntilFound(killSwitchToggle());
      expect(vpnConnectionHandle.isKillSwitchOn(), isFalse);
    });
  });
}
