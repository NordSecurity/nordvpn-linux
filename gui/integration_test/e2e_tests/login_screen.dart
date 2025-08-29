import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/config.dart';

import '../../test/utils/finders.dart';
import '../../test/utils/safe_test_widget.dart';
import '../../test/utils/test_helpers.dart';

void runLoginTests() async {
  group("test login screen", () {
    // use safeTestWidget because login will timeout and generate an exception
    // of the Future value returned by the gRPC call
    safeTestWidgets("test login timeout", (tester) async {
      final app = await tester.setupIntegrationTests(
        config: ConfigImpl(loginTimeoutDuration: Duration(milliseconds: 500)),
      );

      // accept consent
      await app.acceptConsent();

      // go to login screen, login button is enabled
      final loginScreen = await app.goToLoginScreen();
      app.appAccount.delayDuration = Duration(seconds: 5);
      expect(loginScreen.isLoginButtonEnabled(), isTrue);

      // login button is disabled for the duration of login timeout
      await loginScreen.clickLogin();
      await loginScreen.waitUntilFound(
        loginButtonLoadingIndicator(),
        // finished animation == button is enabled again
        finishAnimations: false,
      );
      expect(loginScreen.isLoginButtonEnabled(), isFalse);

      // after timeout, button is enabled again
      await loginScreen.waitForLoginButton();
      expect(loginScreen.isLoginButtonEnabled(), isTrue);
    });
  });
}
