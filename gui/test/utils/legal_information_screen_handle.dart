import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/settings/navigation.dart';

import 'finders.dart';
import 'screen_handle.dart';

final class LegalInformationScreenHandle extends ScreenHandle {
  LegalInformationScreenHandle(super.app);

  bool hasParentBreadcrumb() {
    return parentNavigationBreadcrumb().evaluate().isNotEmpty;
  }

  bool hasCurrentBreadcrumb() {
    return currentNavigationBreadcrumb().evaluate().isNotEmpty;
  }

  String parentBreadcrumbLabel() {
    final finder = parentNavigationBreadcrumb();
    final widget = app.tester.widget<NavigableBreadcrumb>(finder);
    return widget.name;
  }

  String currentBreadcrumbLabel() {
    final finder = currentNavigationBreadcrumb();
    final widget = app.tester.widget<Breadcrumb>(finder);
    return widget.name;
  }

  bool hasTermsAgreementDescription() {
    return find.text(t.ui.termsAgreementDescription).evaluate().isNotEmpty;
  }

  bool hasTermsOfServiceLink() {
    return find.text(t.ui.termsOfService).evaluate().isNotEmpty;
  }

  bool hasAutoRenewalTermsLink() {
    return find.text(t.ui.autoRenewalTerms).evaluate().isNotEmpty;
  }

  bool hasPrivacyPolicyLink() {
    return find.text(t.ui.privacyPolicy).evaluate().isNotEmpty;
  }
}