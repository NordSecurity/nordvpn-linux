import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/settings/account_details_screen.dart';
import 'package:nordvpn/settings/autoconnect_settings.dart';
import 'package:nordvpn/settings/navigation.dart';
import 'package:nordvpn/settings/terms_screen.dart';
import 'package:nordvpn/settings/vpn_connection_settings.dart';
import 'package:nordvpn/vpn/servers_list_card.dart';
import 'package:nordvpn/vpn/vpn_status_card.dart';
import 'package:nordvpn/widgets/login_form.dart';

Finder serverInfoText() {
  final serverInfoTextFinder = find.byKey(VpnWidgetKeys.vpnServerInfoText);
  expect(serverInfoTextFinder, findsOneWidget);
  return serverInfoTextFinder;
}

Finder statusLabelText() {
  final statusInfoTextFinder = find.byKey(VpnWidgetKeys.vpnStatusLabelText);
  expect(statusInfoTextFinder, findsOneWidget);
  return statusInfoTextFinder;
}

Finder loginButton() {
  final loginButtonFinder = find.byKey(LoginWidgetKeys.loginButton);
  expect(loginButtonFinder, findsOneWidget);
  return loginButtonFinder;
}

Finder loginForm() {
  final loginFormFinder = find.byKey(LoginWidgetKeys.loginForm);
  expect(loginFormFinder, findsOneWidget);
  return loginFormFinder;
}

Finder loginButtonLoadingIndicator() {
  final loadingIndicatorFinder = find.byKey(LoginWidgetKeys.loadingIndicator);
  expect(loadingIndicatorFinder, findsOneWidget);
  return loadingIndicatorFinder;
}

Finder vpnStatusCard() {
  final vpnStatusFinder = find.byKey(VpnWidgetKeys.vpnStatusCard);
  expect(vpnStatusFinder, findsOneWidget);
  return vpnStatusFinder;
}

Finder quickConnectButton() {
  final quickConnectFinder = find.byKey(VpnWidgetKeys.vpnQuickConnectButton);
  expect(quickConnectFinder, findsOneWidget);
  return quickConnectFinder;
}

Finder subscriptionPopupText() {
  return find.text(t.ui.subscriptionHasEnded);
}

Finder userInfo() {
  final userInfoFinder = find.byKey(AccountWidgetKeys.userInfo);
  expect(userInfoFinder, findsOneWidget);
  return userInfoFinder;
}

Finder parentNavigationBreadcrumb() {
  final parentBreadcrumbFinder = find.byType(NavigableBreadcrumb);
  expect(parentBreadcrumbFinder, findsOneWidget);
  return parentBreadcrumbFinder;
}

Finder currentNavigationBreadcrumb() {
  final currentNavigationBreadcrumb = find.byKey(
    NavWidgetKeys.currentBreadcrumb,
  );
  expect(currentNavigationBreadcrumb, findsWidgets);
  return currentNavigationBreadcrumb.first;
}

Finder productsList() {
  final productsListFinder = find.byKey(AccountWidgetKeys.productsList);
  expect(productsListFinder, findsOneWidget);
  return productsListFinder;
}

Finder virtualServersListItem() {
  final virtualListItemFinder = find.descendant(
    of: find.byKey(ServerListWidgetKeys.countriesServersList),
    matching: find.textContaining(t.ui.virtual),
  );
  return virtualListItemFinder;
}

Finder vpnConnectionBreadcrumb() {
  final currentBreadcrumb = currentNavigationBreadcrumb();
  final vpnConnectonBreadcrumbText = find.descendant(
    of: currentBreadcrumb,
    matching: find.text(t.ui.vpnConnection),
  );
  expect(vpnConnectonBreadcrumbText, findsOneWidget);
  return vpnConnectonBreadcrumbText;
}

Finder autoConnectSwitch() {
  final autoConnectSwitchFinder = find.byKey(
    VpnSettingsWidgetKeys.autoConnectSwitch,
  );
  expect(autoConnectSwitchFinder, findsOneWidget);
  return autoConnectSwitchFinder;
}

