import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'toasts_provider.g.dart';

@Riverpod(keepAlive: true)
final class Toasts extends _$Toasts {
  @override
  Duration? build() => null;

  void show(Duration t) {
    state = t;
  }

  void closeToast() {
    state = null;
  }
}
