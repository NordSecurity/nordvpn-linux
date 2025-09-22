import 'dart:async';
import 'package:flutter/material.dart';
import 'package:clock/clock.dart';

/// Helper class which manages the delayed loading indicator appearance.
///
/// It informs when to show the loading indicator to avoid rapid change
/// of widgets. There are two durations you can control:
/// - [delayDuration] - how long to wait before displaying the loading indicator
/// - [minDisplayDuration] - if we ARE showing loading indicator, then for how long at least
///
/// If long running operation finishes before `delayDuration`, then the loading indicator
/// does not appear.
///
/// Example:
/// ```dart
/// _loadingManager = DelayedLoadingManager(onUpdate: () {
///   if (mounted) setState(() {});
/// });
/// ...
/// _loadingManager.startLoading();
/// request().then((_) {
///   _loadingManager.stopLoading(false);
/// }).catchError((e) {
///   _loadingManager.stopLoading(true);
/// });
/// ```
///
/// ## Cleanup
///
/// Be sure to call [dispose] when done using it.
final class DelayedLoadingManager {
  // Called on state changes.
  final VoidCallback onUpdate;
  // Called when the loading process is finished.
  final VoidCallback onDone;
  // Called when the loading process is finished with error.
  final VoidCallback? onError;
  // If the operation finished before this delay ended, then
  // we are not showing the loading indicator at all.
  final Duration delayDuration;
  // If we ARE showing the loading indicator, then show it for
  // at least this amount of time.
  final Duration minDisplayDuration;
  // Did we started the whole process of delaying loading indicator?
  bool isLoading = false;
  // Are we showing loading indicator?
  bool showLoadingIndicator = false;

  Timer? _loadingDelayTimer;
  Timer? _hideIndicatorTimer;
  DateTime? _loadingIndicatorStartTime;
  bool _isDisposed = false;

  DelayedLoadingManager({
    required this.onUpdate,
    required this.onDone,
    this.onError,
    this.delayDuration = const Duration(milliseconds: 50),
    this.minDisplayDuration = const Duration(milliseconds: 500),
  });

  void startLoading() {
    if (_isDisposed) return;

    isLoading = true;
    onUpdate();

    // 0. Start delay for showing loading indicator.
    _loadingDelayTimer = Timer(delayDuration, () {
      if (_isDisposed) return;
      // if the time has passed, then lets show loading indicator (the operation
      // is taking long enough to show the loading indicator)
      showLoadingIndicator = true;

      // and immediately note time when we started showing the loading indicator
      _loadingIndicatorStartTime = clock.now();
      onUpdate();
    });
  }

  void stopLoading(bool finishWithError) {
    if (_isDisposed) return;

    // 1. Operation is done, so cancel the delay for showing loading indicator.
    _loadingDelayTimer?.cancel();

    // 2. If we are not showing loading indicator yet, then the loading indicator
    // delay didn't finish - so operation was fast enough - don't show loading
    // indicator at all and that's it.
    if (!showLoadingIndicator) {
      _resetLoadingState(finishWithError);
      return;
    }

    // 3. At this point, we are showing loading indicator. Check for how long we are
    // already showing it.
    final elapsed = clock
        .now()
        .difference(_loadingIndicatorStartTime!)
        .inMilliseconds;

    if (elapsed >= minDisplayDuration.inMilliseconds) {
      // 4a. Operation took long - we were showing loading indicator AND we
      // were showing it for at least minimum amount of time to avoid flickering,
      // so now just finish the loading state.
      _resetLoadingState(finishWithError);
      return;
    }

    // 4b. We are not showing loading indicator long enough. Show it for the
    // remaining period.
    final remainingTime = minDisplayDuration.inMilliseconds - elapsed;
    _hideIndicatorTimer = Timer(Duration(milliseconds: remainingTime), () {
      if (_isDisposed) return;
      _resetLoadingState(finishWithError);
    });
  }

  void _resetLoadingState(bool finishWithError) {
    if (_isDisposed) return;
    isLoading = false;
    showLoadingIndicator = false;
    onUpdate();
    if (!finishWithError) {
      onDone();
    } else if (onError != null) {
      onError!();
    }
  }

  void dispose() {
    _isDisposed = true;
    _loadingDelayTimer?.cancel();
    _hideIndicatorTimer?.cancel();
  }
}
