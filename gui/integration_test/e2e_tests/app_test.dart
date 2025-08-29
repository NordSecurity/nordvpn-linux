import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/service_locator.dart';

import 'account_settings.dart';
import 'custom_dns_settings.dart';
import 'login_screen.dart';
import 'obfuscated_servers.dart';
import 'vpn_screen.dart';
import 'warmup.dart';
import 'connect_settings.dart';
import 'auto_connect_settings.dart';

void main() async {
  WidgetController.hitTestWarningShouldBeFatal = true;

  setUp(() async => await initServiceLocator());
  tearDown(() async => await sl.reset(dispose: true));

  runWarmupTests();
  runVpnScreenTests();
  runAccountSettingsTests();
  runLoginTests();
  runConnectSettingsTests();
  runAutoConnectSettingsTests();
  runCustomDnsTests();
  runObfuscatedServersTests();
}
