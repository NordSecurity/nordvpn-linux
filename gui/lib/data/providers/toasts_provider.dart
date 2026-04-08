import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'toasts_provider.g.dart';

@Riverpod(keepAlive: true)
final class Toasts extends _$Toasts {
  Duration? _pendingDuration;

  @override
  Duration? build() => null;

  /// Stores a pause duration to be shown when the daemon
  /// confirms the PAUSED state. Does not show the toast yet.
  void setPendingDuration(Duration duration) {
    _pendingDuration = duration;
  }

  /// Shows the toast using the previously stored pending
  /// duration. No-op if no pending duration was set.
  void showPending() {
    if (_pendingDuration != null) {
      state = _pendingDuration;
      _pendingDuration = null;
    }
  }

  void closeToast() {
    _pendingDuration = null;
    state = null;
  }
}
