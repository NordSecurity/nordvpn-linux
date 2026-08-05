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

      expect(container.read(popupActionsProvider), isNotNull);
      completer.complete();
      await done;

      expect(readAfterAwait, isTrue);
      expect(container.read(popupActionsProvider), isNull);
    });

    test('progress holds the popup id and the message', () async {
      final container = createContainer();
      final completer = Completer<void>();

      final done = container.read(popupActionsProvider.notifier).run(
        (_) => completer.future,
        popupId: popupId,
        progressMessage: "applying",
      );

      final progress = container.read(popupActionsProvider);
      expect(progress?.popupId, popupId);
      expect(progress?.message, "applying");
      // The indicator is displayed only after the delay passed.
      expect(progress?.showIndicator, isFalse);

      completer.complete();
      await done;
    });

    test('failing action displays the generic error popup', () async {
      final container = createContainer();

      await container.read(popupActionsProvider.notifier).run((_) async {
        throw Exception("action failed");
      }, popupId: popupId);

      expect(container.read(popupActionsProvider), isNull);
      expect(container.read(popupsProvider)?.id, DaemonStatusCode.failure);
    });

    test('state is cleared when the action succeeds', () async {
      final container = createContainer();

      await container
          .read(popupActionsProvider.notifier)
          .run((_) async {}, popupId: popupId);

      expect(container.read(popupActionsProvider), isNull);
      // Successful actions report nothing.
      expect(container.read(popupsProvider), isNull);
    });

    test('only one action runs at a time', () async {
      final container = createContainer();
      final completer = Completer<void>();
      var secondActionExecuted = false;

      final first = container
          .read(popupActionsProvider.notifier)
          .run((_) => completer.future, popupId: popupId);

      await container.read(popupActionsProvider.notifier).run((_) async {
        secondActionExecuted = true;
      }, popupId: popupId + 1);

      expect(secondActionExecuted, isFalse);
      expect(container.read(popupActionsProvider)?.popupId, popupId);

      completer.complete();
      await first;
      expect(container.read(popupActionsProvider), isNull);
    });

    test('action receiving no await still reports progress', () async {
      final container = createContainer();
      PopupActionProgress? progressDuringAction;

      void action(Ref ref) {
        progressDuringAction = container.read(popupActionsProvider);
      }

      await container
          .read(popupActionsProvider.notifier)
          .run(action, popupId: popupId);

      expect(progressDuringAction?.popupId, popupId);
      expect(container.read(popupActionsProvider), isNull);
    });
  });
}
