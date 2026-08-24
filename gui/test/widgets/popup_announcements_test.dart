import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/service_locator.dart';
import 'package:nordvpn/widgets/popups/decision_popup.dart';
import 'package:nordvpn/widgets/popups/info_popup.dart';
import 'package:nordvpn/widgets/popups/rich_popup.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import '../utils/fake_shared_preferences.dart';
import '../utils/finders.dart';
import '../utils/test_helpers.dart';

// Popups are read by the screen reader from the semantics tree:
// - the dialog node carries the popup name (title + message), which is what the
//   screen reader announces when the popup opens;
// - `explicitChildNodes` on that node keeps the name off the buttons, so tabbing
//   between them announces only the focused control.
void main() {
  const title = "Popup title";
  const message = "Popup message";
  const header = "Rich popup header";
  const yesButtonText = "Turn off";
  const noButtonText = "Cancel";
  const actionButtonText = "Renew subscription";

  setUpAll(() async {
    SharedPreferencesAsyncPlatform.instance = FakeSharedPreferencesAsync();
    await initServiceLocator();
  });

  // The test binding reports announce support by default, so only the negative
  // case has to override it.
  Widget withoutAnnounceSupport(Widget child) => Builder(
    builder: (context) => MediaQuery(
      data: MediaQuery.of(context).copyWith(supportsAnnounce: false),
      child: child,
    ),
  );

  Future<void> withSemantics(
    WidgetTester tester,
    Future<void> Function() body,
  ) async {
    final handle = tester.ensureSemantics();
    try {
      await body();
    } finally {
      handle.dispose();
    }
  }

  InfoPopup infoPopup({String text = message}) => InfoPopup(
    metadata: InfoPopupMetadata(id: 1, title: title, message: (_) => text),
  );

  DecisionPopup decisionPopup() => DecisionPopup(
    metadata: DecisionPopupMetadata(
      id: 2,
      title: title,
      message: (_) => message,
      noButtonText: noButtonText,
      yesButtonText: yesButtonText,
      yesAction: (_) {},
    ),
  );

  RichNotificationPopup richPopup({String text = message}) =>
      RichNotificationPopup(
        metadata: RichPopupMetadata(
          id: 3,
          header: header,
          message: (_) => text,
          actionButtonText: actionButtonText,
          image: const SizedBox.shrink(),
          action: (_) {},
        ),
      );

  group('announced when it opens', () {
    testWidgets('info popup announces its title and message', (tester) async {
      await tester.setupWidgetTest(infoPopup());

      final announcements = tester.takeAnnouncements();
      expect(announcements, hasLength(1));
      expect(
        announcements.single.message,
        t.a11y.popupWithContent(title: title, message: message),
      );
    });

    testWidgets('decision popup announces its title and message', (
      tester,
    ) async {
      await tester.setupWidgetTest(decisionPopup());

      expect(
        tester.takeAnnouncements().single.message,
        t.a11y.popupWithContent(title: title, message: message),
      );
    });

    // Rich popups keep their heading in `header` and leave `title` unset, so the
    // announcement must use the header and not the "NordVPN" title fallback.
    testWidgets('rich popup announces its header and message', (tester) async {
      await tester.setupWidgetTest(richPopup());

      final announced = tester.takeAnnouncements().single.message;
      expect(
        announced,
        t.a11y.popupWithContent(title: header, message: message),
      );
      expect(announced, isNot(contains(t.ui.nordVpn)));
    });

    testWidgets('popup with an empty message announces its title alone', (
      tester,
    ) async {
      await tester.setupWidgetTest(infoPopup(text: ""));

      expect(tester.takeAnnouncements().single.message, title);
    });

    testWidgets('popup is announced once, not on every rebuild', (
      tester,
    ) async {
      await tester.setupWidgetTest(infoPopup());
      expect(tester.takeAnnouncements(), hasLength(1));

      await tester.pumpAndSettleWithTimeout();

      expect(tester.takeAnnouncements(), isEmpty);
    });

    testWidgets('nothing is announced when the platform has no support', (
      tester,
    ) async {
      await tester.setupWidgetTest(withoutAnnounceSupport(infoPopup()));

      expect(tester.takeAnnouncements(), isEmpty);
    });
  });

  group('the dialog node stays unnamed', () {
    // A name on the dialog node is re-read on every focus change inside the
    // popup: that is what prepended the popup title and message to each button
    // when tabbing. The popup name belongs to the on-open announcement instead.
    testWidgets('decision popup dialog node carries no name', (tester) async {
      await withSemantics(tester, () async {
        await tester.setupWidgetTest(decisionPopup());

        expect(
          tester.getSemantics(popupSemantics()),
          isSemantics(label: "", scopesRoute: true),
        );
        // the role:dialog node is the one the screen reader names, and a bare
        // Text inside the popup would have its label absorbed into it
        expect(
          tester.getSemantics(find.byType(Dialog)),
          isSemantics(label: ""),
        );
      });
    });

    testWidgets('rich popup dialog node carries no name', (tester) async {
      await withSemantics(tester, () async {
        await tester.setupWidgetTest(richPopup());

        expect(tester.getSemantics(popupSemantics()), isSemantics(label: ""));
        expect(
          tester.getSemantics(find.byType(Dialog)),
          isSemantics(label: ""),
        );
      });
    });
  });

  group('focusable controls announce only themselves', () {
    testWidgets('decision popup buttons carry only their own label', (
      tester,
    ) async {
      await withSemantics(tester, () async {
        await tester.setupWidgetTest(decisionPopup());

        for (final buttonText in [yesButtonText, noButtonText]) {
          final node = tester.getSemantics(find.text(buttonText));
          expect(
            node,
            isSemantics(label: buttonText, isButton: true),
            reason: "$buttonText should be announced as a button",
          );
          expect(
            node.label,
            isNot(contains(message)),
            reason: "$buttonText must not repeat the popup message",
          );
          expect(
            node.label,
            isNot(contains(title)),
            reason: "$buttonText must not repeat the popup title",
          );
        }
      });
    });

    // The reported repro: the reconnect dialog announced each button with the
    // dialog title and message prepended.
    testWidgets('reconnect popup buttons carry only their own label', (
      tester,
    ) async {
      await withSemantics(tester, () async {
        await tester.setupWidgetTest(
          DecisionPopup(
            metadata: DecisionPopupMetadata(
              id: 4,
              title: t.ui.reconnectToChangeProtocol,
              message: (_) => t.ui.reconnectToChangeProtocolDescription,
              noButtonText: t.ui.cancel,
              yesButtonText: t.ui.reconnectNow,
              yesAction: (_) {},
            ),
          ),
        );

        for (final buttonText in [t.ui.cancel, t.ui.reconnectNow]) {
          final node = tester.getSemantics(find.text(buttonText));
          expect(node, isSemantics(label: buttonText, isButton: true));
          expect(
            node.label,
            isNot(contains(t.ui.reconnectToChangeProtocol)),
            reason: "$buttonText must not repeat the dialog title",
          );
          expect(
            node.label,
            isNot(contains(t.ui.reconnectToChangeProtocolDescription)),
            reason: "$buttonText must not repeat the dialog message",
          );
        }
      });
    });

    testWidgets('rich popup action button carries only its own label', (
      tester,
    ) async {
      await withSemantics(tester, () async {
        await tester.setupWidgetTest(richPopup());

        final node = tester.getSemantics(find.text(actionButtonText));
        expect(node, isSemantics(label: actionButtonText, isButton: true));
        expect(node.label, isNot(contains(message)));
        expect(node.label, isNot(contains(header)));
      });
    });

    testWidgets('close button is announced as a button named "Close"', (
      tester,
    ) async {
      await withSemantics(tester, () async {
        await tester.setupWidgetTest(infoPopup());

        expect(
          tester.getSemantics(popupCloseButton()),
          isSemantics(label: t.ui.close, isButton: true),
        );
      });
    });
  });

  group('popup structure', () {
    testWidgets('title is a header node', (tester) async {
      await withSemantics(tester, () async {
        await tester.setupWidgetTest(infoPopup());

        expect(
          tester.getSemantics(popupTitle()),
          isSemantics(label: title, isHeader: true),
        );
      });
    });

    testWidgets('rich popup header is a header node', (tester) async {
      await withSemantics(tester, () async {
        await tester.setupWidgetTest(richPopup());

        expect(
          tester.getSemantics(find.byKey(RichNotificationPopup.headerKey)),
          isSemantics(label: header, isHeader: true),
        );
      });
    });

    testWidgets('message is its own node', (tester) async {
      await withSemantics(tester, () async {
        await tester.setupWidgetTest(infoPopup());

        final node = tester.getSemantics(popupMessage());
        expect(node, isSemantics(label: message));
        expect(
          node.childrenCount,
          0,
          reason:
              "the message must be a leaf node, not a container that "
              "absorbed the label from an ancestor",
        );
      });
    });
  });
}
