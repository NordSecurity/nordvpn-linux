import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/service_locator.dart';

import 'smoke_test_connect.dart';
import 'smoke_test_disconnect.dart';
import 'smoke_test_login.dart';
import 'smoke_test_logout.dart';

void main() {
  WidgetController.hitTestWarningShouldBeFatal = true;

  setUp(() async => await initServiceLocator());
  tearDown(() async => await sl.reset(dispose: true));

  // Call your existing test function
  runLoginSmokeTests();
  runLogoutSmokeTests();
  runConnectSmokeTests();
  runDisconnectSmokeTests();
}
