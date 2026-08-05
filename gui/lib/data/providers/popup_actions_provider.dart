import 'dart:async';

import 'package:clock/clock.dart';
import 'package:flutter/foundation.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/logger.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'popup_actions_provider.g.dart';

// Describes the popup action which is currently running. It is used by
// [PopupActionProgressOverlay] to display the progress indicator.
final class PopupActionProgress {
  // Id of the popup metadata which started the action.
  final int popupId;
  // Optional message displayed together with the progress indicator.
  final String? message;
  // Whether the progress indicator has to be displayed. Fast actions finish
  // before it becomes true, so that the indicator is not flashed for them.
  final bool showIndicator;

  const PopupActionProgress({
    required this.popupId,
    this.message,
    this.showIndicator = false,
  });

  PopupActionProgress copyWith({bool? showIndicator}) => PopupActionProgress(
    popupId: popupId,
    message: message,
    showIndicator: showIndicator ?? this.showIndicator,
  );

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other is PopupActionProgress &&
            popupId == other.popupId &&
            message == other.message &&
            showIndicator == other.showIndicator);
  }

  @override
  int get hashCode => Object.hash(popupId, message, showIndicator);
}

// Executes the popup actions. The state is `null` while nothing is running and
// holds the progress information while an action is in progress.
//
// The popup is closed as soon as the user picks an option, so the action always
// outlives the dialog which triggered it. Actions are therefore invoked with
// this notifier's `ref`, which is owned by the ProviderContainer, and not with
// the `WidgetRef` of the popup. A `WidgetRef` is the widget element itself and
// throws "Cannot use ref after the widget was disposed." on the first
// `ref.read` which happens after an `await`.
//
// `keepAlive` is mandatory here. With autoDispose this provider would be
// disposed as soon as the progress overlay stops watching it, which would
// recreate the very same "used after disposal" problem one layer down.
@Riverpod(keepAlive: true)
final class PopupActions extends _$PopupActions {
  // Actions which finish faster than this don't display the indicator at all.
  @visibleForTesting
  static const indicatorDelay = Duration(milliseconds: 250);
  // Once the indicator is displayed, keep it for at least this long to avoid
  // flickering.
  @visibleForTesting
  static const minIndicatorDuration = Duration(milliseconds: 500);

  @override
  PopupActionProgress? build() => null;

  // Runs [action] and publishes the progress until it is finished.
  // Only one action can run at a time.
  //
  // The state is kept until the progress indicator can be hidden, which is what
  // PopupsListener uses to postpone the next popup. Without it an error reported
  // by the action could be displayed on top of the progress indicator.
  Future<void> run(
    PopupAction action, {
    required int popupId,
    String? progressMessage,
  }) async {
    if (state != null) {
      logger.e("popup action already running, ignoring action for $popupId");
      return;
    }

    state = PopupActionProgress(popupId: popupId, message: progressMessage);

    DateTime? indicatorDisplayedAt;
    final indicatorDelayTimer = Timer(indicatorDelay, () {
      final current = state;
      if (current == null) return;
      indicatorDisplayedAt = clock.now();
      state = current.copyWith(showIndicator: true);
    });

    try {
      await action(ref);
    } catch (e, stackTrace) {
      logger.e("popup action for $popupId failed: $e $stackTrace");
      // Controllers report daemon status codes themselves, this covers the
      // unexpected errors so that the user always gets a feedback.
      ref.read(popupsProvider.notifier).show(DaemonStatusCode.failure);
    } finally {
      indicatorDelayTimer.cancel();
      await _keepIndicatorVisible(indicatorDisplayedAt);
      state = null;
    }
  }

  Future<void> _keepIndicatorVisible(DateTime? displayedAt) async {
    if (displayedAt == null) return;

    final remaining =
        minIndicatorDuration - clock.now().difference(displayedAt);
    if (remaining > Duration.zero) {
      await Future<void>.delayed(remaining);
    }
  }
}
