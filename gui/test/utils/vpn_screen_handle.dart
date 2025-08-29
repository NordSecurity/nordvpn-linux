import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/vpn/servers_list_card.dart';

import 'finders.dart';
import 'screen_handle.dart';

final class VpnScreenHandle extends ScreenHandle {
  VpnScreenHandle(super.app);

  String? findServerInfoText() {
    final widget = app.tester.widget<Text>(serverInfoText());
    return widget.data;
  }

  String? findStatusLabelText() {
    final widget = app.tester.widget<Text>(statusLabelText());
    return widget.data;
  }

  Future<void> quickConnect() async {
    await app.tester.tap(quickConnectButton());
    await app.tester.pump();
  }

  Future<void> connectToCountry(String country) async {
    await app.tester.tap(find.text(country));
    await app.tester.pump();
  }

  Future<void> disconnect() async {
    await app.tester.tap(disconnectButton());
    await app.tester.pump();
  }

  bool isSubscriptionPopupVisible() {
    return subscriptionPopupText().evaluate().isNotEmpty;
  }

  Future<bool> serversListHasVirtualServers() async {
    try {
      await app.tester.scrollUntilVisible(
        virtualServersListItem(),
        100.0,
        scrollable: find.descendant(
          of: find.byKey(ServersListKeys.countriesServersListKey),
          matching: find.byType(Scrollable),
        ),
      );
    } catch (error) {
      if (error is StateError) {
        if (error.message.contains("No element")) {
          return false;
        }
        return error.message.contains("Too many elements");
      }
    }
    return virtualServersListItem().evaluate().isNotEmpty;
  }

  Future<void> clickSpecialtyServersTab() async {
    await app.tester.tap(specialtyServersTab());
    await app.tester.pumpAndSettle();
  }

  Future<void> clickDoubleVpnGroup() async {
    await app.tester.tap(doubleVpnGroupTile());
    await app.tester.pumpAndSettle();
  }

  Future<void> clickOnionOverVpn() async {
    await app.tester.tap(onionOverVpnGroupTile());
    await app.tester.pumpAndSettle();
  }

  Future<void> clickP2p() async {
    await app.tester.tap(p2pGroupTile());
    await app.tester.pumpAndSettle();
  }

  Future<void> clickSearch() async {
    await app.tester.tap(_searchButton());
    await app.tester.pumpAndSettle();
  }

  bool isObfuscationWarningDisplayed() {
    final finder = find.text(t.ui.obfuscationSearchWarning);
    return finder.evaluate().length == 1;
  }

  Future<void> searchServer(String text) async {
    expect(_serversSearchTextField(), findsOne);
    await app.tester.enterText(_serversSearchTextField(), text);
    await app.tester.pumpAndSettle();
  }

  Future<bool> isObfuscationNoResultsFound() async {
    final msgFinder = find.text(t.ui.obfuscationErrorNoServerFound);
    final goToSettingsLabel = find.descendant(
      of: _goToSettings(),
      matching: find.text(t.ui.goToSettings),
    );

    return msgFinder.evaluate().isNotEmpty &&
        _goToSettings().evaluate().isNotEmpty &&
        goToSettingsLabel.evaluate().isNotEmpty;
  }

  // -------------- Finders -------
  Finder _searchButton() {
    final finder = find.byKey(ServersListKeys.searchKey);
    expect(finder, findsOne);
    return finder;
  }

  Finder _serversSearchTextField() {
    final finder = find.descendant(
      of: find.byType(ServersListCard),
      matching: find.byType(TextField),
    );
    expect(finder, findsOne);
    return finder;
  }

  Finder _goToSettings() {
    final goToSettingsFinder = find.descendant(
      of: find.byType(ServersListCard),
      matching: find.byType(TextButton),
    );
    return goToSettingsFinder;
  }

  Finder disconnectButton() {
    final disconnectFinder = find.text(t.ui.disconnect);
    expect(disconnectFinder, findsOneWidget);
    return disconnectFinder;
  }
}
