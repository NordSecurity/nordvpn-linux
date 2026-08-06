import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/data/providers/account_controller.dart';
import 'package:nordvpn/data/providers/popup_actions_provider.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/router/router.dart';
import 'package:nordvpn/widgets/dialog_factory.dart';
import 'package:nordvpn/widgets/popups/decision_popup.dart';
import 'package:nordvpn/widgets/popups/info_popup.dart';
import 'package:nordvpn/widgets/popups/rich_popup.dart';

final class PopupsListener extends ConsumerStatefulWidget {
  final Widget child;

  const PopupsListener({super.key, required this.child});

  @override
  ConsumerState<PopupsListener> createState() => _PopupsListenerState();
}

final class _PopupsListenerState extends ConsumerState<PopupsListener> {
  int? _visiblePopup;

  @override
  Widget build(BuildContext _) {
    ref.listen(accountControllerProvider, (_, next) {
      next.whenData((account) {
        if (account == null) return;
        if (!account.isSubscriptionExpired &&
            _visiblePopup == DaemonStatusCode.accountExpired) {
          closeCurrentPopup();
        }
      });
    });

    // Popups are not withheld while an action is running. An error reported by a
    // long running action has to reach the user as soon as the daemon answers,
    // even if that is after the user navigated to another screen.
    final popupMetadata = ref.watch(popupsProvider);
    if (popupMetadata != null) {
      _showNextPopup(popupMetadata);
    }
    return widget.child;
  }

  void closeCurrentPopup() {
    final ctx = goRouterKey.currentContext;
    if (ctx == null) {
      logger.e("Can't close popup. Context is null.");
      return;
    }
    DialogFactory.close(ctx);
    _visiblePopup = null;
  }

  void _showNextPopup(PopupMetadata metadata) async {
    // This is called from build, so any rebuild while the dialog is displayed
    // would open a second dialog for the same metadata without this guard.
    if (_visiblePopup == metadata.id) return;

    final ctx = goRouterKey.currentContext;
    if (ctx == null) {
      logger.e("Can't display popup. Context is null.");
      return;
    }

    // Notifiers are owned by the ProviderContainer, so they are still usable
    // after the dialog is closed, even if this widget was unmounted meanwhile.
    // `ref` itself wouldn't be.
    final popups = ref.read(popupsProvider.notifier);
    final actions = ref.read(popupActionsProvider.notifier);

    _visiblePopup = metadata.id;

    // Decision popups require explicit user action (yes/no button click)
    // so we disable barrier dismissal to prevent accidental dismissal
    final barrierDismissible = metadata is! DecisionPopupMetadata;

    final choice = await showDialog<DialogResult>(
      context: ctx,
      barrierDismissible: barrierDismissible,
      builder: (_) => buildPopup(metadata),
    );

    _visiblePopup = null;
    // Drain the queue unconditionally. Skipping it when this widget is no longer
    // mounted would keep popupsProvider pointing to this metadata forever and
    // every later popup with the same id would be ignored.
    popups.pop();

    if (metadata is! DecisionPopupMetadata) return;

    final action = switch (choice) {
      DialogResult.yes => metadata.yesAction,
      DialogResult.no => metadata.noAction,
      // Closed with the X icon, the barrier or Esc, no action to run.
      null => null,
    };
    if (action == null) return;

    // Not awaited on purpose: the dialog is closed and the action must not block
    // the popups queue nor the UI while it talks to the daemon.
    unawaited(actions.run(action, popupId: metadata.id));
  }
}

Widget buildPopup(PopupMetadata metadata) {
  switch (metadata) {
    case RichPopupMetadata metadata:
      return RichNotificationPopup(metadata: metadata);
    case DecisionPopupMetadata metadata:
      return DecisionPopup(metadata: metadata);
    case InfoPopupMetadata metadata:
      return InfoPopup(metadata: metadata);
  }
}
