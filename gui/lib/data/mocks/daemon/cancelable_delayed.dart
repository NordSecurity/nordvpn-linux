import 'dart:async';

import 'package:nordvpn/data/mocks/daemon/grpc_server.dart';

// Can be used to run async operations with delay. For each delay a timer is started
// that is stored internally. When dispose is called all the timers are canceled.
// The class register itself to the GrpServer stop and automatically stops all the timers.
class CancelableDelayed {
  final List<Timer> _timers = [];

  CancelableDelayed() {
    GrpcServer().registerOnShutdown(dispose);
  }

  void dispose() {
    for (Timer element in _timers) {
      element.cancel();
    }
    _timers.clear();
  }

  Future<void> delayed(Duration duration) async {
    Completer<void> completer = Completer();
    late Timer timer;
    timer = Timer(duration, () {
      completer.complete(); // Complete the future when the timer fires
      _timers.remove(timer);
    });
    _timers.add(timer);
    return completer.future;
  }
}
