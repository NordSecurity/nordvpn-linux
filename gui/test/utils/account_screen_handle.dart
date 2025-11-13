import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/settings/account_details_screen.dart';
import 'package:nordvpn/widgets/link_types.dart';
import 'package:nordvpn/settings/navigation.dart';

import 'finders.dart';
import 'screen_handle.dart';

final class AccountScreenHandle extends ScreenHandle {
  AccountScreenHandle(super.app);

  bool hasParentBreadcrumb() {
    return parentNavigationBreadcrumb().evaluate().isNotEmpty;
  }

  bool hasCurrentBreadcrumb() {
    return currentNavigationBreadcrumb().evaluate().isNotEmpty;
  }

  bool hasUserInfo() {
    return accountUserInfo().evaluate().isNotEmpty;
  }

  bool hasProductsList() {
    return productsList().evaluate().isNotEmpty;
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

  bool hasSubscriptionInfo() {
    return serviceSubscriptionInfoFinder().evaluate().isNotEmpty;
  }

  String subscriptionActiveUntil() {
    final widget = app.tester.widget<UserInfoEntry>(
      serviceSubscriptionInfoFinder(),
    );
    return widget.description;
  }

  Uri serviceSubscriptionLink() {
    final widget = app.tester.widget<UserInfoEntry>(
      serviceSubscriptionInfoFinder(),
    );
    return widget.link;
  }

  Future<void> clickManageSubscriptionLink() async {
    final linkFinder = find.descendant(
      of: serviceSubscriptionInfoFinder(),
      matching: find.byType(FirstPartyLink<Uri>),
    );
    await app.tester.tap(linkFinder);
    await app.tester.pump();
  }

  bool hasAccountInfo() {
    return accountInfoFinder().evaluate().isNotEmpty;
  }

  String accountEmail() {
    final widget = app.tester.widget<UserInfoEntry>(accountInfoFinder());
    return widget.title;
  }

  String accountCreatedDate() {
    final widget = app.tester.widget<UserInfoEntry>(accountInfoFinder());
    return widget.description;
  }

  Uri accountChangePasswordLink() {
    final widget = app.tester.widget<UserInfoEntry>(accountInfoFinder());
    return widget.link;
  }

  Future<void> clickChangePasswordLink() async {
    final linkFinder = find.descendant(
      of: accountInfoFinder(),
      matching: find.byType(FirstPartyLink<Uri>),
    );
    await app.tester.tap(linkFinder);
    await app.tester.pump();
  }

  Future<void> clickLogOutButton() async {
    await app.tester.tap(logoutButtonFinder());
    await app.tester.pump();
  }
}
