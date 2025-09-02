import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/service_locator.dart';

import '../../test/utils/fakes.dart';
import '../../test/utils/test_helpers.dart';

void main() {
  WidgetController.hitTestWarningShouldBeFatal = true;

  setUp(() async => await initServiceLocator());
  tearDown(() async => await sl.reset(dispose: true));

  // Call your existing test function
  runLogoutSmokeTests();
}

void runLogoutSmokeTests() {
  group("Logout Smoke Tests", () {
    // Manual TCID: LVPN-7014
    testWidgets(" - normal logout", (tester) async {
      final app = await tester.setupIntegrationTests();

      // accept consent
      final account = fakeAccount();

      final accountScreen = await app.goToAccountScreen(account: account);
      await accountScreen.clickLogOutButton();
      // Need due to this: Bad state: Tried to read a provider from a ProviderContainer that was already disposed
      app.appAccount.delayDuration = Duration(seconds: 1);
    });
  });
}
