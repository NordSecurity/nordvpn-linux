import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/internal/urls.dart';
import 'package:url_launcher_platform_interface/url_launcher_platform_interface.dart';

import '../../test/utils/fakes.dart';
import '../../test/utils/mock_url_launcher.dart';
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

      // breadcrumbs are set correctly
      expect(accountScreen.parentBreadcrumbLabel(), equals("Settings"));
      expect(accountScreen.currentBreadcrumbLabel(), equals("Account"));

      // account info elements exist
      expect(accountScreen.hasSubscriptionInfo(), isTrue);
      expect(accountScreen.hasAccountInfo(), isTrue);

      // account created date is displayed
      expect(accountScreen.accountCreatedDate(), isNotEmpty);

      // subscription active until date is displayed
      expect(accountScreen.subscriptionActiveUntil(), isNotEmpty);
      expect(accountScreen.accountEmail(), equals(account.email));
    });

    testWidgets("links have correct text and URIs", (tester) async {
      final app = await tester.setupIntegrationTests();
      final account = fakeAccount();

      final accountScreen = await app.goToAccountScreen(account: account);

      // verify manage subscription link
      expect(
        accountScreen.serviceSubscriptionLink(),
        equals(manageSubscriptionUrl),
        reason: 'Manage subscription link should point to correct URL',
      );

      // verify change password link
      expect(
        accountScreen.accountChangePasswordLink(),
        equals(changePasswordUrl),
        reason: 'Change password link should point to correct URL',
      );
    });

    testWidgets("clicking links launches correct URLs", (tester) async {
      final mockUrlLauncher = MockUrlLauncher();
      UrlLauncherPlatform.instance = mockUrlLauncher;

      final app = await tester.setupIntegrationTests();
      final account = fakeAccount();

      final accountScreen = await app.goToAccountScreen(account: account);

      await accountScreen.clickManageSubscriptionLink();
      await tester.pumpAndSettle();

      expect(mockUrlLauncher.launchedUrls.length, 1);
      expect(
        mockUrlLauncher.launchedUrls.first,
        equals(manageSubscriptionUrl.toString()),
        reason: 'Manage subscription link should launch correct URL',
      );

      // clear for next click
      mockUrlLauncher.launchedUrls.clear();

      await accountScreen.clickChangePasswordLink();
      await tester.pumpAndSettle();

      expect(mockUrlLauncher.launchedUrls.length, 1);
      expect(
        mockUrlLauncher.launchedUrls.first,
        equals(changePasswordUrl.toString()),
        reason: 'Change password link should launch correct URL',
      );
    });
  });
}
