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
    return legalDescriptionFinder().evaluate().isNotEmpty;
  }

  bool hasTermsOfServiceLink() {
    return legalTermsOfServiceLinkFinder().evaluate().isNotEmpty;
  }

  bool hasAutoRenewalTermsLink() {
    return legalAutoRenewalTermsLinkFinder().evaluate().isNotEmpty;
  }

  bool hasPrivacyPolicyLink() {
    return legalPrivacyPolicyLinkFinder().evaluate().isNotEmpty;
  }
}
