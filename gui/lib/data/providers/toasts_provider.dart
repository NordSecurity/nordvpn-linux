import 'package:nordvpn/logger.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'toasts_provider.g.dart';

@Riverpod(keepAlive: true)
final class Toasts extends _$Toasts {
  @override
  Duration? build() => null;

  void show(Duration t) {
    logger.d("toast_provider::show called  with timeout: $t");
    state = t;
  }

  void closeToast() {
    logger.d("toast_provider::closeToast called!");
    state = null;
  }
}
