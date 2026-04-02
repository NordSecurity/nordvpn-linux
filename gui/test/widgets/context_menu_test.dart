import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/widgets/context_menu/context_menu.dart';

import '../utils/test_helpers.dart';

void main() {
  Widget buildMenu({
    required List<ContextMenuItem> items,
    double? width,
  }) {
    return ContextMenu(
      items: items,
      width: width,
      anchorBuilder: (toggleMenu) => ElevatedButton(
        onPressed: toggleMenu,
        child: const Text('Open'),
      ),
    );
  }

  group('ContextMenu', () {
    testWidgets('renders anchor widget when closed', (tester) async {
      await tester.setupWidgetTest(
        buildMenu(items: [ContextMenuItem(label: 'Item', onTap: () {})]),
      );

      expect(find.text('Open'), findsOneWidget);
      expect(find.text('Item'), findsNothing);
    });

    testWidgets('opens menu when anchor is tapped', (tester) async {
      await tester.setupWidgetTest(
        buildMenu(items: [ContextMenuItem(label: 'Item', onTap: () {})]),
      );

      await tester.tap(find.text('Open'));
      await tester.pumpAndSettle();

      expect(find.text('Item'), findsOneWidget);
    });

    testWidgets('shows all provided items', (tester) async {
      await tester.setupWidgetTest(
        buildMenu(
          items: [
            ContextMenuItem(label: 'First', onTap: () {}),
            ContextMenuItem(label: 'Second', onTap: () {}),
            ContextMenuItem(label: 'Third', onTap: () {}),
          ],
        ),
      );

      await tester.tap(find.text('Open'));
      await tester.pumpAndSettle();

      expect(find.text('First'), findsOneWidget);
      expect(find.text('Second'), findsOneWidget);
      expect(find.text('Third'), findsOneWidget);
    });

    testWidgets('calls onTap when item is tapped', (tester) async {
      var tapped = false;
      await tester.setupWidgetTest(
        buildMenu(
          items: [
            ContextMenuItem(label: 'Action', onTap: () => tapped = true),
          ],
        ),
      );

      await tester.tap(find.text('Open'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Action'));
      await tester.pumpAndSettle();

      expect(tapped, isTrue);
    });

    testWidgets('closes menu after item is tapped', (tester) async {
      await tester.setupWidgetTest(
        buildMenu(items: [ContextMenuItem(label: 'Action', onTap: () {})]),
      );

      await tester.tap(find.text('Open'));
      await tester.pumpAndSettle();
      expect(find.text('Action'), findsOneWidget);

      await tester.tap(find.text('Action'));
      await tester.pumpAndSettle();

      expect(find.text('Action'), findsNothing);
    });

    testWidgets('closes menu when barrier is tapped', (tester) async {
      await tester.setupWidgetTest(
        buildMenu(items: [ContextMenuItem(label: 'Item', onTap: () {})]),
      );

      await tester.tap(find.text('Open'));
      await tester.pumpAndSettle();
      expect(find.text('Item'), findsOneWidget);

      // Tap a position that is on the barrier (not on the menu item or anchor).
      await tester.tapAt(const Offset(5, 5));
      await tester.pumpAndSettle();

      expect(find.text('Item'), findsNothing);
    });

    testWidgets('barrier tap does not call item onTap', (tester) async {
      var tapped = false;
      await tester.setupWidgetTest(
        buildMenu(
          items: [ContextMenuItem(label: 'Item', onTap: () => tapped = true)],
        ),
      );

      await tester.tap(find.text('Open'));
      await tester.pumpAndSettle();

      await tester.tapAt(const Offset(5, 5));
      await tester.pumpAndSettle();

      expect(tapped, isFalse);
    });

    testWidgets('applies labelColor to item text', (tester) async {
      const testColor = Colors.red;
      await tester.setupWidgetTest(
        buildMenu(
          items: [
            ContextMenuItem(
              label: 'Colored',
              labelColor: testColor,
              onTap: () {},
            ),
          ],
        ),
      );

      await tester.tap(find.text('Open'));
      await tester.pumpAndSettle();

      final textWidget = tester.widget<Text>(find.text('Colored'));
      expect(textWidget.style?.color, testColor);
    });

    testWidgets('item without labelColor uses theme text color', (tester) async {
      await tester.setupWidgetTest(
        buildMenu(
          items: [ContextMenuItem(label: 'Default', onTap: () {})],
        ),
      );

      await tester.tap(find.text('Open'));
      await tester.pumpAndSettle();

      final textWidget = tester.widget<Text>(find.text('Default'));
      // labelColor is null, so the style color comes from the theme's itemTextStyle.
      // It must be non-null since the theme always provides a color.
      expect(textWidget.style?.color, isNotNull);
    });

    testWidgets('closes via barrier when menu is open', (tester) async {
      await tester.setupWidgetTest(
        buildMenu(items: [ContextMenuItem(label: 'Item', onTap: () {})]),
      );

      await tester.tap(find.text('Open'));
      await tester.pumpAndSettle();
      expect(find.text('Item'), findsOneWidget);

      // When the menu is open the barrier covers the screen.
      // Tap outside the menu panel to close via the barrier.
      await tester.tapAt(const Offset(5, 5));
      await tester.pumpAndSettle();
      expect(find.text('Item'), findsNothing);
    });
  });
}
