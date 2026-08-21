import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/providers/popup_actions_provider.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';

void main() {
  const popupId = 4242;

  ProviderContainer createContainer() {
    final container = ProviderContainer();
    addTearDown(container.dispose);
    return container;
  }

  group('PopupActions', () {
    // The popup is closed before the action starts, so the action must be able
    // to use its ref after an await. This is what used to throw
    // "Cannot use ref after the widget was disposed." when the action was
    // executed with the WidgetRef of the popup.
    test('action can use ref after an await', () async {
      final container = createContainer();
      final completer = Completer<void>();
      var readAfterAwait = false;

      final done = container.read(popupActionsProvider.notifier).run((ref) async {
        await completer.future;
        // Would throw if the ref was scoped to the closed dialog.
        ref.read(popupsProvider.notifier);
        readAfterAwait = true;
      }, popupId: popupId);

      expect(container.read(popupActionsProvider), contains(popupId));
      completer.complete();
      await done;

      expect(readAfterAwait, isTrue);
      expect(container.read(popupActionsProvider), isEmpty);
    });

    test('failing action displays the generic error popup', () async {
      final container = createContainer();

      await container.read(popupActionsProvider.notifier).run((_) async {
        throw Exception("action failed");
      }, popupId: popupId);

      expect(container.read(popupActionsProvider), isEmpty);
      expect(container.read(popupsProvider)?.id, DaemonStatusCode.failure);
    });

    test('state is cleared when the action succeeds', () async {
      final container = createContainer();

      await container
          .read(popupActionsProvider.notifier)
          .run((_) async {}, popupId: popupId);

      expect(container.read(popupActionsProvider), isEmpty);
      // Successful actions report nothing.
      expect(container.read(popupsProvider), isNull);
    });

    // Guards against running the same action twice when the confirmation button
    // is clicked repeatedly.
    test('the same popup id is not started twice', () async {
      final container = createContainer();
      final completer = Completer<void>();
      var secondActionExecuted = false;

      final first = container
          .read(popupActionsProvider.notifier)
          .run((_) => completer.future, popupId: popupId);

      await container.read(popupActionsProvider.notifier).run((_) async {
        secondActionExecuted = true;
      }, popupId: popupId);

      expect(secondActionExecuted, isFalse);
      expect(container.read(popupActionsProvider), contains(popupId));

      completer.complete();
      await first;
      expect(container.read(popupActionsProvider), isEmpty);
    });

    // Nothing blocks the UI while an action runs, so the user can confirm
    // another popup meanwhile. Those actions have to run instead of being
    // silently dropped.
    test('different popup ids run concurrently', () async {
      final container = createContainer();
      final completer = Completer<void>();
      var secondActionExecuted = false;

      final first = container
          .read(popupActionsProvider.notifier)
          .run((_) => completer.future, popupId: popupId);

      await container.read(popupActionsProvider.notifier).run((_) async {
        secondActionExecuted = true;
      }, popupId: popupId + 1);

      expect(secondActionExecuted, isTrue);
      // The first action is still running and was not affected by the second.
      expect(container.read(popupActionsProvider), contains(popupId));
      expect(
        container.read(popupActionsProvider),
        isNot(contains(popupId + 1)),
      );

      completer.complete();
      await first;
      expect(container.read(popupActionsProvider), isEmpty);
    });

    test('synchronous action is tracked while it runs', () async {
      final container = createContainer();
      Set<int>? stateDuringAction;

      void action(Ref ref) {
        stateDuringAction = container.read(popupActionsProvider);
      }

      await container
          .read(popupActionsProvider.notifier)
          .run(action, popupId: popupId);

      expect(stateDuringAction, contains(popupId));
      expect(container.read(popupActionsProvider), isEmpty);
    });
  });
}
