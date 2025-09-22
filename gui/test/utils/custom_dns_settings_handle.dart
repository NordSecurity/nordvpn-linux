import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/settings/custom_dns.dart';
import 'package:nordvpn/widgets/advanced_list_tile.dart';
import 'package:nordvpn/widgets/loading_button.dart';
import 'package:nordvpn/widgets/on_off_switch.dart';

import 'finders.dart';
import 'screen_handle.dart';
import 'test_helpers.dart';

final class CustomDnsSettingsHandle extends ScreenHandle {
  CustomDnsSettingsHandle(super.app);

  Future<void> waitUntilVisible() async {
    final customDnsFinder = find.byType(CustomDns);
    await waitUntilFound(customDnsFinder);
    expect(customDnsFinder, findsOneWidget);
  }

  // Checks if the custom DNS switch is on
  bool isDnsEnabled() {
    final toggle = app.tester.widget<OnOffSwitch>(_onOnToggle());
    return toggle.value;
  }

  // tap on the custom DNS switch
  Future<void> tapOnOffSwitch() async {
    await app.tester.tap(tapAreaInOnOffSwitch(_onOnToggle()));
  }

  // check if the form containing add new DNS server is enabled
  bool isAddDnsFormEnabled() {
    final widget = app.tester.widget<AdvancedListTile>(_addDnsForm());
    return widget.enabled;
  }

  // check if the Add button for adding a new DNS server is enabled
  bool isAddButtonEnabled() {
    final widget = app.tester.widget<TextButton>(addButton());
    return widget.enabled;
  }

  // check if the Add button for adding a new DNS server is visible
  bool isAddButtonVisible() {
    final widget = app.tester.widget<Opacity>(_addButtonOpacity());
    return widget.opacity != 0.0;
  }

  // enter text into the input field
  Future<void> enterDnsAddress(String server) async {
    final textField = addDnsInput();
    await app.tester.enterText(textField, server);
  }

  // check if the input field is empty
  bool isInputFieldEmpty() {
    final textField = app.tester.widget<TextField>(addDnsInput());
    return textField.controller!.text.isEmpty;
  }

  // searches a server into the servers list and taps on its removes button
  Future<void> deleteServer(String server) async {
    final loadingButton = find.byKey(CustomDnsKeys.removeButton(server));
    final button = find.descendant(
      of: loadingButton,
      matching: find.byType(IconButton),
    );
    expect(
      find.descendant(
        of: button,
        matching: app.tester.findSvgWithPath("bin.svg"),
      ),
      findsOne,
    );
    await app.tester.tap(button);
  }

  // check if the dialog that TP is enabled is displayed
  bool isDisableTpPopupDisplayed() {
    final strings = [
      t.ui.threatProtectionWillTurnOffDescription,
      t.ui.threatProtectionWillTurnOffDescription,
      t.ui.setCustomDns,
    ];

    return strings.every((e) => find.text(e).evaluate().length == 1);
  }

  // ---------------------- finders --------------------------

  // find the Add button
  Finder addButton() {
    return find.descendant(
      of: _addDnsForm(),
      matching: find.byType(TextButton),
    );
  }

  // find the input text field
  Finder addDnsInput() {
    return find.descendant(of: _addDnsForm(), matching: find.byType(TextField));
  }

  // find servers listview
  Finder serversList() {
    return find.byKey(CustomDnsKeys.serversList);
  }

  // find a server into the servers listview
  Finder serversItem(String server) {
    return find.descendant(of: serversList(), matching: find.text(server));
  }

  // find the switch to enabled/disable custom DNS
  Finder _onOnToggle() {
    return find.byKey(CustomDnsKeys.onOffSwitch);
  }

  // find the form to used to add new server
  Finder _addDnsForm() {
    return find.byKey(CustomDnsKeys.addDnsForm);
  }

  // find the opacity from the LoadingButton of the Add
  Finder _addButtonOpacity() {
    final loadingButton = find.ancestor(
      of: addButton(),
      matching: find.byType(LoadingButton<TextButton>),
    );
    expect(loadingButton, findsOne);
    return find.descendant(of: loadingButton, matching: find.byType(Opacity));
  }
}
