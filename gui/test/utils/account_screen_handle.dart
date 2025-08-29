import 'package:nordvpn/settings/account_details_screen.dart';
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
    return userInfo().evaluate().isNotEmpty;
  }

  bool hasProductsList() {
    return productsList().evaluate().isNotEmpty;
  }

  bool hasFooterLinks() {
    return footerLinks().evaluate().isNotEmpty;
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

  String userEmail() {
    final finder = userInfo();
    final widget = app.tester.widget<UserInfo>(finder);
    return widget.userAccount!.email;
  }

  Future<void> clickLogOutButton() async {
    await app.tester.tap(logoutButtonFinder());
    await app.tester.pump();
  }
}
