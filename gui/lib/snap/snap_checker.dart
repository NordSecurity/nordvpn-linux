import 'dart:io';

class SnapChecker {
  /// Permissions required by the app
  static const RequiredPermissions = [
    "network",
    "network-bind",
    "network-control",
    "firewall-control",
    "network-observe",
    "home",
    "login-session-observe",
    "system-observe",
    "hardware-observe",
  ];


  static bool isSnapContext() {
    return Platform.environment.containsKey('SNAP');
  }


  static Future<List<String>?> getMissingPermissions() async {
    if (!isSnapContext()) {
      return null;
    }

    final missing = <String>[];

    for (final iface in RequiredPermissions) {
      try {
        final result = await Process.run('snapctl', ['is-connected', iface]);


        if (result.exitCode != 0) {
          missing.add(iface);
          continue;
        }

      } catch (e) {
        print("Exception checking $iface: $e");
        missing.add(iface);
      }
    }

    return missing;
  }
}