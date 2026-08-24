import 'package:flutter/material.dart';
import 'package:nordvpn/widgets/loading_button.dart';

import 'finders.dart';
import 'screen_handle.dart';

final class AutoConnectSettingsScreenHandle extends ScreenHandle {
  AutoConnectSettingsScreenHandle(super.app);

  bool isSecureMyConnectionButtonEnabled() {
    final widget = app.tester.widget<LoadingElevatedButton>(
      secureMyConnectionButton(),
    );
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

  Future<void> secureMyConnection() async {
    await app.tester.tap(secureMyConnectionButton());
    await app.tester.pumpAndSettle();
  }
}
