import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/i18n/strings.g.dart';

import 'finders.dart';
import 'screen_handle.dart';

final class LoginScreenHandle extends ScreenHandle {
  LoginScreenHandle(super.tester);

  bool isLoginButtonEnabled() {
    final loginButtonFinder = loginButton();
    expect(loginButtonFinder, findsOneWidget);
    final loginButtonWidgetFinder = find.descendant(
      of: loginButtonFinder,
      matching: find.byType(ElevatedButton),
    );
    expect(loginButtonWidgetFinder, findsOneWidget);
    final widget = app.tester.widget<ElevatedButton>(loginButtonWidgetFinder);
    return widget.enabled;
  }

  Future<void> clickLogin() async {
    await app.tester.tap(loginButton());
    await app.tester.pump();
  }

  Future<void> waitForLoadingIndicator() async {
    await waitUntilFound(
      loginButtonLoadingIndicator(),
      // finished animation == button is enabled again
      finishAnimations: false,
    );
  }

  Future<void> waitForLoginButton() async {
    await waitUntilFound(find.text(t.ui.logIn));
  }

  Future<void> clickTurnOffKillSwitch() async {
    await app.tester.tap(killSwitchCheckBox());
    await app.tester.pump();
  }
}
