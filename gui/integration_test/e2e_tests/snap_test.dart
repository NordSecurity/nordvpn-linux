import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/mocks/daemon/mock_snap_interceptor.dart';
import 'package:nordvpn/i18n/strings.g.dart';

import '../../test/utils/finders.dart';
import '../../test/utils/test_helpers.dart';

String collectAllSelectableText(Finder parent) {
  final finder = find.descendant(
    of: parent,
    matching: find.byWidgetPredicate((widget) => widget is SelectableText),
  );

  return finder
      .evaluate()
      .map((e) {
        final w = e.widget;
        if (w is SelectableText) {
          return w.data ?? '';
        }
        return '';
      })
      .join(' ');
}

void runSnapErrorScreenTests() async {
  group("test snap error screen", () {
    testWidgets("appears when missing snap interfaces", (tester) async {
      final app = await tester.setupIntegrationTests();
      app.snapInterceptor.setEnabled(true);

      await tester.pumpUntilFound(snapErrorScreenTitle());

      final titleFinder = snapErrorScreenTitle();
      final descriptionFinder = snapErrorScreenDescription();

      expect(titleFinder, findsOne);
      expect(descriptionFinder, findsOne);

      final title = tester.widget<Text>(titleFinder);
      final description = tester.widget<Text>(descriptionFinder);

      expect(title.data, t.ui.snapScreenTitle);
      expect(description.data, t.ui.snapScreenDescription);

      final commands = collectAllSelectableText(snapErrorScreenCopyField());
      for (final item in mockedMissingConnections) {
        expect(commands, contains('sudo snap connect nordvpn:$item'));
      }
    });
  });
}
