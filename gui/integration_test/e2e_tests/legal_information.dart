import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';

import '../../test/utils/test_helpers.dart';

void runLegalInformationTests() async {
  group("test legal information screen", () {
    testWidgets("displays all required content", (tester) async {
      final app = await tester.setupIntegrationTests();

      final legalScreen = await app.goToLegalInformationScreen();

      // page structure
      expect(legalScreen.hasParentBreadcrumb(), isTrue);
      expect(legalScreen.hasCurrentBreadcrumb(), isTrue);

      // breadcrumbs are set correctly
      expect(legalScreen.parentBreadcrumbLabel(), equals(t.ui.settings));
      expect(legalScreen.currentBreadcrumbLabel(), equals(t.ui.terms));

      // content is displayed
      expect(legalScreen.hasTermsAgreementDescription(), isTrue);
      expect(legalScreen.hasTermsOfServiceLink(), isTrue);
      expect(legalScreen.hasAutoRenewalTermsLink(), isTrue);
      expect(legalScreen.hasPrivacyPolicyLink(), isTrue);
    });
  });
}