Finder killSwitchToggle() {
  final killSwitchFinder = find.byKey(VpnSettingsWidgetKeys.killSwitch);
  expect(killSwitchFinder, findsOneWidget);
  return killSwitchFinder;
}

Finder autoConnectTile() {
  final autoConnectTile = find.byKey(VpnSettingsWidgetKeys.autoConnectTile);
  expect(autoConnectTile, findsOneWidget);
  return autoConnectTile;
}

Finder autoConnectPanel() {
  final autoConnectPanelFinder = find.byKey(
    AutoConnectWidgetKeys.autoConnectPanel,
  );
  expect(autoConnectPanelFinder, findsOneWidget);
  return autoConnectPanelFinder;
}

Finder connectNowButton() {
  final connectButtonFinder = find.byKey(
    AutoConnectWidgetKeys.connectNowButton,
  );
  expect(connectButtonFinder, findsOneWidget);
  return connectButtonFinder;
}

Finder serverTileWithText(String text) {
  return find.descendant(
    of: serversList(),
    matching: find.widgetWithText(ListTile, text),
  );
}

Finder serversList() {
  final serversListFinder = find.byKey(AutoConnectWidgetKeys.serversListCard);
  expect(serversListFinder, findsOneWidget);
  return find.descendant(
    of: serversListFinder,
    matching: find.byType(Scrollable),
  );
}

Finder autoConnectServer() {
  final column = find.descendant(
    of: autoConnectServerInfo(),
    matching: find.byType(Column),
  );
  final serverLabel = find.descendant(of: column, matching: find.byType(Text));
  return serverLabel.first;
}

Finder autoConnectServerInfo() {
  final autoConnectPanelFinder = find.byKey(AutoConnectWidgetKeys.serverInfo);
  expect(autoConnectPanelFinder, findsOneWidget);
  return autoConnectPanelFinder;
}

Finder specialtyServersTab() {
  final specialtyServersLabelFinder = find.byKey(
    ServerListWidgetKeys.specialtyServersTab,
  );
  expect(specialtyServersLabelFinder, findsOne);
  return specialtyServersLabelFinder;
}

Finder doubleVpnGroupTile() {
  final doubleVpnGroupTile = find.byKey(ServerListWidgetKeys.doubleVpn);
  expect(doubleVpnGroupTile, findsOne);
  return doubleVpnGroupTile;
}

Finder onionOverVpnGroupTile() {
  final onionOverVpnGroupTile = find.byKey(ServerListWidgetKeys.onionOverVpn);
  expect(onionOverVpnGroupTile, findsOne);
  return onionOverVpnGroupTile;
}

Finder p2pGroupTile() {
  final p2pGroupTile = find.byKey(ServerListWidgetKeys.p2p);
  expect(p2pGroupTile, findsOne);
  return p2pGroupTile;
}

// Find the gesture area from an OnOffSwitch
Finder tapAreaInOnOffSwitch(Finder onOffSwitch) {
  expect(onOffSwitch, findsOneWidget);
  return find.descendant(
    of: onOffSwitch,
    matching: find.byType(GestureDetector),
  );
}

Finder killSwitchCheckBox() {
  final killSwitchCheckboxFinder = find.byType(Checkbox);
  expect(killSwitchCheckboxFinder, findsOneWidget);
  return killSwitchCheckboxFinder;
}

Finder logoutButtonFinder() {
  final logoutButton = find.byKey(AccountWidgetKeys.logoutButton);
  expect(logoutButton, findsOne);
  return logoutButton;
}

Finder legalDescriptionFinder() {
  return find.byKey(LegalInformationKeys.descriptionKey);
}

Finder legalTermsOfServiceLinkFinder() {
  return find.byKey(LegalInformationKeys.termsOfServiceLinkKey);
}

Finder legalAutoRenewalTermsLinkFinder() {
  return find.byKey(LegalInformationKeys.autoRenewalTermsLinkKey);
}

Finder legalPrivacyPolicyLinkFinder() {
  return find.byKey(LegalInformationKeys.privacyPolicyLinkKey);
}
