import 'package:nordvpn/logger.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'toasts_provider.g.dart';

@Riverpod(keepAlive: true)
final class Toasts extends _$Toasts {
  Duration? _timeout;

  @override
  Duration? build() => null;

  void show(Duration t) {
    logger.e("toast_provider::show called!");
    if (_timeout == null) {
      logger.e("showing new toast with timeout: $t");
      _timeout = t;
      state = t;
    }
  }
  void closeToast() {
    logger.e("toast_provider::closeToast called!");
    _timeout = null;
    state = null;
  }
}
