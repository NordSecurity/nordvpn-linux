import 'dart:async';

import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/logger.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'popup_actions_provider.g.dart';

// Executes the popup actions. The state holds the ids of the popups whose
// actions are currently running.
//
// The popup is closed as soon as the user picks an option, so the action always
// outlives the dialog which triggered it. Actions are therefore invoked with
// this notifier's `ref`, which is owned by the ProviderContainer, and not with
// the `WidgetRef` of the popup. A `WidgetRef` is the widget element itself and
// throws "Cannot use ref after the widget was disposed." on the first
// `ref.read` which happens after an `await`.
//
// The same property is what lets an error reported by the action be displayed
// after the user navigated to another screen: neither the action nor the
// controllers it calls depend on the widget which started them.
//
// Nothing displays a progress indicator for these actions. The reconnect and the
// settings changes are already reflected on the screen the user is on (the
// connection card draws a spinner around the flag while connecting, the protocol
// list greys out, ...), and the screen is deliberately left usable while the
// action runs.
//
// `keepAlive` is mandatory here. With autoDispose this provider would be
// disposed as soon as nothing watched it, which would recreate the very same
// "used after disposal" problem one layer down.
@Riverpod(keepAlive: true)
final class PopupActions extends _$PopupActions {
  @override
  Set<int> build() => const {};

  // Runs [action] for [popupId].
  //
  // Actions for different popups are allowed to overlap: nothing blocks the UI
  // while they run, the daemon serializes the gRPC calls and each action reports
  // its own result through popupsProvider. Only a repeated dispatch of the same
  // popup is dropped, which covers a double click on the confirmation button.
  // Rejecting everything while any action runs would instead silently discard
  // the second change, and for the protocol change the pending value is already
  // consumed by then.
  Future<void> run(PopupAction action, {required int popupId}) async {
    if (state.contains(popupId)) {
      logger.e("popup action for $popupId is already running");
      return;
    }

    state = {...state, popupId};

    try {
      await action(ref);
    } catch (e, stackTrace) {
      logger.e("popup action for $popupId failed: $e $stackTrace");
      // Controllers report daemon status codes themselves, this covers the
      // unexpected errors so that the user always gets a feedback.
      ref.read(popupsProvider.notifier).show(DaemonStatusCode.failure);
    } finally {
      state = {...state}..remove(popupId);
    }
  }
}
