import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/settings/account_details_screen.dart';
import 'package:nordvpn/settings/autoconnect_settings.dart';
import 'package:nordvpn/settings/navigation.dart';
import 'package:nordvpn/vpn/servers_list_card.dart';
import 'package:nordvpn/vpn/vpn_status_card.dart';
import 'package:nordvpn/widgets/advanced_list_tile.dart';
import 'package:nordvpn/widgets/login_form.dart';
import 'package:nordvpn/widgets/on_off_switch.dart';

Finder serverInfoText() {
  final serverInfoFinder = find.byType(VpnServerInfo);
  expect(serverInfoFinder, findsOneWidget);
  final serverInfoTextFinder = find.descendant(
    of: serverInfoFinder,
    matching: find.byType(Text),
  );
  expect(serverInfoTextFinder, findsOneWidget);
  return serverInfoTextFinder;
}

Finder statusLabelText() {
  final statusLabelFinder = find.byType(VpnStatusLabel);
  expect(statusLabelFinder, findsOneWidget);
  final statusInfoTextFinder = find.descendant(
    of: statusLabelFinder,
    matching: find.byType(Text),
  );
  expect(statusInfoTextFinder, findsOneWidget);
  return statusInfoTextFinder;
}

Finder loginButton() {
  final loginButtonFinder = find.byType(LoginButton);
  expect(loginButtonFinder, findsOneWidget);
  return loginButtonFinder;
}

Finder loginForm() {
  final loginFormFinder = find.byType(LoginForm);
  expect(loginFormFinder, findsOneWidget);
  return loginFormFinder;
}

Finder loginButtonLoadingIndicator() {
  final loginButtonFinder = loginButton();
  expect(loginButtonFinder, findsOneWidget);
  final loadingIndicatorFinder = find.descendant(
    of: loginButtonFinder,
    matching: find.byType(CircularProgressIndicator),
  );
  expect(loadingIndicatorFinder, findsOneWidget);
  return loadingIndicatorFinder;
}

Finder vpnStatusCard() {
  final vpnStatusFinder = find.byType(VpnStatusCard);
  expect(vpnStatusFinder, findsOneWidget);
  return vpnStatusFinder;
}

Finder quickConnectButton() {
  final quickConnectFinder = find.text(t.ui.quickConnect);
  expect(quickConnectFinder, findsOneWidget);
  return quickConnectFinder;
}

Finder subscriptionPopupText() {
  return find.text(t.ui.subscriptionHasEnded);
}

Finder userInfo() {
  final userInfoFinder = find.byType(UserInfo);
  expect(userInfoFinder, findsOneWidget);
  return userInfoFinder;
}

Finder parentNavigationBreadcrumb() {
  final parentBreadcrumbFinder = find.byType(NavigableBreadcrumb);
  expect(parentBreadcrumbFinder, findsOneWidget);
  return parentBreadcrumbFinder;
}

Finder currentNavigationBreadcrumb() {
  final currentNavigationBreadcrumb = find.byType(Breadcrumb);
  expect(currentNavigationBreadcrumb, findsWidgets);
  return currentNavigationBreadcrumb.first;
}

Finder productsList() {
  final productsListFinder = find.byType(ProductsList);
  expect(productsListFinder, findsOneWidget);
  return productsListFinder;
}

Finder footerLinks() {
  final footerLinksFinder = find.byType(FooterLinks);
  expect(footerLinksFinder, findsOneWidget);
  return footerLinksFinder;
}

Finder virtualServersListItem() {
  final virtualListItemFinder = find.descendant(
    of: find.byKey(ServersListKeys.countriesServersListKey),
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
  final autoConnectSwitchFinder = find.descendant(
    of: _settingsTileContaining(t.ui.autoConnect),
    matching: find.byType(OnOffSwitch),
  );
  expect(autoConnectSwitchFinder, findsOneWidget);
  return autoConnectSwitchFinder;
}

Finder killSwitchToggle() {
  final killSwitchFinder = find.descendant(
    of: _settingsTileContaining(t.ui.killSwitch),
    matching: find.byType(OnOffSwitch),
  );
  expect(killSwitchFinder, findsOneWidget);
  return killSwitchFinder;
}

Finder _settingsTileContaining(String text) {
  return find.widgetWithText(AdvancedListTile, text);
}

Finder autoConnectTile() {
  final autoConnectTile = _settingsTileContaining("${t.ui.autoConnectTo}:");
  expect(autoConnectTile, findsOneWidget);
  return autoConnectTile;
}

Finder autoConnectPanel() {
  final autoConnectPanelFinder = find.byType(AutoconnectPanel);
  expect(autoConnectPanelFinder, findsOneWidget);
  return autoConnectPanelFinder;
}

Finder connectNowButton() {
  final autoConnectPanelFinder = autoConnectPanel();
  final connectButtonFinder = find.descendant(
    of: autoConnectPanelFinder,
    matching: find.widgetWithText(ElevatedButton, t.ui.connectNow),
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
  final serversListFinder = find.byType(ServersListCard);
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
  final autoConnectPanelFinder = find.byType(AutoConnectServerInfo);
  expect(autoConnectPanelFinder, findsOneWidget);
  return autoConnectPanelFinder;
}

Finder specialtyServersTab() {
  final specialtyServersLabelFinder = find.text(t.ui.specialServers);
  expect(specialtyServersLabelFinder, findsOne);
  return specialtyServersLabelFinder;
}

Finder doubleVpnGroupTile() {
  final doubleVpnGroupTile = find.text(t.ui.doubleVpn);
  expect(doubleVpnGroupTile, findsOne);
  return doubleVpnGroupTile;
}

Finder onionOverVpnGroupTile() {
  final onionOverVpnGroupTile = find.text(t.ui.onionOverVpn);
  expect(onionOverVpnGroupTile, findsOne);
  return onionOverVpnGroupTile;
}

Finder p2pGroupTile() {
  final p2pGroupTile = find.text(t.ui.p2p);
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
  final logoutButton = find.text(t.ui.logout);
  expect(logoutButton, findsOne);
  return logoutButton;
}