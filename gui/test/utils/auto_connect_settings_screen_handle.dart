import 'package:flutter/material.dart';

import 'finders.dart';
import 'screen_handle.dart';

final class AutoConnectSettingsScreenHandle extends ScreenHandle {
  AutoConnectSettingsScreenHandle(super.app);

  bool isConnectNowButtonEnabled() {
    final widget = app.tester.widget<ElevatedButton>(connectNowButton());
    return widget.onPressed != null;
  }

  Future<void> clickListTile({required String withText}) async {
    await app.tester.tap(serverTileWithText(withText));
    await app.waitForUiUpdates();
  }

  String? autoConnectServerLabel() {
    final serverLabel = app.tester.widget<Text>(autoConnectServer());
    return serverLabel.data;
  }

  Future<void> connectNow() async {
    await app.tester.tap(connectNowButton());
    await app.tester.pumpAndSettle();
  }
}
