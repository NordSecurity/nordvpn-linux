import 'package:flutter_test/flutter_test.dart';

import '../../test/utils/fakes.dart';
import '../../test/utils/test_helpers.dart';

void runAccountSettingsTests() async {
  group("test account screen", () {
    testWidgets("basic", (tester) async {
      final app = await tester.setupIntegrationTests();
      final account = fakeAccount();

      final accountScreen = await app.goToAccountScreen(account: account);

      // page structure
      expect(accountScreen.hasParentBreadcrumb(), isTrue);
      expect(accountScreen.hasCurrentBreadcrumb(), isTrue);
      expect(accountScreen.hasUserInfo(), isTrue);
      expect(accountScreen.hasProductsList(), isTrue);
      expect(accountScreen.hasFooterLinks(), isTrue);

      // breadcrumbs are set correctly
      expect(accountScreen.parentBreadcrumbLabel(), equals("Settings"));
      expect(accountScreen.currentBreadcrumbLabel(), equals("Account"));

      // user's email is displayed
      expect(accountScreen.userEmail(), equals(account.email));
    });
  });
}
